package main

import (
	"bufio"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

type repo struct {
	root string // directory containing .git
}

type object struct {
	br      *bufio.Reader
	closers []io.Closer
}

func (o *object) Close() {
	for _, c := range o.closers {
		c.Close()
	}
}

func (r *repo) gitDir() string { return filepath.Join(r.root, ".git") }

func (r *repo) objectPath(sha string) string {
	return filepath.Join(r.gitDir(), "objects", sha[:2], sha[2:])
}

func (r *repo) objectDir(sha string) string {
	return filepath.Join(r.gitDir(), "objects", sha[:2])
}

func (r *repo) readObject(sha string) (*object, error) {
	f, err := os.Open(r.objectPath(sha))
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

func (r *repo) writeObject(contents []byte) (string, error) {

	sum := sha1.Sum(contents)
	sha := hex.EncodeToString(sum[:])

	if err := os.MkdirAll(r.objectDir(sha), 0775); err != nil {
		return "", err
	}
	file, err := os.Create(r.objectPath(sha))
	if err != nil {
		return "", err
	}
	zw := zlib.NewWriter(file)
	if _, err := zw.Write(contents); err != nil {
		file.Close()
		return "", err
	}
	if err := zw.Close(); err != nil {
		file.Close()
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	return sha, nil
}
