package main

import (
	"fmt"
	"os"
)

var commands = map[string]func(args []string) error{
	"init":        initRepo,
	"cat-file":    catFile,
	"hash-object": hashObject,
	"ls-tree":     lsTree,
}

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	cmd, ok := commands[os.Args[1]]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", os.Args[1])
		os.Exit(1)
	}
	if err := cmd(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
