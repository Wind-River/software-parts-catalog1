package generics

type Queue[E any] []E

func (q *Queue[E]) Push(element E) {
	*q = append(*q, element)
}

func (q *Queue[E]) Pop() E {
	ret := (*q)[0]

	*q = (*q)[1:]

	return ret
}

func (q Queue[E]) Length() int {
	return len(q)
}

func NewQueue[E any]() Queue[E] {
	return make([]E, 0)
}
