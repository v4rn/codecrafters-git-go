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
)

type object struct {
	br      *bufio.Reader
	closers []io.Closer
}

func (o *object) Close() {
	for _, c := range o.closers {
		c.Close()
	}
}

func readObject(sha string) (*object, error) {
	f, err := os.Open(filepath.Join(".git", "objects", sha[:2], sha[2:]))
	if err != nil {
		return nil, err
	}
	zr, err := zlib.NewReader(f)
	if err != nil {
		f.Close()
		return nil, err
	}

	br := bufio.NewReader(zr)

	if _, err := br.ReadString(0); err != nil {
		zr.Close()
		f.Close()
		return nil, err
	}
	return &object{br, []io.Closer{zr, f}}, nil
}

func writeObject(contents []byte) string {

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

	return sha
}
