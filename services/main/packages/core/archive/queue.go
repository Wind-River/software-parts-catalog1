// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package archive

import (
	"os"

	"github.com/pkg/errors"
)

type HeapDir struct {
	Path     string
	Priority int64
	Index    int
}

// PriorityQueue
// We've used a min priority queue from back when we ran into memory problems processing archives with large amounts of files/directories
// The theory is that smaller directories would not lead to as many new directories, so that we can keep the total amount of items in the queue low
// TODO replace with generics
func NewHeapDir(filePath string) (*HeapDir, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "NewHeapDir directory does not exist")
	}

	h := new(HeapDir)
	h.Path = filePath
	h.Priority = stat.Size()
	h.Index = -1

	return h, nil
}

type DirectoryQueue []*HeapDir

func (dq DirectoryQueue) Len() int { return len(dq) }
func (dq DirectoryQueue) Less(i, j int) bool {
	//DirectoryQueue returns lowest priority
	return dq[i].Priority < dq[j].Priority
}
func (dq DirectoryQueue) Swap(i, j int) {
	dq[i], dq[j] = dq[j], dq[i]
	dq[i].Index = i
	dq[j].Index = j
}
func (dq *DirectoryQueue) Push(x interface{}) {
	n := len(*dq)
	item := x.(*HeapDir)
	item.Index = n
	*dq = append(*dq, item)
	//fmt.Printf("Push(%s): len(dq) -> %d\n", item.Path, dq.Len())
}
func (dq *DirectoryQueue) Pop() interface{} {
	old := *dq
	n := len(old)
	item := old[n-1]
	item.Index = -1 //Mark that item is no longer in queue
	*dq = old[0 : n-1]
	//fmt.Printf("Pop -> %s: len(dq) -> %d\n", item.Path, dq.Len())
	return item
}
