package tools

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"charm.land/fantasy"
	"github.com/zhiqiang-hhhh/smith/internal/render"
)

const RenderDiagramToolName = "render_diagram"

//go:embed render_diagram.md
var renderDiagramDescription []byte

type RenderDiagramParams struct {
	Format      string `json:"format" description:"Diagram format to render. Only 'mermaid' is supported."`
	Title       string `json:"title,omitempty" description:"Optional page title for the rendered diagram."`
	Content     string `json:"content" description:"Diagram source content."`
	ExpireAfter int    `json:"expire_after,omitempty" description:"Optional time in seconds before the rendered page expires."`
}

func NewRenderDiagramTool(server *render.Server) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		RenderDiagramToolName,
		string(renderDiagramDescription),
		func(ctx context.Context, params RenderDiagramParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for render_diagram")
			}
			if server == nil {
				return fantasy.NewTextErrorResponse("render server is not available"), nil
			}

			expireAfter := time.Duration(params.ExpireAfter) * time.Second
			result, err := server.Render(sessionID, params.Format, params.Title, params.Content, expireAfter)
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			content := fmt.Sprintf("Diagram rendered successfully. URL: %s", result.URL)
			return fantasy.WithResponseMetadata(fantasy.NewTextResponse(content), result), nil
		},
	)
}
