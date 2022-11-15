package generics

import (
	"sort"
)

// Construct a directed graph of packages and sub-packages, to detect quines

type orderable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~uintptr | ~float32 | ~float64 | ~string
}

type DAGNode[E any, I orderable] struct {
	ID       I
	Value    E
	SubNodes DAGNodeList[E, I]
}

func dfsNodeCycleSearch[E any, I orderable](visited map[I]bool, root *DAGNode[E, I], current *DAGNode[E, I]) bool {
	if root == current {
		return true
	}

	if current == nil {
		current = root
	}

	for _, subNode := range current.SubNodes {
		if visited[subNode.ID] {
			continue // skip node already visited
		}

		visited[subNode.ID] = true
		if dfsNodeCycleSearch(visited, root, subNode) {
			return true
		}
	}

	return false
}

func (node *DAGNode[E, I]) IsInCycle() bool {
	return dfsNodeCycleSearch(make(map[I]bool), node, nil)
}

type DAGNodeList[E any, I orderable] []*DAGNode[E, I] // implement sort.Interface

func (list DAGNodeList[E, I]) Len() int {
	return len(list)
}

func (list DAGNodeList[E, I]) Less(i, j int) bool {
	return list[i].ID < list[j].ID
}

func (list DAGNodeList[E, I]) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list DAGNodeList[E, I]) Search(id I) *DAGNode[E, I] {
	index := sort.Search(list.Len(), func(i int) bool {
		return list[i].ID == id
	})

	if index < list.Len() {
		return list[index]
	}

	return nil
}

func (list *DAGNodeList[E, I]) Add(new *DAGNode[E, I]) {
	if list.Search(new.ID) != nil {
		return // node already in list
	}

	var slice []*DAGNode[E, I] = *list
	slice = append(slice, new)

	*list = slice
	sort.Sort(list)
}

type NodeGraph[E any, I orderable] map[I]*DAGNode[E, I]

func NewDAG[E any, I orderable]() *NodeGraph[E, I] {
	var graph NodeGraph[E, I] = make(map[I]*DAGNode[E, I])
	return &graph
}

func (g *NodeGraph[E, I]) Insert(element E, id I, subNodes ...*DAGNode[E, I]) *DAGNode[E, I] {
	var m map[I]*DAGNode[E, I] = *g
	if node, ok := m[id]; ok {
		return node
	}

	newNode := &DAGNode[E, I]{
		ID:       id,
		Value:    element,
		SubNodes: make([]*DAGNode[E, I], 0, len(subNodes)),
	}
	for _, subNode := range subNodes {
		if subNode == nil {
			continue
		}

		newNode.SubNodes.Add(subNode)
	}

	m[id] = newNode

	return m[id]
}
