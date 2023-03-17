// Copyright (c) 2020 Wind River Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//       http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software  distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
// OR CONDITIONS OF ANY KIND, either express or implied.

package generics

// a type-safe generic queue
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
