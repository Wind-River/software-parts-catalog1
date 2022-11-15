package generics

import (
	"sort"
	"testing"
)

func TestHeapPushPop(t *testing.T) {
	type fields[E any] struct {
		heap []E
		less func(E, E) bool
	}
	type args[E any] struct {
		input []E
	}
	intTests := []struct {
		name   string
		fields fields[int]
		args   args[int]
	}{
		{
			name: "1 2 3",
			fields: fields[int]{
				heap: []int{},
				less: func(a, b int) bool {
					return a < b
				},
			},
			args: args[int]{
				input: []int{1, 2, 3},
			},
		},
	}
	for _, tt := range intTests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Heap[int]{
				internalHeap: internalHeap[int]{
					heap: tt.fields.heap,
					less: tt.fields.less,
				},
			}

			want := make([]int, len(tt.args.input))
			copy(want, tt.args.input)
			sort.Ints(want)

			for _, v := range tt.args.input {
				h.Push(v)
			}

			for i, w := range want {
				g := h.Pop()

				if w != g {
					t.Errorf("Heap.Pop(%d) = %v, want %v", i, g, w)
					return
				}
			}
		})
	}
}
