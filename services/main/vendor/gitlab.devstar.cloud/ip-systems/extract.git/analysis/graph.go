package analysis

// TODO abstact infinite loop protection

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/pkg/errors"
)

// Construct a directed graph of packages and sub-packages, to detect quines

type MD5 [16]byte
type SHA1 [20]byte
type SHA256 [32]byte

type HASH SHA256
type PackageChecksum [32 + 4]byte // sha256 + int

type PackageNode struct {
	Name string
	Path string
	Size int
	Sha1 SHA1
	Hash HASH

	SubPackages PackageNodeList
}

func CalculatePackageChecksum(hash HASH, size int) PackageChecksum {
	var ret [32 + 4]byte
	for i, b := range hash {
		ret[i] = b
	}

	intSlice := make([]byte, 4)
	binary.LittleEndian.PutUint32(intSlice, uint32(size))
	for i, b := range intSlice {
		ret[32+i] = b
	}

	return ret
}

func (node PackageNode) Checksum() PackageChecksum {
	return CalculatePackageChecksum(node.Hash, node.Size)
}

func dfsNodeCycleSearch(visited map[PackageChecksum]bool, root *PackageNode, current *PackageNode) bool {
	if root == current {
		return true
	}

	if current == nil {
		current = root
	}

	for _, subPackage := range current.SubPackages {
		checksum := subPackage.Checksum()
		if visited[checksum] {
			continue // skip node already visited
		}

		visited[checksum] = true
		if dfsNodeCycleSearch(visited, root, subPackage) {
			return true
		}
	}

	return false
}

func (node *PackageNode) IsInCycle() bool {
	return dfsNodeCycleSearch(make(map[PackageChecksum]bool), node, nil)
}

type PackageNodeList []*PackageNode // implement sort.Interface

func (list PackageNodeList) Len() int {
	return len(list)
}

func (list PackageNodeList) Less(i, j int) bool {
	// compare hash
	var hashA []byte = list[i].Hash[:]
	var hashB []byte = list[j].Hash[:]

	cmp := bytes.Compare(hashA, hashB)

	if cmp < 0 {
		return true
	} else if cmp > 0 {
		return false
	}

	// equivalent hashes, compare size
	sizeA := list[i].Size
	sizeB := list[j].Size

	if sizeA < sizeB {
		return true
	} else {
		return false
	}
}

func (list PackageNodeList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list PackageNodeList) Search(checksum PackageChecksum) *PackageNode {
	index := sort.Search(list.Len(), func(i int) bool {
		for i, v := range list[i].Checksum() {
			if v != checksum[i] {
				return false
			}
		}

		return true
	})

	if index < list.Len() {
		return list[index]
	}

	return nil
}

func (list *PackageNodeList) Add(new *PackageNode) {
	if list.Search(new.Checksum()) != nil {
		return // node already in list
	}

	var slice []*PackageNode = *list
	slice = append(slice, new)

	*list = slice
	sort.Sort(list)
}

type PackageGraph map[PackageChecksum]*PackageNode

func NewPackageGraph() *PackageGraph {
	var graph PackageGraph = make(map[PackageChecksum]*PackageNode)
	return &graph
}

func (g *PackageGraph) Insert(name string, path string, size int,
	sha1 SHA1, hash HASH,
	subPackages ...*PackageNode) *PackageNode {
	checksum := CalculatePackageChecksum(hash, size)
	var m map[PackageChecksum]*PackageNode = *g
	if node, ok := m[checksum]; ok {
		return node
	}

	newNode := &PackageNode{
		name, path, size, sha1, hash,
		make([]*PackageNode, 0, len(subPackages)),
	}
	for _, subPackage := range subPackages {
		if subPackage == nil {
			continue
		}

		newNode.SubPackages.Add(subPackage)
	}

	m[checksum] = newNode

	return m[checksum]
}

func (g *PackageGraph) InsertHexString(name string, path string, size int,
	sha1HexString string, hashHexString string,
	subPackages ...*PackageNode) (*PackageNode, error) {
	if hashHexString == "" {
		err := errors.New(fmt.Sprintf("empty hash for %s:%s", name, hashHexString))
		return nil, err
	}

	if sha1HexString == "" {
		err := errors.New(fmt.Sprintf("empty sha1 for %s:%s", name, sha1HexString))
		return nil, err
	}

	sha1Slice, err := hex.DecodeString(sha1HexString)
	if err != nil {
		err = errors.Wrapf(err, "error decoding sha1 \"%s\"", sha1HexString)
		return nil, err
	}
	var sha1 SHA1
	for i, v := range sha1Slice {
		sha1[i] = v
	}

	hash256Slice, err := hex.DecodeString(hashHexString)
	if err != nil {
		err = errors.Wrapf(err, "error decoding hash \"%s\"", hashHexString)
		return nil, err
	}
	var hash HASH
	for i, v := range hash256Slice {
		hash[i] = v
	}

	return g.Insert(name, path, size, sha1, hash, subPackages...), nil
}
