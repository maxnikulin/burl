// Copyright (C) 2021 Max Nikulin
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package burl_fuzzy

import (
	"container/heap"
	"sort"
	"strings"
)

type WeightedString struct {
	Weight EditorDistance
	Word   string
}

func (a *WeightedString) Less(b *WeightedString) bool {
	if a.Weight != b.Weight {
		return a.Weight < b.Weight
	}
	lenA := len(a.Word)
	lenB := len(b.Word)
	if lenA != lenB {
		return lenA < lenB
	}
	return a.Word < b.Word
}

type WeightedStringHeap []*WeightedString

func (h *WeightedStringHeap) Len() int { return len(*h) }

func (h *WeightedStringHeap) Less(i, j int) bool {
	a := (*h)[i]
	b := (*h)[j]
	return b.Less(a)
}

func (h WeightedStringHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *WeightedStringHeap) Push(x interface{}) {
	item := x.(*WeightedString)
	*h = append(*h, item)
}

func (h *WeightedStringHeap) Pop() interface{} {
	item := (*h)[len(*h)-1]
	*h = (*h)[0 : len(*h)-1]
	return item
}

func (h *WeightedStringHeap) Add(item *WeightedString, limit int) {
	if h.Len() < limit {
		heap.Push(h, item)
	} else if item.Less((*h)[0]) {
		(*h)[0] = item
		heap.Fix(h, 0)
	}
}

func DefaultDistanceLimit(length int) EditorDistance {
	switch {
	case length < 2:
		return 0
	case length < 4:
		return 1
	case length < 7:
		return 2
	}
	return 3
}

func GetSubqueryDistance(query string, word string, limit EditorDistance) (distance EditorDistance) {
	switch {
	case limit < 0:
		return EditorDistance(len(word)) * SubstitutionDistance
	case limit == 0:
		if strings.Contains(word, query) {
			return 0
		} else {
			return EditorDistance(len(word)) * SubstitutionDistance
		}
	}
	lenDiff := EditorDistance(len(query)-len(word)) * InsertionDistance
	if lenDiff > limit {
		return lenDiff
	}
	return SearchDistanceDL(query, word)
}

type SubqueryLimit struct {
	query string
	limit EditorDistance
}

func MakeSubqueryLimit(query []string, limit EditorDistance) []SubqueryLimit {
	result := make([]SubqueryLimit, 0, len(query))
	for _, w := range query {
		result = append(result, SubqueryLimit{w, MinDistance(limit, DefaultDistanceLimit(len(w)))})
	}
	return result
}

func QueryDistance(query []SubqueryLimit, word string, limit EditorDistance) (ok bool, distance EditorDistance) {
	for _, q := range query {
		distance += GetSubqueryDistance(q.query, word, limit-distance)
		if distance > limit {
			return false, distance
		}
	}
	return true, distance
}

// TODO rewrite using channel
// TODO return err if limit is out of range
func ProcessUrlSource(queryWords []string, source func(func(string) bool), limitPtr *int, tolerancePtr *EditorDistance) WeightedStringHeap {
	limit := 10
	if limitPtr != nil {
		limit = *limitPtr
	}
	toleranceLimit := EditorDistance(5)
	if tolerancePtr != nil {
		toleranceLimit = *tolerancePtr
	}
	h := WeightedStringHeap{}
	heap.Init(&h)
	subqueryLimit := MakeSubqueryLimit(queryWords, toleranceLimit) // len(queryWords)*3/2) // FIXME word length

	cb := func(word string) bool {
		ok, distance := QueryDistance(subqueryLimit, word, toleranceLimit)
		if !ok {
			return true
		}
		h.Add(&WeightedString{distance, word}, limit)
		return true
	}
	source(cb)

	sort.Sort(&h)
	return h
}
