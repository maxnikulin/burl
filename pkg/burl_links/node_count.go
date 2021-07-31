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

package burl_links

import (
	"encoding/json"
	"reflect"
)

type LimitCountAttrs struct {
	Count       int `json:"total"`
	TargetCount int `json:"filtered"`
}

type LimitCountNode struct {
	Attrs *LimitCountAttrs `burl:"flatten" json:"attr"`
	Node  TreeBaseNode     `burl:"flatten"` //`burl,json:"node"`
}

var _ TreeIntermediateBaseNode = (*LimitCountNode)(nil)

func (t *LimitCountNode) GetChildrenNodes() []TreeBaseNode {
	return t.Node.GetChildrenNodes()
}

func (t *LimitCountNode) OwnLinksCount() int {
	return t.Node.OwnLinksCount()
}

func (t *LimitCountNode) Empty() bool {
	return t.Node.Empty()
}

func (t *LimitCountNode) AppendChild(c TreeBaseNode) {
	t.Node.(TreeIntermediateBaseNode).AppendChild(c)
}

func reflectFlatten(m map[string]interface{}, p interface{}) {
	queue := make([]interface{}, 1, 8)
	queue[0] = p
	for len(queue) != 0 {
		p = queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if typed, ok := p.(BurlTyped); ok {
			m["_type"] = typed.BurlType()
		}
		v := reflect.ValueOf(p).Elem()
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if _, ok := field.Tag.Lookup("burl"); ok {
				queue = append(queue, v.Field(i).Interface())
				continue
			}
			name := field.Name
			if ann, ok := field.Tag.Lookup("json"); ok {
				name = ann
			}
			m[name] = v.Field(i).Interface()
		}
	}
}

func (p *LimitCountNode) MarshalJSON() ([]byte, error) {
	if node, ok := p.Node.(*TreeChildrenNode); ok {
		if _, ok := node.Props.(*FileGroupProps); ok {
			if children := node.GetChildrenNodes(); len(children) == 1 {
				return json.Marshal(children[0])
			}
		}
	}
	fields := map[string]interface{}{}
	reflectFlatten(fields, p)
	return json.Marshal(fields)
}

type DepthFirstQueueItem struct {
	Node interface{}
	Post bool
}

type DepthFirstQueue struct {
	Queue []DepthFirstQueueItem
}

func NewDepthFirstQueue(root interface{}) DepthFirstQueue {
	q := DepthFirstQueue{make([]DepthFirstQueueItem, 0, 16)}
	if root != nil {
		q.Push(root)
	}
	return q
}

func (q *DepthFirstQueue) Pop() DepthFirstQueueItem {
	tip := &q.Queue[len(q.Queue)-1]
	if tip.Post {
		q.Queue = q.Queue[:len(q.Queue)-1]
		return *tip
	}
	retval := *tip
	tip.Post = true
	return retval
}

func (q *DepthFirstQueue) Empty() bool {
	return !(len(q.Queue) > 0)
}

func (q *DepthFirstQueue) Push(node interface{}) {
	q.Queue = append(q.Queue, DepthFirstQueueItem{node, false})
}

func CountDescendants(inputTree TreeBaseNode, filter Filter) *LimitCountNode {
	if inputTree == nil {
		return nil
	}
	attrNodeStack := make([]*LimitCountNode, 0, 16)
	queue := NewDepthFirstQueue(inputTree)
	for !queue.Empty() {
		item := queue.Pop()
		node := item.Node.(TreeBaseNode)
		if item.Post {
			tip := attrNodeStack[len(attrNodeStack)-1]
			if len(attrNodeStack) > 1 {
				attrNodeStack = attrNodeStack[:len(attrNodeStack)-1]
				if tip != nil {
					count := tip.Attrs.Count
					if count > 0 {
						parent := attrNodeStack[len(attrNodeStack)-1]
						parent.Attrs.Count += count
						parent.AppendChild(tip)
					}
				}
			} else {
				break
			}
		} else {
			var newAttrNode *LimitCountNode = nil
			newNode := node.(TreeClonableBaseNode).Clone(filter)
			if newNode != nil {
				newAttrNode = &LimitCountNode{
					&LimitCountAttrs{newNode.OwnLinksCount(), 0},
					newNode,
				}
			}
			attrNodeStack = append(attrNodeStack, newAttrNode)
			// TODO bad interface, should be newNode
			children := node.GetChildrenNodes()
			for i := len(children); i > 0; i-- {
				queue.Push(children[i-1])
			}
		}
	}
	return attrNodeStack[0]
}

func FilterChildrenCount(tree *LimitCountNode, target int) *LimitCountNode {
	tree.Attrs.TargetCount = target
	queue := NewDepthFirstQueue(tree)
	for !queue.Empty() {
		item := queue.Pop()
		node := item.Node.(*LimitCountNode)
		if item.Post {
			node.Attrs.TargetCount = node.OwnLinksCount()
			for _, child := range node.Node.GetChildrenNodes() {
				node.Attrs.TargetCount += child.(*LimitCountNode).Attrs.TargetCount
			}
		} else {
			node.Node.(LimitChildrenCountNode).LimitChildrenCount(node.Attrs.TargetCount)
			// TODO to helper
			for _, child := range node.Node.GetChildrenNodes() {
				queue.Push(child)
			}
		}
	}
	return tree
}
