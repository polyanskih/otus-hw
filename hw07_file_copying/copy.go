package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	if fromPath == "" || toPath == "" {
		return errors.New("source and destination paths are required")
	}
	if offset < 0 || limit < 0 {
		return errors.New("offset and limit must be non-negative")
	}

	info, err := os.Stat(fromPath)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return ErrUnsupportedFile
	}
	source, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()

	info, err = source.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return ErrUnsupportedFile
	}
	if info.Size() == 0 {
		var probe [1]byte
		n, err := source.Read(probe[:])
		if n > 0 {
			return ErrUnsupportedFile
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}
	if offset > info.Size() {
		return ErrOffsetExceedsFileSize
	}
	if _, err := source.Seek(offset, io.SeekStart); err != nil {
		return err
	}
	toPath, existingMode, err := copyDestination(toPath, info)
	if err != nil {
		return err
	}
	size := info.Size() - offset
	if limit > 0 {
		size = min(size, limit)
	}
	return writeCopy(source, toPath, existingMode, size)
}

func copyDestination(path string, source fs.FileInfo) (string, *fs.FileMode, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return path, nil, nil
	}
	if err != nil {
		return "", nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		path, err = filepath.EvalSymlinks(path)
		if err != nil {
			return "", nil, err
		}
		info, err = os.Stat(path)
		if err != nil {
			return "", nil, err
		}
	}
	if !info.Mode().IsRegular() {
		return "", nil, fmt.Errorf("destination %q: %w", path, ErrUnsupportedFile)
	}
	if os.SameFile(source, info) {
		return "", nil, errors.New("source and destination are the same file")
	}
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return "", nil, err
	}
	mode := info.Mode().Perm()
	return path, &mode, file.Close()
}

func writeCopy(source io.Reader, toPath string, existingMode *fs.FileMode, size int64) error {
	dir, _ := filepath.Split(toPath)
	if dir == "" {
		dir = "."
	}
	target, err := os.CreateTemp(dir, ".go-cp-*")
	if err != nil {
		return err
	}
	installed := false
	defer func() {
		if !installed {
			_ = os.Remove(target.Name())
		}
	}()
	progress := copyProgress{size: size}
	progress.draw(0)
	defer func() { _, _ = fmt.Fprintln(os.Stderr) }()
	_, err = io.CopyN(io.MultiWriter(target, &progress), source, size)
	if err == nil && existingMode != nil {
		err = target.Chmod(*existingMode)
	}
	if err := errors.Join(err, target.Close()); err != nil {
		return err
	}
	if err := os.Rename(target.Name(), toPath); err != nil {
		return err
	}
	installed = true
	progress.draw(100)
	return nil
}

type copyProgress struct {
	size, written int64
	lastPercent   int
}

func (p *copyProgress) Write(data []byte) (int, error) {
	p.written += int64(len(data))
	percent := int(float64(p.written) / float64(p.size) * 100)
	if percent > p.lastPercent && percent < 100 {
		p.draw(percent)
	}
	return len(data), nil
}

func (p *copyProgress) draw(percent int) {
	const width = 20
	filled := percent * width / 100
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", width-filled)
	_, _ = fmt.Fprintf(os.Stderr, "\r[%s] %3d%%", bar, percent)
	p.lastPercent = percent
}
