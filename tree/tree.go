package tree

import (
	"path/filepath"
	"log"
	"os"
)

type Tree struct {
	// The path to the root of the tree.
	RootPath string

	// The root Node.
	RootNode *Node

	// The nodes of the tree, indexed via the path of the node.
	Nodes map[string]*Node
}

// Create a new Tree at a given path.
func NewTree(path string) *Tree {
	// Ensure the path exists
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	// Create the base tree object
	tree := Tree{
		RootPath: absPath,
		RootNode: nil,
		Nodes: make(map[string]*Node),
	}

	// Populate nodes for files and directories
	tree.populateNodes()
	return &tree
}

// Populates filesystem info about every directory/file/symlink in the tree.
func (tree *Tree) populateNodes() {
	log.Printf("Populating backup tree at %s...\n", tree.RootPath)
	filepath.Walk(tree.RootPath, tree.populateNode)
	log.Printf("Populated %d items.\n", len(tree.Nodes))

	log.Println("Aggrigating metrics...")
	opts := DefaultWalkOps()
	opts.WalkOrder = ChildrenFirst
	tree.RootNode.Walk(bubbleUpMetrics, opts)
	log.Printf("Total backup size: %d bytes.\n", tree.RootNode.TotalSize)
}

// Populates filesystem info about a single directory/file/symlink in a tree.
func (tree *Tree) populateNode(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Fatal(err)
	}

	// Figure out who the parent is
	var parent *Node
	if path != tree.RootPath {
		parent = tree.Nodes[filepath.Dir(path)]
	}
	
	// Create the node
	node := tree.newNode(path, info, parent)
	tree.Nodes[path] = node

	return nil
}

