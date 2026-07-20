package main

import (
	"log"

	//"github.com/divy4/remote-backup-manager/config"
	"github.com/divy4/remote-backup-manager/tree"
)

func main() {
	t := tree.NewTree()
	log.Println(t.RootPath)
}
