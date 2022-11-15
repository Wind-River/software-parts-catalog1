package generics

type PriorityItem[E any] struct {
	item     E
	priority int
}

type PriorityQueue[E any] struct {
	*Heap[PriorityItem[E]]
	prioritize func(E) int
}

func (mpq *PriorityQueue[E]) Push(e E) {
	mpq.Heap.Push(PriorityItem[E]{
		item:     e,
		priority: mpq.prioritize(e),
	})
}

func (mpq *PriorityQueue[E]) Pop() E {
	return mpq.Heap.Pop().item
}

func NewMinPriorityQueue[E any](prioritize func(E) int) *PriorityQueue[E] {
	less := func(a PriorityItem[E], b PriorityItem[E]) bool {
		return a.priority < b.priority
	}

	return &PriorityQueue[E]{
		Heap:       NewHeap(less),
		prioritize: prioritize,
	}
}

func NewMaxPriorityQueue[E any](prioritize func(E) int) *PriorityQueue[E] {
	less := func(a PriorityItem[E], b PriorityItem[E]) bool {
		return a.priority > b.priority
	}

	return &PriorityQueue[E]{
		Heap:       NewHeap(less),
		prioritize: prioritize,
	}
}
