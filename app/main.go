package main

import (
	"bufio"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var modeNames = map[string]string{
	"40000":  "tree",
	"100644": "blob",
	"100755": "blob",
	"120000": "blob",
	"160000": "commit",
}

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

	case "hash-object":

		// build blob contents and create a sha1
		filename := os.Args[3]
		contents, err := os.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}
		header := fmt.Sprintf("blob %d\x00", len(contents))
		blob := append([]byte(header), contents...)

		sum := sha1.Sum(blob)
		sha := hex.EncodeToString(sum[:])

		// build the blob object file
		blob_path := path.Join(".git", "objects", sha[:2])
		if err := os.MkdirAll(blob_path, 0775); err != nil {
			log.Fatal(err)
		}
		file, err := os.Create(path.Join(blob_path, sha[2:]))
		if err != nil {
			log.Fatal(err)
		}
		zw := zlib.NewWriter(file)
		zw.Write(blob)
		zw.Close()

		fmt.Println(sha)

	case "ls-tree":

		tree := os.Args[3]
		filename := filepath.Join(".git", "objects", tree[:2], tree[2:])
		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		zr, err := zlib.NewReader(file)
		defer zr.Close()

		br := bufio.NewReader(zr)

		if _, err := br.ReadString(0); err != nil {
			log.Fatal(err)
		}

		for {
			s, err := br.ReadString(0)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			_, name, _ := strings.Cut(strings.TrimSuffix(s, "\x00"), " ")

			var sha [20]byte
			if _, err := io.ReadFull(br, sha[:]); err != nil {
				log.Fatal(err)
			}

			fmt.Println(name)
			// modeType, ok := modeNames[mode]
			// if !ok {
			// 	modeType = "unk"
			// }
			// fmt.Println(mode, modeType, name, hex.EncodeToString(sha[:]))
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
