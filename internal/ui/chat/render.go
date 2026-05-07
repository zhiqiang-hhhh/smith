package chat

import (
	"encoding/json"

	"github.com/zhiqiang-hhhh/smith/internal/agent/tools"
	"github.com/zhiqiang-hhhh/smith/internal/message"
	"github.com/zhiqiang-hhhh/smith/internal/ui/styles"
)

// RenderDiagramToolMessageItem is a message item that represents a render_diagram tool call.
type RenderDiagramToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*RenderDiagramToolMessageItem)(nil)

// NewRenderDiagramToolMessageItem creates a new [RenderDiagramToolMessageItem].
func NewRenderDiagramToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &RenderDiagramToolRenderContext{}, canceled)
}

// RenderDiagramToolRenderContext renders render_diagram tool messages.
type RenderDiagramToolRenderContext struct{}

// RenderTool implements the [ToolRenderer] interface.
func (r *RenderDiagramToolRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	if opts.IsPending() {
		return pendingTool(sty, "Render Diagram", opts.Anim, opts.Compact, opts.CreatedAt)
	}

	var params tools.RenderDiagramParams
	if err := json.Unmarshal([]byte(opts.ToolCall.Input), &params); err != nil {
		return toolErrorContent(sty, &message.ToolResult{Content: "Invalid parameters"}, cappedWidth)
	}

	toolParams := []string{params.Format}
	if params.Title != "" {
		toolParams = append(toolParams, "title", params.Title)
	}

	header := toolHeader(sty, opts.Status, "Render Diagram", cappedWidth, opts.Compact, toolParams...)
	if opts.Compact {
		return header
	}

	if earlyState, ok := toolEarlyStateContent(sty, opts, cappedWidth); ok {
		return joinToolParts(header, earlyState)
	}

	if opts.HasEmptyResult() {
		return header
	}

	bodyWidth := cappedWidth - toolBodyLeftPaddingTotal
	body := sty.Tool.Body.Render(toolOutputPlainContent(sty, opts.Result.Content, bodyWidth, opts.ExpandedContent))
	return joinToolParts(header, body)
}
