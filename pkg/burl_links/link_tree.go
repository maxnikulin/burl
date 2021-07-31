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
	"sort"
)

type TreeBaseNode interface {
	GetChildrenNodes() []TreeBaseNode
	Empty() bool
	OwnLinksCount() int
}

type TreeIntermediateBaseNode interface {
	TreeBaseNode
	AppendChild(TreeBaseNode)
}

type BurlTyped interface {
	BurlType() string
}

type TreeNodeProps interface {
	BurlTyped
}

type TreeClonableBaseNode interface {
	TreeBaseNode
	Clone(filter Filter) TreeBaseNode
}

type LimitChildrenCountNode interface {
	LimitChildrenCount(target int) int
}

type Link struct {
	URL         string `json:"url"`
	Description string `json:"descr,omitempty"`
	LineNo      int    `json:"lineNo"`
}

func (l *Link) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Link
		Type string `json:"_type"`
	}{
		Link: *l,
		Type: "Link",
	})
}

// Group of links that belongs directly to TreeBaseNode
// (title, before other headings), not to any descendant heading.
type TreeLeafNode struct {
	Links []*Link `json:"links"`
}

var _ TreeClonableBaseNode = (*TreeLeafNode)(nil)
var _ LimitChildrenCountNode = (*TreeLeafNode)(nil)
var _ BurlTyped = (*TreeLeafNode)(nil)

func (_ *TreeLeafNode) BurlType() string {
	return "Body"
}

func (t *TreeLeafNode) Clone(filter Filter) TreeBaseNode {
	children := make([]*Link, 0, cap(t.Links))
	for _, link := range t.Links {
		if filter(link) {
			children = append(children, link)
		}
	}
	if len(children) == 0 {
		return nil
	}
	return &TreeLeafNode{children}
}

func (t *TreeLeafNode) GetChildrenNodes() (result []TreeBaseNode) {
	return
}

func (t *TreeLeafNode) OwnLinksCount() int {
	return len(t.Links)
}

func (t *TreeLeafNode) Empty() bool {
	return len(t.Links) == 0
}

func (t *TreeLeafNode) LimitChildrenCount(target int) int {
	if len(t.Links) == 0 {
		// should not happen
		return 0
	}
	if len(t.Links) < target {
		target = len(t.Links)
	}
	links := make([]*Link, 0, target)
	seen := make(map[string]bool)
	for _, l := range t.Links {
		if _, has := seen[l.URL]; has {
			continue
		}
		seen[l.URL] = true
		links = append(links, l)
		if len(links) >= target {
			break
		}
	}
	t.Links = links
	return len(t.Links)
}

type FileProps struct {
	Path string `json:"path"`
}

var _ TreeNodeProps = (*FileProps)(nil)

func (_ *FileProps) BurlType() string {
	return "File"
}

type FileGroupProps struct{}

var _ TreeNodeProps = (*FileGroupProps)(nil)

func (_ *FileGroupProps) BurlType() string {
	return "FileGroup"
}

// Intermediate node of link tree
type TreeChildrenNode struct {
	Props    TreeNodeProps  `burl:"flatten" json:"prop"`
	Children []TreeBaseNode `json:"children"`
}

var _ TreeClonableBaseNode = (*TreeChildrenNode)(nil)
var _ LimitChildrenCountNode = (*TreeChildrenNode)(nil)
var _ TreeIntermediateBaseNode = (*TreeChildrenNode)(nil)

func NewTreeChildrenNode(props TreeNodeProps) TreeChildrenNode {
	return TreeChildrenNode{props, make([]TreeBaseNode, 0, 4)}
}

func (t *TreeChildrenNode) OwnLinksCount() int {
	return 0
}

func (t *TreeChildrenNode) Empty() bool {
	return len(t.Children) == 0
}

func (t *TreeChildrenNode) GetChildrenNodes() []TreeBaseNode {
	return t.Children
}

func (t *TreeChildrenNode) AppendChild(c TreeBaseNode) {
	t.Children = append(t.Children, c)
}

func (t TreeChildrenNode) Clone(_ Filter) TreeBaseNode {
	t.Children = make([]TreeBaseNode, 0, cap(t.Children))
	return &t
}

func (t *TreeChildrenNode) AddChild(c TreeBaseNode) {
	t.Children = append(t.Children, TreeBaseNode(c))
}

func (t *TreeChildrenNode) AddLink(link *Link) {
	var leaf *TreeLeafNode
	if len(t.Children) > 0 {
		leaf = t.Children[0].(*TreeLeafNode)
	} else {
		leaf = &TreeLeafNode{make([]*Link, 0, 4)}
		t.AddChild(leaf)
	}
	leaf.Links = append(leaf.Links, link)
}

type TreeBaseNodeSortItem struct {
	Index int
	Attrs *LimitCountAttrs
}

func (t *TreeChildrenNode) LimitChildrenCount(target int) int {
	if len(t.Children) == 0 {
		// should not happen
		return 0
	}

	sorted := make([]TreeBaseNodeSortItem, 0, len(t.Children))
	treeNodeGreater := func(a int, b int) bool {
		return sorted[a].Attrs.Count > sorted[b].Attrs.Count
	}
	for i, node := range t.Children {
		attrs := node.(*LimitCountNode).Attrs
		sorted = append(sorted, TreeBaseNodeSortItem{i, attrs})
	}
	sort.SliceStable(sorted, treeNodeGreater)
	sum := 0
	i := 0
	if target < len(sorted) {
		sorted = sorted[:target]
	}
	for ; i < len(sorted); i++ {
		count := sorted[i].Attrs.Count
		if count > 0 {
			sum += count
		} else {
			break
		}
	}
	sorted = sorted[:i]
	keep := make([]bool, len(t.Children))
	for i > 0 {
		i--
		attr := sorted[i].Attrs
		count := attr.Count
		itemCount := count * target / sum
		if itemCount == 0 {
			itemCount = 1
		}
		sum -= count
		target -= itemCount
		attr.TargetCount = itemCount
		keep[sorted[i].Index] = true
	}
	children := make([]TreeBaseNode, 0, len(sorted))
	for i, child := range t.Children {
		if keep[i] {
			children = append(children, child)
		}
	}
	t.Children = children
	return len(t.Children)
}

func ForEachLink(tree TreeBaseNode, filter Filter) {
	queue := NewDepthFirstQueue(tree)
	for !queue.Empty() {
		item := queue.Pop()
		node := item.Node.(TreeBaseNode)
		if item.Post {
			continue
		} else {
			if leaf, ok := node.(*TreeLeafNode); ok {
				for _, link := range leaf.Links {
					if !filter(link) {
						return
					}
				}
			}
			children := node.GetChildrenNodes()
			for i := len(children); i > 0; i-- {
				queue.Push(children[i-1])
			}
		}
	}
}
