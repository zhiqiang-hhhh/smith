package message

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func makeTestAttachments(n int, contentSize int) []Attachment {
	attachments := make([]Attachment, n)
	content := []byte(strings.Repeat("x", contentSize))
	for i := range n {
		attachments[i] = Attachment{
			FilePath: fmt.Sprintf("/path/to/file%d.txt", i),
			MimeType: "text/plain",
			Content:  content,
		}
	}
	return attachments
}

func TestRepairUnfinished(t *testing.T) {
	t.Parallel()

	t.Run("noop for user messages", func(t *testing.T) {
		t.Parallel()
		m := Message{Role: User, Parts: []ContentPart{TextContent{Text: "hello"}}}
		m.RepairUnfinished()
		require.False(t, m.IsFinished())
	})

	t.Run("noop for already finished assistant", func(t *testing.T) {
		t.Parallel()
		m := Message{Role: Assistant, Parts: []ContentPart{
			TextContent{Text: "done"},
			Finish{Reason: FinishReasonEndTurn},
		}}
		m.RepairUnfinished()
		require.Equal(t, FinishReasonEndTurn, m.FinishReason())
	})

	t.Run("repairs empty assistant message", func(t *testing.T) {
		t.Parallel()
		m := Message{Role: Assistant, Parts: []ContentPart{}}
		m.RepairUnfinished()
		require.True(t, m.IsFinished())
		require.Equal(t, FinishReasonError, m.FinishReason())
		require.Equal(t, "Interrupted", m.FinishPart().Message)
	})

	t.Run("repairs assistant with partial content", func(t *testing.T) {
		t.Parallel()
		m := Message{Role: Assistant, Parts: []ContentPart{
			TextContent{Text: "partial response"},
		}}
		m.RepairUnfinished()
		require.True(t, m.IsFinished())
		require.Equal(t, FinishReasonError, m.FinishReason())
	})

	t.Run("marks unfinished tool calls as finished", func(t *testing.T) {
		t.Parallel()
		m := Message{Role: Assistant, Parts: []ContentPart{
			ToolCall{ID: "tc1", Name: "bash", Finished: false},
			ToolCall{ID: "tc2", Name: "view", Finished: true},
		}}
		m.RepairUnfinished()
		require.True(t, m.IsFinished())
		tcs := m.ToolCalls()
		require.Len(t, tcs, 2)
		require.True(t, tcs[0].Finished)
		require.True(t, tcs[1].Finished)
	})
}

func BenchmarkPromptWithTextAttachments(b *testing.B) {
	cases := []struct {
		name        string
		numFiles    int
		contentSize int
	}{
		{"1file_100bytes", 1, 100},
		{"5files_1KB", 5, 1024},
		{"10files_10KB", 10, 10 * 1024},
		{"20files_50KB", 20, 50 * 1024},
	}

	for _, tc := range cases {
		attachments := makeTestAttachments(tc.numFiles, tc.contentSize)
		prompt := "Process these files"

		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = PromptWithTextAttachments(prompt, attachments)
			}
		})
	}
}
