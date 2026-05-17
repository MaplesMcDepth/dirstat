package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path string, size int) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent directory: %v", err)
	}
	if err := os.WriteFile(path, bytes.Repeat([]byte("x"), size), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = original
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	return string(output)
}

func TestHumanSize(t *testing.T) {
	tests := map[int64]string{
		0:                   "0B",
		999:                 "999B",
		1024:                "1.0K",
		1024 * 1024:         "1.0M",
		1024 * 1024 * 5 / 2: "2.5M",
		1024 * 1024 * 1024:  "1.0G",
	}

	for input, want := range tests {
		if got := humanSize(input); got != want {
			t.Fatalf("humanSize(%d) = %q, want %q", input, got, want)
		}
	}
}

func TestDirSizeIncludesNestedFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "top.txt"), 3)
	writeFile(t, filepath.Join(root, "nested", "child.txt"), 5)

	got, err := dirSize(root)
	if err != nil {
		t.Fatalf("dirSize returned error: %v", err)
	}

	if got != 8 {
		t.Fatalf("dirSize() = %d, want 8", got)
	}
}

func TestShowTopSortsBySizeAndSkipsHiddenFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "large.txt"), 10)
	writeFile(t, filepath.Join(root, "small.txt"), 2)
	writeFile(t, filepath.Join(root, ".secret"), 99)

	output := captureStdout(t, func() {
		showTop(root, 10, -1, false, "size", false)
	})

	largeIndex := strings.Index(output, "large.txt")
	smallIndex := strings.Index(output, "small.txt")
	if largeIndex == -1 || smallIndex == -1 {
		t.Fatalf("expected visible files in output:\n%s", output)
	}
	if largeIndex > smallIndex {
		t.Fatalf("expected large.txt before small.txt in output:\n%s", output)
	}
	if strings.Contains(output, ".secret") {
		t.Fatalf("hidden file should be skipped:\n%s", output)
	}
}

func TestShowTopHonorsMaxDepth(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "top.txt"), 3)
	writeFile(t, filepath.Join(root, "nested", "child.txt"), 5)

	output := captureStdout(t, func() {
		showTop(root, 10, 0, true, "name", false)
	})

	if !strings.Contains(output, "top.txt") {
		t.Fatalf("expected top-level file in output:\n%s", output)
	}
	if !strings.Contains(output, "nested") {
		t.Fatalf("expected top-level directory in output:\n%s", output)
	}
	if strings.Contains(output, "child.txt") {
		t.Fatalf("nested child should be skipped at max depth 0:\n%s", output)
	}
}

func TestShowTopUsesHumanReadableSizes(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "blob.bin"), 2048)

	output := captureStdout(t, func() {
		showTop(root, 10, -1, true, "size", true)
	})

	if !strings.Contains(output, "2.0K") {
		t.Fatalf("expected human-readable size in output:\n%s", output)
	}
}
