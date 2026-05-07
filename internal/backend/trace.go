package backend

import (
	"context"

	"github.com/zhiqiang-hhhh/smith/internal/trace"
)

// TraceSave saves a trace snapshot for a workspace session.
func (b *Backend) TraceSave(ctx context.Context, workspaceID, sessionID string, snapshot trace.Snapshot) (trace.Record, error) {
	ws, err := b.GetWorkspace(workspaceID)
	if err != nil {
		return trace.Record{}, err
	}

	return ws.Traces.Save(ctx, sessionID, snapshot)
}

// TraceGet retrieves a trace record by ID.
func (b *Backend) TraceGet(ctx context.Context, workspaceID, traceID string) (trace.Record, error) {
	ws, err := b.GetWorkspace(workspaceID)
	if err != nil {
		return trace.Record{}, err
	}

	return ws.Traces.Get(ctx, traceID)
}
