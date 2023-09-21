package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"github.com/codeclysm/extract/v3"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	ZIP = iota
	TAR
	BUFFERFILE string = "buffer"
)

type Folder struct {
	childs []*Folder
	files  []File
	parent *Folder
}

type File struct {
	name string
	size uint
}

type FileSystem struct {
	zipR        *zip.ReadCloser
	tarR        fs.FS
	currentPath []string
	mode        int
}

func newFileSystem(path string, mode int) (FileSystem, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileSystem{}, err
	}
	buffer := bytes.NewBuffer(data)

	switch mode {
	case ZIP:
		err := extract.Zip(context.TODO(), buffer, BUFFERFILE, nil)
		if err != nil {
			return FileSystem{}, err
		}
		return FileSystem{currentPath: []string{BUFFERFILE, strings.TrimRight(path, ".zip")}, mode: ZIP}, nil
	case TAR:
		err := extract.Tar(context.TODO(), buffer, BUFFERFILE, nil)
		if err != nil {
			return FileSystem{}, err
		}
		return FileSystem{currentPath: []string{BUFFERFILE,
			strings.TrimRight(strings.TrimRight(path, ".tar"), ".gz")}, mode: TAR}, nil
	}
	return FileSystem{}, fmt.Errorf("unsupported file extension: %d", mode)
}

func (fs *FileSystem) close() error {
	return os.RemoveAll(BUFFERFILE)
}

func (fs *FileSystem) cd(path string, cmd int) error {
	defer func() {

	}()
	switch cmd {
	case CD_BACK:
		if len(fs.currentPath) > 2 {
			(*fs).currentPath = fs.currentPath[:len(fs.currentPath)-1]
		}
		return nil
	case CD_TO:
		if strings.Count(path, "/") == len(path) {
			(*fs).currentPath = fs.currentPath[:2]
			return nil
		}

		if f, err := os.Stat(filepath.Join(fs.currentPath...) + string(os.PathSeparator) + path); !f.IsDir() || err != nil {
			if err != nil {
				return err
			}
			return fmt.Errorf("%s is folder. `cd` can ONLY use for directories", f.Name())
		}

		(*fs).currentPath = append(fs.currentPath, strings.Split(path, "/")...)
		return nil
	}
	return fmt.Errorf("unexpected command: %d", cmd)
}

func (fs *FileSystem) ls(path string, cmd int) ([]string, error) {
	var (
		f   *os.File
		err error
	)
	defer f.Close()

	switch cmd {
	case LS:
		f, err = os.Open(filepath.Join(fs.currentPath...))
		if err != nil {
			return nil, err
		}
	case LS_TO:
		f, err = os.Open(filepath.Join(fs.currentPath...) + string(os.PathSeparator) + path)
		if err != nil {
			return nil, err
		}
	}

	files, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	res := make([]string, 0)
	for _, file := range files {
		res = append(res, file.Name())
	}
	return res, nil
}

func (fs *FileSystem) pwd() string {
	return "." + string(os.PathSeparator) + filepath.Join(fs.currentPath[1:]...)
}

func (fs *FileSystem) terminalPWD() string {
	path := "~"
	if len(fs.currentPath) > 2 {
		path = path + strings.Join(fs.currentPath[2:], string(os.PathSeparator))
	}
	return path
}

func (fs *FileSystem) cat(from string, to *os.File) error {
	f, err := os.Open(filepath.Join(fs.currentPath...) + string(os.PathSeparator) + from)
	defer f.Close()

	if err != nil {
		return err
	}

	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is folder. `cat` can ONLY use for files", from)
	}

	_, err = io.Copy(to, f)
	return err
}
