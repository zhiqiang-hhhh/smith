package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type mockFileTracker struct {
	files []string
}

func (m *mockFileTracker) RecordRead(_ context.Context, _, _ string)             {}
func (m *mockFileTracker) LastReadTime(_ context.Context, _, _ string) time.Time { return time.Time{} }
func (m *mockFileTracker) ListReadFiles(_ context.Context, _ string) ([]string, error) {
	return m.files, nil
}

func TestLoadRecentlyReadFiles_NilTracker(t *testing.T) {
	t.Parallel()
	result := loadRecentlyReadFiles(context.Background(), nil, "session-1")
	if result != "" {
		t.Errorf("expected empty string for nil tracker, got %q", result)
	}
}

func TestLoadRecentlyReadFiles_NoFiles(t *testing.T) {
	t.Parallel()
	ft := &mockFileTracker{files: nil}
	result := loadRecentlyReadFiles(context.Background(), ft, "session-1")
	if result != "" {
		t.Errorf("expected empty string for no files, got %q", result)
	}
}

func TestLoadRecentlyReadFiles_ReturnsContent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f1 := filepath.Join(dir, "main.go")
	f2 := filepath.Join(dir, "util.go")
	os.WriteFile(f1, []byte("package main\nfunc main() {}"), 0o644)
	os.WriteFile(f2, []byte("package main\nfunc util() {}"), 0o644)

	ft := &mockFileTracker{files: []string{f1, f2}}
	result := loadRecentlyReadFiles(context.Background(), ft, "session-1")

	if !strings.Contains(result, "<recently_read_files>") {
		t.Error("expected <recently_read_files> tag")
	}
	if !strings.Contains(result, "package main") {
		t.Error("expected file content in output")
	}
	if !strings.Contains(result, f1) {
		t.Error("expected file path in output")
	}
}

func TestLoadRecentlyReadFiles_LimitsToMaxFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var files []string
	for i := range 10 {
		f := filepath.Join(dir, "file"+string(rune('a'+i))+".go")
		os.WriteFile(f, []byte("content"), 0o644)
		files = append(files, f)
	}

	ft := &mockFileTracker{files: files}
	result := loadRecentlyReadFiles(context.Background(), ft, "session-1")

	// Should only include first maxRecentFiles (5).
	count := strings.Count(result, "<file path=")
	if count != maxRecentFiles {
		t.Errorf("expected %d files, got %d", maxRecentFiles, count)
	}
}

func TestLoadRecentlyReadFiles_TruncatesLargeFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "big.go")
	os.WriteFile(f, []byte(strings.Repeat("x", maxRecentFileSize+1000)), 0o644)

	ft := &mockFileTracker{files: []string{f}}
	result := loadRecentlyReadFiles(context.Background(), ft, "session-1")

	if !strings.Contains(result, "... [truncated]") {
		t.Error("expected truncation marker")
	}
}

func TestLoadRecentlyReadFiles_SkipsMissingFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	existing := filepath.Join(dir, "exists.go")
	os.WriteFile(existing, []byte("package main"), 0o644)

	ft := &mockFileTracker{files: []string{
		filepath.Join(dir, "nonexistent.go"),
		existing,
	}}
	result := loadRecentlyReadFiles(context.Background(), ft, "session-1")

	if !strings.Contains(result, "exists.go") {
		t.Error("expected existing file in output")
	}
	count := strings.Count(result, "<file path=")
	if count != 1 {
		t.Errorf("expected 1 file, got %d", count)
	}
}
