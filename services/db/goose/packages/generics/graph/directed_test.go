package graph

import (
	"testing"
)

func binaryTreeLeft(node int64) int64 {
	return node * 2
}
func binaryTreeRight(node int64) int64 {
	return node*2 + 1
}
func TestAddEdgeBinaryTree(t *testing.T) {
	t.Run("AddEdgeBinaryTree", func(t *testing.T) {
		graph := NewDirectedGraph[int64, int64]()
		rootNode := graph.Insert(1, 1)
		left := graph.Insert(binaryTreeLeft(1), binaryTreeLeft(1))
		right := graph.Insert(binaryTreeRight(1), binaryTreeRight(1))
		rootNode.Edges.Add(left)
		rootNode.Edges.Add(right)

		var sum int64 = 0
		if err := graph.TraverseUniqueEdges(func(id int64) error {
			sum += id
			if id != 1 && id <= 10 {
				current := graph.Get(id)
				left := graph.Insert(binaryTreeLeft(id), binaryTreeLeft(id))
				right := graph.Insert(binaryTreeRight(id), binaryTreeRight(id))
				current.Edges.Add(left)
				current.Edges.Add(right)
			}
			return nil
		},
			1,
		); err != nil {
			t.Error(err)
			t.FailNow()
		}

		if sum != (1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 + 10 + 11 + 12 + 13 + 14 + 15 + 16 + 17 + 18 + 19 + 20 + 21) {
			t.Errorf("sum expected: 231; got: %d", sum)
			t.FailNow()
		}
	})
}
