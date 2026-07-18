package tree

import (
	"os"
	"time"
)

type Node struct {
	// The tree the node belongs to
	Tree *Tree

	// The parent node. Set to nil on the root node.
	Parent *Node

	// The children of the node.
	Children []*Node

	// The root node of the section.
	SectionRoot *Node

	// The absolute path of the node.
	Path string

	// If the node is a directory.
	IsDir bool

	// If the node is a file.
	IsFile bool

	// If the node is a symbolic link.
	IsSymLink bool

	// The file size of the individual directory/file/symlink.
	IndividualSize int64

	// The file size of the directory/file/symlink and everything underneath it
	// that is also part of the same section.
	SectionSize int64

	// The file size of the directory/file/symlink and everything underneath it.
	// Equal to IndividualSize when not a directory.
	TotalSize int64

	// When the file/directory was last modified
	ModTime time.Time
}

// Creates a new node. Assumes the parent has already been created by this
// function.
func (tree *Tree) newNode(path string, info os.FileInfo, parent *Node) *Node {
	// Create node
	node := Node {
		Tree: tree,
		Parent: parent,
		SectionRoot: nil, // Filled in later in this function
		Children: make([]*Node, 0),
		Path: path,

		IsDir: info.IsDir(),
		IsFile: info.Mode().IsRegular(),
		IsSymLink: info.Mode() & os.ModeSymlink > 0,

		IndividualSize: info.Size(),
		// Fully populated later via BubbleUpMetrics
		SectionSize: info.Size(),
		TotalSize: info.Size(),

		ModTime: info.ModTime(),
	}

	// If we're at root, assign the tree root set itself as the section root.
	if parent == nil {
		tree.RootNode = &node
		node.SectionRoot = &node
	// Otherwise, assign the parent's section root as this node's section root.
	} else {
		node.SectionRoot = node.Parent.SectionRoot
	}

	// Add node as child of parent
	if parent != nil {
		parent.Children = append(parent.Children, &node)
	}
	return &node
}

// Bubbles up aggregate info (e.g. TotalSize) from a node to it's parent.
// Assumes the children of this node have already executed this function.
func bubbleUpMetrics(node *Node) {
	// Can't bubble up from the root
	if node.IsRoot() {
		return
	}
	node.Parent.TotalSize += node.TotalSize
	if node.SectionRoot == node.Parent.SectionRoot {
		node.Parent.SectionSize += node.SectionSize
	}
}

// Basic attributes

// Check if the node is the root node.
func (node *Node) IsRoot() bool {
	return node.Path == node.Tree.RootPath
}


