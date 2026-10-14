package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

var modeNames = map[string]string{
	"40000":  "tree",
	"100644": "blob",
}

func (r *repo) initRepo(w io.Writer) error {
	for _, dir := range []string{
		r.gitDir(),
		path.Join(r.gitDir(), "objects"),
		path.Join(r.gitDir(), "refs"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	headFileContents := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(
		path.Join(r.gitDir(), "HEAD"), headFileContents, 0644); err != nil {
		return err
	}

	fmt.Fprintln(w, "Initialized git directory")
	return nil
}

func (r *repo) catFile(w io.Writer, objectSha string) error {
	object, err := r.readObject(objectSha)
	if err != nil {
		return err
	}
	defer object.Close()

	if _, err := io.Copy(w, object.br); err != nil {
		return err
	}
	return nil
}

func (r *repo) hashBlob(filename string) (string, error) {
	fileContents, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	header := fmt.Sprintf("blob %d\x00", len(fileContents))
	contents := append([]byte(header), fileContents...)

	return r.writeObject(contents)
}

func (r *repo) hashTree(fpath string) (string, error) {
	dirContents, err := os.ReadDir(fpath)
	if err != nil {
		return "", err
	}
	var treeFileContents []byte
	for _, content := range dirContents {
		if content.Name() == ".git" {
			continue
		}

		var line []byte
		var sha string
		var header string
		var err error

		if content.IsDir() {
			sha, err = r.hashTree(path.Join(fpath, content.Name()))
			if err != nil {
				return "", err
			}
			header = fmt.Sprintf("40000 %s\x00", content.Name())
		} else {
			sha, err = r.hashBlob(path.Join(fpath, content.Name()))
			if err != nil {
				return "", err
			}
			header = fmt.Sprintf("100644 %s\x00", content.Name())
		}

		byte_sha, _ := hex.DecodeString(sha)
		line = append([]byte(header), byte_sha...)
		treeFileContents = append(treeFileContents, line...)
	}
	header := fmt.Sprintf("tree %d\x00", len(treeFileContents))
	treeFileContents = append([]byte(header), treeFileContents...)

	return r.writeObject((treeFileContents))
}

func (r *repo) hashObject(w io.Writer, filename string) error {
	sha, err := r.hashBlob(filename)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, sha)
	return nil
}

func (r *repo) lsTree(w io.Writer, treeObjectSha string, nameOnlyFormat bool) error {
	object, err := r.readObject(treeObjectSha)
	if err != nil {
		return err
	}
	defer object.Close()

	for {
		s, err := object.br.ReadString(0)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		mode, name, found := strings.Cut(strings.TrimSuffix(s, "\x00"), " ")
		if !found {
			return fmt.Errorf("corrupt tree object")
		}

		var sha [20]byte
		if _, err := io.ReadFull(object.br, sha[:]); err != nil {
			return err
		}

		modeType, ok := modeNames[mode]
		if !ok {
			modeType = "unk"
		}
		if nameOnlyFormat {
			fmt.Fprintln(w, name)
		} else {
			fmt.Fprintln(w, mode, modeType, name, hex.EncodeToString(sha[:]))
		}
	}
	return nil
}

func (r *repo) writeTree(w io.Writer) error {
	sha, err := r.hashTree(".")
	if err != nil {
		return err
	}
	fmt.Fprintln(w, sha)
	return nil
}

func (r *repo) commitTree(
	w io.Writer,
	treeObjectSha string,
	commitSha string,
	message string,
) error {

	contents := fmt.Sprintf("tree %s\n", treeObjectSha)
	if commitSha != "" {
		contents += fmt.Sprintf("parent %s\n", commitSha)
	}
	contents += "author John Doe <john@example.com> 1234567890 +0000\n"
	contents += "committer John Doe <john@example.com> 1234567890 +0000\n"
	contents += "\n"
	contents += message + "\n"

	contents = fmt.Sprintf("commit %d\x00%s", len(contents), contents)

	sha, err := r.writeObject([]byte(contents))
	if err != nil {
		return err
	}

	fmt.Fprintln(w, sha)
	return nil
}
