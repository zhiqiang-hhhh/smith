package diff

import (
	"strings"
	"time"

	"github.com/aymanbagabas/go-udiff"
)

// diffTimeout is the maximum time allowed for diff generation before
// returning a fallback message. Prevents hangs on very large files.
const diffTimeout = 5 * time.Second

// GenerateDiff creates a unified diff from two file contents. Returns the
// diff string plus the number of added and removed lines. If the diff takes
// longer than diffTimeout, returns a fallback message.
func GenerateDiff(beforeContent, afterContent, fileName string) (string, int, int) {
	fileName = strings.TrimPrefix(fileName, "/")

	type result struct {
		unified string
	}
	ch := make(chan result, 1)
	go func() {
		ch <- result{unified: udiff.Unified("a/"+fileName, "b/"+fileName, beforeContent, afterContent)}
	}()

	var unified string
	select {
	case r := <-ch:
		unified = r.unified
	case <-time.After(diffTimeout):
		return "[Diff generation timed out — files are too large or too different to diff efficiently]", 0, 0
	}

	additions := 0
	removals := 0
	for line := range strings.SplitSeq(unified, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			additions++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			removals++
		}
	}

	return unified, additions, removals
}
