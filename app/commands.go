package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var modeNames = map[string]string{
	"40000":  "tree",
	"100644": "blob",
}

func initRepo(args []string) error {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		}
	}

	headFileContents := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
	}

	fmt.Println("Initialized git directory")
	return nil
}

func catFile(args []string) error {
	fs := flag.NewFlagSet("cat-file", flag.ExitOnError)
	_ = fs.Bool("p", true, "pretty-print object contents")
	fs.Parse(args)

	object, err := readObject(fs.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	if _, err := io.Copy(os.Stdout, object.br); err != nil {
		log.Fatal(err)
	}
	object.Close()
	return nil
}

func hashObject(args []string) error {
	fs := flag.NewFlagSet("hash-object", flag.ExitOnError)
	_ = fs.Bool("w", true, "write the contents")
	fs.Parse(args)

	contents, err := os.ReadFile(fs.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	sha := writeObject(contents)
	fmt.Println(sha)
	return nil
}

func lsTree(args []string) error {
	fs := flag.NewFlagSet("ls-tree", flag.ExitOnError)
	nameOnly := fs.Bool("name-only", false, "print only names")
	fs.Parse(args)

	object, err := readObject(fs.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	for {
		s, err := object.br.ReadString(0)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		mode, name, found := strings.Cut(strings.TrimSuffix(s, "\x00"), " ")
		if !found {
			log.Fatal("Corrupt tree object")
		}

		var sha [20]byte
		if _, err := io.ReadFull(object.br, sha[:]); err != nil {
			log.Fatal(err)
		}

		modeType, ok := modeNames[mode]
		if !ok {
			modeType = "unk"
		}
		if *nameOnly {
			fmt.Println(name)
		} else {
			fmt.Println(mode, modeType, name, hex.EncodeToString(sha[:]))
		}
	}
	object.Close()
	return nil
}
