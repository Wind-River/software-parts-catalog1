package generics

import "container/heap"

type internalHeap[E any] struct { // implement heap.Interface
	heap []E
	less func(E, E) bool
}

func (pq internalHeap[E]) Len() int {
	return len(pq.heap)
}

func (pq internalHeap[E]) Less(i, j int) bool {
	return pq.less(pq.heap[i], pq.heap[j])
}

func (pq *internalHeap[E]) Swap(i, j int) {
	pq.heap[i], pq.heap[j] = pq.heap[j], pq.heap[i]
}

func (pq *internalHeap[E]) Push(e any) {
	pq.heap = append(pq.heap, e.(E))
}

func (pq *internalHeap[E]) Pop() any {
	e := pq.heap[pq.Len()-1]
	pq.heap = pq.heap[0 : pq.Len()-1]

	return e
}

type Heap[E any] struct {
	internalHeap[E]
}

func NewHeap[E any](less func(E, E) bool) *Heap[E] {
	return &Heap[E]{
		internalHeap: internalHeap[E]{
			heap: make([]E, 0),
			less: less,
		},
	}
}

func (h *Heap[E]) Push(e E) {
	heap.Push(&h.internalHeap, e)
}

func (h *Heap[E]) Pop() E {
	return heap.Pop(&h.internalHeap).(E)
}
