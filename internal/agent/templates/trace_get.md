Retrieves a previously recorded trace by its ID. Traces are debug recordings of Smith's internal events (agent lifecycle, tool calls, errors, etc.) that were saved during previous sessions.

<when_to_use>
Use this tool when:
- The user provides a trace ID (trc_xxx) and asks you to analyze it
- You need to examine trace data from a previous session for debugging

DO NOT use this tool when:
- You want to start a new trace (use the /trace command instead)
- The trace data is already in the current conversation
</when_to_use>

<parameters>
- trace_id: The trace ID to retrieve (required, format: trc_xxx)
</parameters>
