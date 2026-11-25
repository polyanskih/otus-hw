package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	const source = "0123456789"
	tests := []struct {
		name     string
		source   string
		existing string
		offset   int64
		limit    int64
		want     string
	}{
		{name: "whole file", source: source, want: source},
		{name: "offset and limit", source: source, offset: 2, limit: 3, want: "234"},
		{name: "byte range inside UTF-8 rune", source: "aёz", offset: 1, limit: 1, want: "\xd1"},
		{name: "offset without limit", source: source, offset: 7, want: "789"},
		{name: "limit past EOF", source: source, offset: 8, limit: 20, want: "89"},
		{name: "offset at EOF", source: source, offset: 10, want: ""},
		{name: "empty input", source: "", want: ""},
		{name: "overwrite longer output", source: source, existing: "stale data", limit: 2, want: "01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			from := filepath.Join(dir, "source")
			to := filepath.Join(dir, "destination")
			writeTestFile(t, from, tt.source)
			if tt.existing != "" {
				writeTestFile(t, to, tt.existing)
			}
			if err := Copy(from, to, tt.offset, tt.limit); err != nil {
				t.Fatalf("Copy: %v", err)
			}
			assertFileContents(t, to, tt.want)
		})
	}
}

func TestCopyInvalidInputPreservesDestination(t *testing.T) {
	dir := t.TempDir()
	from := filepath.Join(dir, "source")
	writeTestFile(t, from, "source content")
	tests := []struct {
		name   string
		from   string
		offset int64
		limit  int64
		want   error
	}{
		{name: "offset past EOF", from: from, offset: 100, want: ErrOffsetExceedsFileSize},
		{name: "negative offset", from: from, offset: -1},
		{name: "negative limit", from: from, limit: -1},
		{name: "missing input", from: filepath.Join(dir, "missing"), want: os.ErrNotExist},
		{name: "nonregular input", from: dir, want: ErrUnsupportedFile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			to := filepath.Join(t.TempDir(), "destination")
			const original = "original destination"
			writeTestFile(t, to, original)
			err := Copy(tt.from, to, tt.offset, tt.limit)
			if err == nil {
				t.Error("Copy unexpectedly succeeded with invalid input")
			} else if tt.want != nil && !errors.Is(err, tt.want) {
				t.Errorf("Copy returned %v, want error %v", err, tt.want)
			}
			assertFileContents(t, to, original)
		})
	}
}

func TestCopyRejectsSourceAlias(t *testing.T) {
	dir := t.TempDir()
	from := filepath.Join(dir, "source")
	to := filepath.Join(dir, "alias")
	const original = "source content"
	writeTestFile(t, from, original)
	if err := os.Link(from, to); err != nil {
		t.Fatal(err)
	}
	if err := Copy(from, to, 0, 2); err == nil {
		t.Error("Copy unexpectedly succeeded with a hard link to the source")
	}
	assertFileContents(t, from, original)
	assertFileContents(t, to, original)
}

func TestCopyRejectsUnwritableDestination(t *testing.T) {
	dir := t.TempDir()
	from := filepath.Join(dir, "source")
	to := filepath.Join(dir, "destination")
	writeTestFile(t, from, "new contents")
	writeTestFile(t, to, "original")
	if err := os.Chmod(to, 0o400); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(to, os.O_WRONLY, 0)
	if err == nil {
		_ = file.Close()
		t.Skip("the OS permits opening a read-only file for writing")
	}
	if !errors.Is(err, os.ErrPermission) {
		t.Fatal(err)
	}
	if err := Copy(from, to, 0, 0); !errors.Is(err, os.ErrPermission) {
		t.Errorf("Copy returned %v, want permission error", err)
	}
	assertFileContents(t, to, "original")
}

func TestCopyUsesResolvedDestinationParent(t *testing.T) {
	root := t.TempDir()
	from := filepath.Join(root, "source")
	writeTestFile(t, from, "content")
	nested := filepath.Join(root, "real", "nested")
	destination := filepath.Join(root, "real", "destination")
	for _, dir := range []string{nested, destination} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	link := filepath.Join(root, "shortcut")
	if err := os.Symlink(nested, link); err != nil {
		t.Fatal(err)
	}
	to := link + string(filepath.Separator) + filepath.Join("..", "destination", "output")
	if err := Copy(from, to, 0, 0); err != nil {
		t.Fatal(err)
	}
	assertFileContents(t, filepath.Join(destination, "output"), "content")
}

func TestCopyRejectsUnknownSizeRegularFile(t *testing.T) {
	const from = "/proc/self/cmdline"
	info, err := os.Stat(from)
	if err != nil || !info.Mode().IsRegular() || info.Size() != 0 {
		t.Skip("a zero-size procfs file is unavailable")
	}
	contents, err := os.ReadFile(from)
	if err != nil || len(contents) == 0 {
		t.Skip("the procfs file has no readable contents")
	}
	to := filepath.Join(t.TempDir(), "destination")
	writeTestFile(t, to, "original")
	if err := Copy(from, to, 0, 0); !errors.Is(err, ErrUnsupportedFile) {
		t.Errorf("Copy returned %v, want %v", err, ErrUnsupportedFile)
	}
	assertFileContents(t, to, "original")
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertFileContents(t *testing.T, path, want string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(contents); got != want {
		t.Errorf("contents of %s = %q, want %q", path, got, want)
	}
}
