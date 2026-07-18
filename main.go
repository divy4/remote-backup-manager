package main

import (
	"github.com/divy4/remote-backup-manager/tree"
)

func main() {
	path := "/backup/backup/backintime/glados/dan/default/last_snapshot/backup/"
	t := tree.NewTree(path)
	t.PopulateNodes()
}

