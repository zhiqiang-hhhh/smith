package tools

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateString_NoTruncation(t *testing.T) {
	t.Parallel()
	result := TruncateString("hello", 10)
	if result != "hello" {
		t.Errorf("expected no truncation, got %q", result)
	}
}

func TestTruncateString_ASCII(t *testing.T) {
	t.Parallel()
	content := strings.Repeat("a", 100)
	result := TruncateString(content, 20)
	if !strings.Contains(result, "... [") {
		t.Error("expected truncation marker")
	}
	if strings.HasPrefix(result, strings.Repeat("a", 10)) {
		// First half preserved.
	} else {
		t.Error("expected first half to be preserved")
	}
}

func TestTruncateString_UTF8Safety(t *testing.T) {
	t.Parallel()
	// Mix of ASCII, CJK (3 bytes each), and emoji (4 bytes each).
	content := strings.Repeat("你好世界🎉", 100)
	result := TruncateString(content, 20)

	if !utf8.ValidString(result) {
		t.Error("truncated result contains invalid UTF-8")
	}
	if !strings.Contains(result, "... [") {
		t.Error("expected truncation marker")
	}
}

func TestTruncateString_ChineseCharacters(t *testing.T) {
	t.Parallel()
	content := strings.Repeat("测试内容", 50)
	result := TruncateString(content, 30)

	if !utf8.ValidString(result) {
		t.Error("truncated result contains invalid UTF-8")
	}
	// Verify the start and end are valid Chinese text.
	runes := []rune(content)
	expectedStart := string(runes[:15])
	if !strings.HasPrefix(result, expectedStart) {
		t.Errorf("expected start %q, got prefix of %q", expectedStart, result[:30])
	}
}

func TestTruncateString_ExactBoundary(t *testing.T) {
	t.Parallel()
	content := strings.Repeat("x", 100)
	result := TruncateString(content, 100)
	if result != content {
		t.Error("content at exact maxLen should not be truncated")
	}
}

func TestTruncateString_LineCount(t *testing.T) {
	t.Parallel()
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line content here"
	}
	content := strings.Join(lines, "\n")
	result := TruncateString(content, 40)
	if !strings.Contains(result, "lines truncated") {
		t.Error("expected line count in truncation notice")
	}
}
