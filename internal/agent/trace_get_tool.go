package agent

import (
	"context"
	_ "embed"
	"fmt"

	"charm.land/fantasy"
	"github.com/zhiqiang-hhhh/smith/internal/agent/tools"
	"github.com/zhiqiang-hhhh/smith/internal/trace"
)

//go:embed templates/trace_get.md
var traceGetToolDescription []byte

func (c *coordinator) traceGetTool(traceService trace.Service) fantasy.AgentTool {
	return fantasy.NewParallelAgentTool(
		tools.TraceGetToolName,
		string(traceGetToolDescription),
		func(ctx context.Context, params tools.TraceGetParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.TraceID == "" {
				return fantasy.NewTextErrorResponse("trace_id is required"), nil
			}

			record, err := traceService.Get(ctx, params.TraceID)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to get trace %s: %s", params.TraceID, err)), nil
			}

			return fantasy.NewTextResponse(fmt.Sprintf("Trace %s (session: %s, events: %d)\n\n%s", record.ID, record.SessionID, record.EventCount, record.DataJSONL)), nil
		})
}
