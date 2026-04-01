package agent

import (
	"testing"
	"time"

	"charm.land/fantasy"
)

func makeToolResultMessage(toolCallID, text string) fantasy.Message {
	return fantasy.Message{
		Role: fantasy.MessageRoleTool,
		Content: []fantasy.MessagePart{
			fantasy.ToolResultPart{
				ToolCallID: toolCallID,
				Output: fantasy.ToolResultOutputContentText{
					Text: text,
				},
			},
		},
	}
}

func makeUserMessage(text string) fantasy.Message {
	return fantasy.NewUserMessage(text)
}

func makeAssistantMessage(text string) fantasy.Message {
	return fantasy.Message{
		Role: fantasy.MessageRoleAssistant,
		Content: []fantasy.MessagePart{
			fantasy.TextPart{Text: text},
		},
	}
}

func TestClearOldToolResults_BelowThreshold(t *testing.T) {
	t.Parallel()
	msgs := []fantasy.Message{
		makeUserMessage("hello"),
		makeAssistantMessage("hi"),
		makeToolResultMessage("t1", "result1"),
	}
	result := clearOldToolResults(msgs, 100_000, 50_000) // 50% < 60%
	tr, ok := fantasy.AsMessagePart[fantasy.ToolResultPart](result[2].Content[0])
	if !ok {
		t.Fatal("expected ToolResultPart")
	}
	text, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](tr.Output)
	if !ok {
		t.Fatal("expected text output")
	}
	if text.Text != "result1" {
		t.Errorf("expected tool result to be preserved, got %q", text.Text)
	}
}

func TestClearOldToolResults_AboveThreshold(t *testing.T) {
	t.Parallel()
	// Create 15 tool result messages so 5 old ones get cleared
	// (keepRecentToolResults = 10).
	msgs := []fantasy.Message{makeUserMessage("hello")}
	for i := range 15 {
		msgs = append(msgs, makeToolResultMessage(
			"t"+string(rune('a'+i)),
			"long result content "+string(rune('a'+i)),
		))
	}
	result := clearOldToolResults(msgs, 100_000, 70_000) // 70% > 60%

	// First 5 tool results (indices 1-5) should be cleared.
	for i := 1; i <= 5; i++ {
		tr, ok := fantasy.AsMessagePart[fantasy.ToolResultPart](result[i].Content[0])
		if !ok {
			t.Fatalf("message %d: expected ToolResultPart", i)
		}
		text, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](tr.Output)
		if !ok {
			t.Fatalf("message %d: expected text output", i)
		}
		if text.Text != clearedToolResultText {
			t.Errorf("message %d: expected cleared text, got %q", i, text.Text)
		}
	}

	// Last 10 tool results (indices 6-15) should be preserved.
	for i := 6; i <= 15; i++ {
		tr, ok := fantasy.AsMessagePart[fantasy.ToolResultPart](result[i].Content[0])
		if !ok {
			t.Fatalf("message %d: expected ToolResultPart", i)
		}
		text, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](tr.Output)
		if !ok {
			t.Fatalf("message %d: expected text output", i)
		}
		if text.Text == clearedToolResultText {
			t.Errorf("message %d: expected preserved text, got cleared", i)
		}
	}
}

func TestClearOldToolResults_FewerThanKeep(t *testing.T) {
	t.Parallel()
	msgs := []fantasy.Message{
		makeUserMessage("hello"),
		makeToolResultMessage("t1", "result1"),
		makeToolResultMessage("t2", "result2"),
	}
	result := clearOldToolResults(msgs, 100_000, 80_000) // 80% > 60%

	// Only 2 tool results which is less than keepRecentToolResults (10),
	// so all should be preserved.
	for _, i := range []int{1, 2} {
		tr, ok := fantasy.AsMessagePart[fantasy.ToolResultPart](result[i].Content[0])
		if !ok {
			t.Fatal("expected ToolResultPart")
		}
		text, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](tr.Output)
		if !ok {
			t.Fatal("expected text output")
		}
		if text.Text == clearedToolResultText {
			t.Errorf("message %d: expected preserved text, got cleared", i)
		}
	}
}

func TestClearOldToolResults_PreservesErrorResults(t *testing.T) {
	t.Parallel()
	msgs := []fantasy.Message{makeUserMessage("hello")}
	for range 15 {
		msgs = append(msgs, fantasy.Message{
			Role: fantasy.MessageRoleTool,
			Content: []fantasy.MessagePart{
				fantasy.ToolResultPart{
					ToolCallID: "e1",
					Output: fantasy.ToolResultOutputContentError{
						Error: nil,
					},
				},
			},
		})
	}
	result := clearOldToolResults(msgs, 100_000, 70_000)

	// Error results should never be cleared.
	for i := 1; i < len(result); i++ {
		tr, ok := fantasy.AsMessagePart[fantasy.ToolResultPart](result[i].Content[0])
		if !ok {
			t.Fatal("expected ToolResultPart")
		}
		if _, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentError](tr.Output); !ok {
			t.Errorf("message %d: expected error output type to be preserved", i)
		}
	}
}

func TestClearOldToolResults_ZeroContextWindow(t *testing.T) {
	t.Parallel()
	msgs := []fantasy.Message{
		makeToolResultMessage("t1", "result"),
	}
	result := clearOldToolResults(msgs, 0, 0)
	tr, _ := fantasy.AsMessagePart[fantasy.ToolResultPart](result[0].Content[0])
	text, _ := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](tr.Output)
	if text.Text != "result" {
		t.Error("expected no clearing when context window is 0")
	}
}

func TestClearOldToolResults_SkipsNonToolMessages(t *testing.T) {
	t.Parallel()
	msgs := []fantasy.Message{
		makeUserMessage("hello"),
		makeAssistantMessage("hi"),
		makeUserMessage("world"),
	}
	// Ensure non-tool messages are completely untouched.
	for range 15 {
		msgs = append(msgs, makeToolResultMessage("t", "data"))
	}
	result := clearOldToolResults(msgs, 100_000, 70_000)

	// User/assistant messages should be untouched.
	if result[0].Role != fantasy.MessageRoleUser {
		t.Error("first message should remain user")
	}
	if result[1].Role != fantasy.MessageRoleAssistant {
		t.Error("second message should remain assistant")
	}
}

func TestClearToolResultsAfterIdleGap_NoGap(t *testing.T) {
	t.Parallel()
	msgs := []fantasy.Message{
		makeToolResultMessage("t1", "data1"),
	}
	// Recent timestamp — no clearing.
	result := clearToolResultsAfterIdleGap(msgs, time.Now().Unix())
	tr, _ := fantasy.AsMessagePart[fantasy.ToolResultPart](result[0].Content[0])
	text, _ := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](tr.Output)
	if text.Text == clearedToolResultText {
		t.Error("should not clear when gap is short")
	}
}

func TestClearToolResultsAfterIdleGap_LargeGap(t *testing.T) {
	t.Parallel()
	msgs := []fantasy.Message{makeUserMessage("hello")}
	for range 10 {
		msgs = append(msgs, makeToolResultMessage("t", "data"))
	}
	// Timestamp from 2 hours ago.
	oldTime := time.Now().Add(-2 * time.Hour).Unix()
	result := clearToolResultsAfterIdleGap(msgs, oldTime)

	// First 5 tool results (indices 1-5) should be cleared.
	// keepRecentToolResultsAfterIdle = 5, so last 5 kept.
	clearedCount := 0
	for i := 1; i <= len(result)-1; i++ {
		if result[i].Role != fantasy.MessageRoleTool {
			continue
		}
		tr, ok := fantasy.AsMessagePart[fantasy.ToolResultPart](result[i].Content[0])
		if !ok {
			continue
		}
		text, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](tr.Output)
		if !ok {
			continue
		}
		if text.Text == clearedToolResultText {
			clearedCount++
		}
	}
	if clearedCount != 5 {
		t.Errorf("expected 5 cleared results after idle gap, got %d", clearedCount)
	}
}

func TestClearToolResultsAfterIdleGap_ZeroTimestamp(t *testing.T) {
	t.Parallel()
	msgs := []fantasy.Message{
		makeToolResultMessage("t1", "data1"),
	}
	result := clearToolResultsAfterIdleGap(msgs, 0)
	tr, _ := fantasy.AsMessagePart[fantasy.ToolResultPart](result[0].Content[0])
	text, _ := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](tr.Output)
	if text.Text == clearedToolResultText {
		t.Error("should not clear when timestamp is zero")
	}
}
