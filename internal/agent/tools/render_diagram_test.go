package tools

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"charm.land/fantasy"
	"github.com/stretchr/testify/require"
	"github.com/zhiqiang-hhhh/smith/internal/render"
)

func TestRenderDiagramTool(t *testing.T) {
	t.Parallel()

	srv, err := render.NewServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = srv.Shutdown(context.Background())
	})

	tool := NewRenderDiagramTool(srv)

	params := RenderDiagramParams{
		Format:      "mermaid",
		Title:       "Sequence",
		Content:     "sequenceDiagram\nAlice->>Bob: Hello",
		ExpireAfter: int((2 * time.Minute).Seconds()),
	}
	input, err := json.Marshal(params)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), SessionIDContextKey, "session-abc")
	resp, err := tool.Run(ctx, fantasy.ToolCall{
		ID:    "call-1",
		Name:  RenderDiagramToolName,
		Input: string(input),
	})
	require.NoError(t, err)
	require.False(t, resp.IsError)
	require.Contains(t, resp.Content, "URL:")
	require.Contains(t, resp.Content, srv.BaseURL())

	var meta render.RenderResult
	require.NoError(t, json.Unmarshal([]byte(resp.Metadata), &meta))
	require.Equal(t, "session-abc", meta.SessionID)
	require.Equal(t, "mermaid", meta.Format)
	require.NotEmpty(t, meta.URL)
}

func TestRenderDiagramToolRequiresSessionID(t *testing.T) {
	t.Parallel()

	srv, err := render.NewServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = srv.Shutdown(context.Background())
	})

	tool := NewRenderDiagramTool(srv)

	params := RenderDiagramParams{
		Format:  "mermaid",
		Content: "graph TD\nA-->B",
	}
	input, err := json.Marshal(params)
	require.NoError(t, err)

	_, err = tool.Run(context.Background(), fantasy.ToolCall{
		ID:    "call-2",
		Name:  RenderDiagramToolName,
		Input: string(input),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "session ID is required")
}
