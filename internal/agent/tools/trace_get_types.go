package tools

const TraceGetToolName = "trace_get"

type TraceGetParams struct {
	TraceID string `json:"trace_id" description:"The trace ID (e.g. trc_xxx) to retrieve"`
}
