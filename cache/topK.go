package cache

type TopK interface {
	Add(string, int) AddResult
	List() []Item
	Expelled() <-chan Item
	Fading()
	Total() int64
}
