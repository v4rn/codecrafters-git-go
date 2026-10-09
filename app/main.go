package main

import (
	"bufio"
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// Usage: your_program.sh <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintf(os.Stderr, "Logs from your program will appear here!\n")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":

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

	case "cat-file":

		blob := os.Args[3]
		filename := filepath.Join(".git", "objects", blob[:2], blob[2:])
		file, err := os.Open(filename) // compressed bytes
		if err != nil {
			log.Fatal(err)
		}
		zr, err := zlib.NewReader(file)
		defer zr.Close()

		br := bufio.NewReader(zr)

		if _, err := br.ReadString(0); err != nil {
			log.Fatal(err)
		}

		if _, err := io.Copy(os.Stdout, br); err != nil {
			log.Fatal(err)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
