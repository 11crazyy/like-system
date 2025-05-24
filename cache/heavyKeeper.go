package cache

import (
	"container/heap"
	"math"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
)

// 一个键值对
type Item struct {
	key   string
	count int
}

// 添加操作的结果
type AddResult struct {
	ExpelledKey string //删除的key
	IsHotKey    bool   //是否是热门key
	CurrentKey  string //当前key
}

// Node用于最小堆中的节点
type Node struct {
	key   string
	count int
	index int //堆中的索引
}

type HeavyKeeper struct {
	k           int //topK数量
	buckets     [][]*Bucket
	width       int            //哈希表宽度
	depth       int            //哈希表深度
	minCount    int            //最小计数阈值
	lookupTable []float64      //衰减查找表
	minHeap     *lockedMinHeap //带锁的最小堆
	expelledCh  chan Item
	total       int64
	rnd         *rand.Rand
}

// HeavyKeeper中的桶
type Bucket struct {
	count uint64
	sync.Mutex
	fingerprint uint64
}

const lookupTableSize = 256

// NewHeavyKeeper 创建新的HeavyKeeper实例
func NewHeavyKeeper(k, width, depth, minCount int, decay float64) *HeavyKeeper {
	hk := &HeavyKeeper{
		k:          k,
		width:      width,
		depth:      depth,
		minCount:   minCount,
		expelledCh: make(chan Item, 1000), // 缓冲通道
		rnd:        rand.New(rand.NewSource(rand.Int63())),
	}

	// 初始化衰减查找表
	hk.lookupTable = make([]float64, lookupTableSize)
	for i := 0; i < lookupTableSize; i++ {
		hk.lookupTable[i] = math.Pow(decay, float64(i))
	}

	// 初始化桶
	hk.buckets = make([][]*Bucket, depth)
	for i := 0; i < depth; i++ {
		hk.buckets[i] = make([]*Bucket, width)
		for j := 0; j < width; j++ {
			hk.buckets[i][j] = &Bucket{}
		}
	}

	// 初始化最小堆
	mh := make(minHeapImpl, 0, k)
	heap.Init(&mh)

	hk = &HeavyKeeper{
		minHeap: &lockedMinHeap{
			heap: &mh,
		},
	}
	return hk
}

// Add 添加一个key并增加其计数
func (hk *HeavyKeeper) Add(key string, increment int) AddResult {
	keyBytes := []byte(key)
	fingerprint := murmurHash(keyBytes)
	maxCount := 0

	// 1. 更新所有哈希桶
	for i := 0; i < hk.depth; i++ {
		bucketIdx := int(murmurHashWithSeed(keyBytes, uint32(i))) % hk.width
		bucket := hk.buckets[i][bucketIdx]

		bucket.Lock()
		if bucket.count == 0 {
			// 空桶，直接占用
			bucket.fingerprint = fingerprint
			bucket.count = uint64(increment)
			maxCount = max(maxCount, increment)
		} else if bucket.fingerprint == fingerprint {
			// 相同key，增加计数
			bucket.count += uint64(increment)
			maxCount = max(maxCount, int(bucket.count))
		} else {
			// 不同key，应用衰减策略
			for j := 0; j < increment; j++ {
				decayIdx := min(bucket.count, lookupTableSize-1)
				decay := hk.lookupTable[decayIdx]
				if hk.rnd.Float64() < decay {
					bucket.count--
					if bucket.count == 0 {
						bucket.fingerprint = fingerprint
						bucket.count = uint64(increment - j)
						maxCount = max(maxCount, int(bucket.count))
						break
					}
				}
			}
		}
		bucket.Unlock()
	}

	// 2. 更新总计数
	atomic.AddInt64(&hk.total, int64(increment))

	// 3. 如果计数不足，直接返回
	if maxCount < hk.minCount {
		return AddResult{}
	}

	// 4. 更新TopK堆
	var expelledKey string
	var isHot bool

	hk.minHeap.Lock()
	defer hk.minHeap.Unlock()

	// 查找是否已在堆中
	for i, node := range *hk.minHeap.heap {
		if node.key == key {
			// 已存在，更新计数
			(*hk.minHeap.heap)[i].count = maxCount
			heap.Fix(hk.minHeap.heap, i)
			isHot = true
			break
		}
	}

	if !isHot {
		// 新key，检查是否能进入TopK
		if hk.minHeap.heap.Len() < hk.k || maxCount >= (*hk.minHeap.heap)[0].count {
			newNode := &Node{key: key, count: maxCount}
			if hk.minHeap.heap.Len() >= hk.k {
				// 堆已满，移除最小元素
				expelled := heap.Pop(hk.minHeap.heap).(*Node)
				expelledKey = expelled.key
				select {
				case hk.expelledCh <- Item{key: expelled.key, count: expelled.count}:
				default: // 如果通道满，丢弃
				}
			}
			heap.Push(hk.minHeap.heap, newNode)
			isHot = true
		}
	}

	return AddResult{
		ExpelledKey: expelledKey,
		IsHotKey:    isHot,
		CurrentKey:  key,
	}
}

// List 获取TopK列表
func (hk *HeavyKeeper) List() []Item {
	hk.minHeap.Lock()
	defer hk.minHeap.Unlock()

	items := make([]Item, hk.minHeap.heap.Len())
	for i, node := range *hk.minHeap.heap {
		items[i] = Item{key: node.key, count: node.count}
	}

	// 按计数降序排序
	sort.Slice(items, func(i, j int) bool {
		return items[i].count > items[j].count
	})
	return items
}

// Expelled 获取被挤出的key通道
func (hk *HeavyKeeper) Expelled() <-chan Item {
	return hk.expelledCh
}

// Fading 衰减所有计数
func (hk *HeavyKeeper) Fading() {
	// 衰减桶计数
	for i := 0; i < hk.depth; i++ {
		for j := 0; j < hk.width; j++ {
			hk.buckets[i][j].Lock()
			hk.buckets[i][j].count >>= 1
			hk.buckets[i][j].Unlock()
		}
	}

	// 衰减堆中计数
	hk.minHeap.Lock()
	for i := range *hk.minHeap.heap {
		(*hk.minHeap.heap)[i].count >>= 1
	}
	heap.Init(hk.minHeap.heap) // 重新堆化
	hk.minHeap.Unlock()

	// 衰减总计数
	atomic.StoreInt64(&hk.total, atomic.LoadInt64(&hk.total)>>1)
}

// Total 获取总计数
func (hk *HeavyKeeper) Total() int64 {
	return atomic.LoadInt64(&hk.total)
}

// murmurHash 实现MurmurHash3算法
func murmurHash(data []byte) uint64 {
	// 这里简化实现，实际应使用完整MurmurHash3
	var h uint64 = 0xdeadbeef
	const prime uint64 = 0x9e3779b185ebcaa7
	for _, b := range data {
		h ^= uint64(b)
		h *= prime
		h ^= h >> 33
	}
	return h
}

func murmurHashWithSeed(data []byte, seed uint32) uint32 {
	// 带种子的MurmurHash实现
	// 简化版，实际应使用完整实现
	var h uint32 = seed ^ uint32(len(data))
	const prime uint32 = 0x5bd1e995
	for _, b := range data {
		h ^= uint32(b)
		h *= prime
		h ^= h >> 15
	}
	return h
}
