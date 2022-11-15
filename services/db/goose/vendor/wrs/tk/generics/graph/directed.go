package graph

import (
	"fmt"
	"path/filepath"
	"sort"
	"wrs/tk/generics"

	"github.com/pkg/errors"
)

// Construct a directed graph of packages and sub-packages, to detect quines

type orderable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~uintptr | ~float32 | ~float64 | ~string
}

type DirectedNode[E any, I orderable] struct {
	ID    I
	Value E
	Edges DirectedNodeList[E, I]
}

func dfsNodeCycleSearch[E any, I orderable](visited map[I]bool, root *DirectedNode[E, I], current *DirectedNode[E, I]) bool {
	if root == current {
		return true
	}

	if current == nil {
		current = root
	}

	for _, edge := range current.Edges {
		if visited[edge.ID] {
			continue // skip node already visited
		}

		visited[edge.ID] = true
		if dfsNodeCycleSearch(visited, root, edge) {
			return true
		}
	}

	return false
}

func (node *DirectedNode[E, I]) IsInCycle() bool {
	return dfsNodeCycleSearch(make(map[I]bool), node, nil)
}

type DirectedNodeList[E any, I orderable] []*DirectedNode[E, I] // implement sort.Interface

func (list DirectedNodeList[E, I]) Len() int {
	return len(list)
}

func (list DirectedNodeList[E, I]) Less(i, j int) bool {
	return list[i].ID < list[j].ID
}

func (list DirectedNodeList[E, I]) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list DirectedNodeList[E, I]) Search(id I) *DirectedNode[E, I] {
	index := sort.Search(list.Len(), func(i int) bool {
		return list[i].ID == id
	})

	if index < list.Len() {
		return list[index]
	}

	return nil
}

func (list *DirectedNodeList[E, I]) Add(new *DirectedNode[E, I]) {
	if list.Search(new.ID) != nil {
		return // node already in list
	}

	var slice []*DirectedNode[E, I] = *list
	slice = append(slice, new)

	*list = slice
	sort.Sort(list)
}

type DirectedGraph[E any, I orderable] map[I]*DirectedNode[E, I]

func NewDirectedGraph[E any, I orderable]() *DirectedGraph[E, I] {
	var graph DirectedGraph[E, I] = make(map[I]*DirectedNode[E, I])
	return &graph
}

func (g DirectedGraph[E, I]) Length() int {
	return len(g)
}

func (g *DirectedGraph[E, I]) Insert(element E, id I, edges ...*DirectedNode[E, I]) *DirectedNode[E, I] {
	if node := g.Get(id); node != nil {
		return node
	}

	newNode := &DirectedNode[E, I]{
		ID:    id,
		Value: element,
		Edges: make([]*DirectedNode[E, I], 0, len(edges)),
	}
	for _, subNode := range edges {
		if subNode == nil {
			continue
		}

		newNode.Edges.Add(subNode)
	}

	var m map[I]*DirectedNode[E, I] = *g
	m[id] = newNode

	return m[id]
}

func (g DirectedGraph[E, I]) Get(id I) *DirectedNode[E, I] {
	var m map[I]*DirectedNode[E, I] = g
	if node, ok := m[id]; ok {
		return node
	}

	return nil
}

type DirectedEdge[I orderable] struct {
	History string
	FromID  I
	ToID    I
}

func (g DirectedGraph[E, I]) TraverseUniqueEdges(visitor func(element E) error, rootIDs ...I) error {
	if len(rootIDs) == 0 {
		return errors.New("any root id is requried")
	}

	seenEdges := make(map[DirectedEdge[I]]bool)
	edgeQueue := generics.NewQueue[DirectedEdge[I]]()

	// visit roots
	for _, rootID := range rootIDs {
		rootNode := g.Get(rootID)

		if err := visitor(rootNode.Value); err != nil {
			return err
		}

		for _, edge := range rootNode.Edges {
			if edge.ID == rootNode.ID { // short-circuit cycle to root
				continue
			}

			edgeQueue.Push(DirectedEdge[I]{
				History: "/",
				FromID:  rootNode.ID,
				ToID:    edge.ID,
			})
		}
	}

	for edgeQueue.Length() > 0 {
		edge := edgeQueue.Pop()

		currentNode := g.Get(edge.ToID)
		if currentNode == nil { // Should not be possible
			return errors.New(fmt.Sprintf("%#v ToID %v does not exist", edge, edge.ToID))
		}

		// visit node
		if err := visitor(currentNode.Value); err != nil {
			return err
		}

		for _, subNode := range currentNode.Edges {
			subEdge := DirectedEdge[I]{
				History: filepath.Join(edge.History, fmt.Sprintf("%v", currentNode.ID)),
				FromID:  currentNode.ID,
				ToID:    subNode.ID,
			}

			if seenEdges[subEdge] { // skip adding known edge
				continue
			}
			seenEdges[subEdge] = true
			edgeQueue.Push(subEdge)
		}
	}

	return nil
}
