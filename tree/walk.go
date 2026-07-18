package tree

import "log"

type WalkFunc func(node *Node)

type WalkOrder int64
const (
	ParentsFirst WalkOrder = iota
	ChildrenFirst
)

type WalkOpts struct {
	// The order the nodes should be iterated through.
	WalkOrder WalkOrder
}

// Walk through a tree at a given start point.
func (tree *Tree) Walk(startPath string, walkFn WalkFunc, walkOpts *WalkOpts) {
	tree.Nodes[startPath].Walk(walkFn, walkOpts)
}

// Walk through a tree at a given node.
func (node *Node) Walk(walkFn WalkFunc, opts *WalkOpts) {
	if walkFn == nil {
		log.Panic("walkFn cannot be nil!")
	}

	// Keep a stack of nodes to visit and what index of it's children we're at
	var nodes []*Node
	var child []int
	nodes = append(nodes, node)
	child = append(child, 0)

	i := 0
	for i >= 0 {
		// If we're visiting parents first and this is the first time we're seeing
		// the node, visit it.
		if opts.WalkOrder == ParentsFirst && child[i] == 0 {
			walkFn(nodes[i])
		}

		// If we haven't visited every child yet, add the next child to the stack,
		// increment the current child counter, and move one layer deeper.
		if child[i] < len(nodes[i].Children) {
			nodes = append(nodes, nodes[i].Children[child[i]])
			child = append(child, 0)
			child[i]++
			i++

		// If we have visited all children, move one layer up.
		} else {
			// If we're visiting parents last, visit the node on our way out.
			if opts.WalkOrder == ChildrenFirst {
				walkFn(nodes[i])
			}
			nodes = nodes[:i]
			child = child[:i]
			i--
		}
	}
}

// Returns the default opts for walking through a tree.
func DefaultWalkOps() *WalkOpts {
	return &WalkOpts{
		WalkOrder: ParentsFirst,
	}
}
