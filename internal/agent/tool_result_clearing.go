package agent

import (
	"time"

	"charm.land/fantasy"
)

const (
	// clearToolResultsRatio is the context usage ratio above which old tool
	// results are proactively cleared to delay full summarization.
	clearToolResultsRatio = 0.6

	// keepRecentToolResults is the number of most recent tool result messages
	// to preserve when clearing old results.
	keepRecentToolResults = 10

	// clearedToolResultText is the placeholder text for cleared tool results.
	clearedToolResultText = "[Tool result cleared to save context]"

	// idleGapThreshold is the duration of inactivity after which old tool
	// results are cleared before the next request. When the user returns
	// after this gap, the provider's prompt cache is likely cold, so
	// clearing old results reduces the token cost of the full re-encode.
	idleGapThreshold = 30 * time.Minute

	// keepRecentToolResultsAfterIdle is the number of recent tool results
	// to keep during time-gap-based clearing. More aggressive than the
	// ratio-based clearing since the cache is cold anyway.
	keepRecentToolResultsAfterIdle = 5
)

// clearOldToolResults proactively replaces old tool result content with a
// short placeholder when context usage exceeds a threshold. This is cheaper
// than full summarization and delays the need for it. Only text-based tool
// results are cleared; errors and media are left intact. The most recent N
// tool result messages are always preserved.
func clearOldToolResults(messages []fantasy.Message, contextWindow, usedTokens int64) []fantasy.Message {
	if contextWindow <= 0 {
		return messages
	}
	ratio := float64(usedTokens) / float64(contextWindow)
	if ratio < clearToolResultsRatio {
		return messages
	}

	return clearToolResultsKeeping(messages, keepRecentToolResults)
}

// clearToolResultsAfterIdleGap checks if the gap between the last assistant
// message and now exceeds the idle threshold. If so, it aggressively clears
// old tool results since the provider's prompt cache is likely cold.
func clearToolResultsAfterIdleGap(messages []fantasy.Message, lastAssistantTime int64) []fantasy.Message {
	if lastAssistantTime <= 0 {
		return messages
	}

	gap := time.Since(time.Unix(lastAssistantTime, 0))
	if gap < idleGapThreshold {
		return messages
	}

	return clearToolResultsKeeping(messages, keepRecentToolResultsAfterIdle)
}

// clearToolResultsKeeping replaces old tool result text with a placeholder,
// keeping the most recent N tool result messages.
func clearToolResultsKeeping(messages []fantasy.Message, keep int) []fantasy.Message {
	toolResultCount := 0
	keepFromIndex := 0
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == fantasy.MessageRoleTool {
			toolResultCount++
			if toolResultCount >= keep {
				keepFromIndex = i
				break
			}
		}
	}

	for i := range messages[:keepFromIndex] {
		if messages[i].Role != fantasy.MessageRoleTool {
			continue
		}
		for j, part := range messages[i].Content {
			tr, ok := fantasy.AsMessagePart[fantasy.ToolResultPart](part)
			if !ok {
				continue
			}
			if _, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](tr.Output); !ok {
				continue
			}
			tr.Output = fantasy.ToolResultOutputContentText{
				Text: clearedToolResultText,
			}
			messages[i].Content[j] = tr
		}
	}

	return messages
}
