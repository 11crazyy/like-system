package cache

import "sync"

// minHeapImpl 实现container/heap接口的最小堆
type minHeapImpl []*Node

func (mh minHeapImpl) Len() int { return len(mh) }

func (mh minHeapImpl) Less(i, j int) bool {
	return mh[i].count < mh[j].count
}

func (mh minHeapImpl) Swap(i, j int) {
	mh[i], mh[j] = mh[j], mh[i]
	mh[i].index = i
	mh[j].index = j
}

func (mh *minHeapImpl) Push(x interface{}) {
	n := len(*mh)
	node := x.(*Node)
	node.index = n
	*mh = append(*mh, node)
}

func (mh *minHeapImpl) Pop() interface{} {
	old := *mh
	n := len(old)
	node := old[n-1]
	node.index = -1 // 标记为已移除
	*mh = old[0 : n-1]
	return node
}

// 为minHeapImpl添加锁保护
type lockedMinHeap struct {
	sync.Mutex
	heap *minHeapImpl
}
