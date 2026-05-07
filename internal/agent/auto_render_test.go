package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zhiqiang-hhhh/smith/internal/message"
	"github.com/zhiqiang-hhhh/smith/internal/render"
)

func TestMermaidBlockRegex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single block",
			input:    "Here is a diagram:\n```mermaid\nsequenceDiagram\n    A->>B: Hello\n```\nDone.",
			expected: []string{"sequenceDiagram\n    A->>B: Hello"},
		},
		{
			name:     "multiple blocks",
			input:    "```mermaid\ngraph TD\n    A-->B\n```\ntext\n```mermaid\nflowchart LR\n    X-->Y\n```",
			expected: []string{"graph TD\n    A-->B", "flowchart LR\n    X-->Y"},
		},
		{
			name:     "no blocks",
			input:    "No mermaid here, just ```go\nfmt.Println()\n```",
			expected: nil,
		},
		{
			name:     "empty mermaid block",
			input:    "```mermaid\n\n```",
			expected: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matches := mermaidBlockRegex.FindAllStringSubmatch(tt.input, -1)
			if tt.expected == nil {
				assert.Empty(t, matches)
				return
			}
			require.Len(t, matches, len(tt.expected))
			for i, m := range matches {
				assert.Equal(t, tt.expected[i], m[1])
			}
		})
	}
}

func TestAutoRenderMermaidBlocks(t *testing.T) {
	t.Parallel()

	srv, err := render.NewServer()
	require.NoError(t, err)
	t.Cleanup(func() { srv.Shutdown(t.Context()) })

	env := testEnv(t)
	sess, err := env.sessions.Create(t.Context(), "test")
	require.NoError(t, err)

	t.Run("renders mermaid block and appends URL", func(t *testing.T) {
		agent := &sessionAgent{
			renderServer: srv,
			messages:     env.messages,
		}

		created, err := env.messages.Create(t.Context(), sess.ID, message.CreateMessageParams{
			Role: message.Assistant,
			Parts: []message.ContentPart{
				message.TextContent{Text: "Here:\n```mermaid\nsequenceDiagram\n    A->>B: hi\n```"},
			},
		})
		require.NoError(t, err)

		agent.autoRenderMermaidBlocks(t.Context(), sess.ID, &created)

		require.True(t, len(created.Parts) > 1, "expected additional part with URL")
		lastPart := created.Parts[len(created.Parts)-1]
		tc, ok := lastPart.(message.TextContent)
		require.True(t, ok)
		assert.Contains(t, tc.Text, "Diagram rendered:")
		assert.Contains(t, tc.Text, srv.BaseURL())
	})

	t.Run("skips when render_diagram tool was already called", func(t *testing.T) {
		agent := &sessionAgent{
			renderServer: srv,
			messages:     env.messages,
		}

		created, err := env.messages.Create(t.Context(), sess.ID, message.CreateMessageParams{
			Role: message.Assistant,
			Parts: []message.ContentPart{
				message.TextContent{Text: "```mermaid\ngraph TD\n    A-->B\n```"},
				message.ToolCall{ID: "tc-1", Name: "render_diagram", Input: "{}"},
			},
		})
		require.NoError(t, err)

		agent.autoRenderMermaidBlocks(t.Context(), sess.ID, &created)

		assert.Len(t, created.Parts, 2, "should not add extra parts")
	})

	t.Run("no-op when no mermaid blocks", func(t *testing.T) {
		agent := &sessionAgent{
			renderServer: srv,
			messages:     env.messages,
		}

		created, err := env.messages.Create(t.Context(), sess.ID, message.CreateMessageParams{
			Role: message.Assistant,
			Parts: []message.ContentPart{
				message.TextContent{Text: "No diagrams here."},
			},
		})
		require.NoError(t, err)

		agent.autoRenderMermaidBlocks(t.Context(), sess.ID, &created)

		assert.Len(t, created.Parts, 1)
	})
}
