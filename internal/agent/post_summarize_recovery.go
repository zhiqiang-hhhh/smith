package agent

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/zhiqiang-hhhh/smith/internal/filetracker"
)

const (
	// maxRecentFiles is the maximum number of recently-read files to
	// re-inject after summarization.
	maxRecentFiles = 5

	// maxRecentFileSize is the maximum character size per re-injected file.
	maxRecentFileSize = 5_000

	// maxTotalRecentFileSize is the total character budget for all
	// re-injected files combined.
	maxTotalRecentFileSize = 50_000
)

// loadRecentlyReadFiles loads the content of the most recently read files for
// a session and formats them for re-injection into the prompt after
// summarization. This restores context about files the agent was recently
// working with.
func loadRecentlyReadFiles(ctx context.Context, ft filetracker.Service, sessionID string) string {
	if ft == nil {
		return ""
	}

	files, err := ft.ListReadFiles(ctx, sessionID)
	if err != nil {
		slog.Warn("Failed to list read files for post-summary recovery", "error", err)
		return ""
	}

	if len(files) == 0 {
		return ""
	}

	// ListReadFiles returns files ordered by read_at DESC (most recent
	// first). Take at most maxRecentFiles.
	if len(files) > maxRecentFiles {
		files = files[:maxRecentFiles]
	}

	var sb strings.Builder
	sb.WriteString("<recently_read_files>\n")
	sb.WriteString("These files were recently read before the conversation was summarized. Their content is provided here to restore context.\n\n")

	totalSize := 0
	filesAdded := 0
	for _, path := range files {
		if totalSize >= maxTotalRecentFileSize {
			break
		}

		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		text := string(content)
		remaining := maxTotalRecentFileSize - totalSize
		limit := min(maxRecentFileSize, remaining)
		if len(text) > limit {
			text = text[:limit] + "\n... [truncated]"
		}

		fmt.Fprintf(&sb, "<file path=%q>\n%s\n</file>\n\n", path, text)
		totalSize += len(text)
		filesAdded++
	}

	if filesAdded == 0 {
		return ""
	}

	sb.WriteString("</recently_read_files>")
	return sb.String()
}
