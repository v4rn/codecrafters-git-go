package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	var err error = nil

	r := &repo{root: "."}

	switch os.Args[1] {
	case "init":
		err = r.initRepo(os.Stdout)

	case "cat-file":
		fs := flag.NewFlagSet("cat-file", flag.ExitOnError)
		_ = fs.Bool("p", true, "pretty-print object contents")
		fs.Parse(os.Args[2:])

		err = r.catFile(os.Stdout, fs.Arg(0))

	case "hash-object":
		fs := flag.NewFlagSet("hash-object", flag.ExitOnError)
		_ = fs.Bool("w", true, "write the contents")
		fs.Parse(os.Args[2:])

		err = r.hashObject(os.Stdout, fs.Arg(0))

	case "ls-tree":
		fs := flag.NewFlagSet("ls-tree", flag.ExitOnError)
		nameOnly := fs.Bool("name-only", false, "print only names")
		fs.Parse(os.Args[2:])

		err = r.lsTree(os.Stdout, fs.Arg(0), *nameOnly)

	case "write-tree":
		err = r.writeTree(os.Stdout)

	case "commit-tree":
		fs := flag.NewFlagSet("commit-tree", flag.ExitOnError)
		commitSha := fs.String("p", "", "commit sha")
		message := fs.String("m", "", "commit message")

		treeSha := os.Args[2]
		fs.Parse(os.Args[3:])

		if *message == "" {
			fs.Usage()
			os.Exit(1)
		}

		err = r.commitTree(os.Stdout, treeSha, *commitSha, *message)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", os.Args[1])
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
