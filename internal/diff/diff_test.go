package diff

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateDiff_BasicChange(t *testing.T) {
	t.Parallel()
	before := "line1\nline2\nline3\n"
	after := "line1\nmodified\nline3\n"

	d, additions, removals := GenerateDiff(before, after, "test.go")

	require.Contains(t, d, "+modified")
	require.Contains(t, d, "-line2")
	require.Equal(t, 1, additions)
	require.Equal(t, 1, removals)
}

func TestGenerateDiff_NoChange(t *testing.T) {
	t.Parallel()
	content := "same\ncontent\n"

	d, additions, removals := GenerateDiff(content, content, "test.go")

	require.Empty(t, d)
	require.Equal(t, 0, additions)
	require.Equal(t, 0, removals)
}

func TestGenerateDiff_EmptyToContent(t *testing.T) {
	t.Parallel()
	d, additions, removals := GenerateDiff("", "new\ncontent\n", "test.go")

	require.Contains(t, d, "+new")
	require.Contains(t, d, "+content")
	require.Equal(t, 2, additions)
	require.Equal(t, 0, removals)
}

func TestGenerateDiff_ContentToEmpty(t *testing.T) {
	t.Parallel()
	d, additions, removals := GenerateDiff("old\ncontent\n", "", "test.go")

	require.Contains(t, d, "-old")
	require.Contains(t, d, "-content")
	require.Equal(t, 0, additions)
	require.Equal(t, 2, removals)
}

func TestGenerateDiff_LeadingSlashStripped(t *testing.T) {
	t.Parallel()
	d, _, _ := GenerateDiff("a\n", "b\n", "/src/main.go")

	require.Contains(t, d, "a/src/main.go")
	require.NotContains(t, d, "a//src/main.go")
}

func TestGenerateDiff_MultipleChanges(t *testing.T) {
	t.Parallel()
	before := strings.Join([]string{"a", "b", "c", "d", "e"}, "\n")
	after := strings.Join([]string{"a", "B", "c", "D", "e"}, "\n")

	_, additions, removals := GenerateDiff(before, after, "test.go")

	require.Equal(t, 2, additions)
	require.Equal(t, 2, removals)
}
