package tools

const MemorySearchToolName = "memory_search"

type MemorySearchParams struct {
	Query string `json:"query" description:"The query describing what information to search for in the session transcript"`
}
