package agent

import (
	"testing"

	"github.com/zhiqiang-hhhh/smith/internal/message"
)

func TestWasInterrupted_NoMessages(t *testing.T) {
	t.Parallel()
	if wasInterrupted(nil) {
		t.Error("should return false for empty messages")
	}
}

func TestWasInterrupted_NormalFinish(t *testing.T) {
	t.Parallel()
	msgs := []message.Message{
		{Role: message.User, Parts: []message.ContentPart{message.TextContent{Text: "hello"}}},
		{Role: message.Assistant, Parts: []message.ContentPart{
			message.TextContent{Text: "response"},
			message.Finish{Reason: message.FinishReasonEndTurn},
		}},
	}
	if wasInterrupted(msgs) {
		t.Error("normal finish should not be detected as interrupted")
	}
}

func TestWasInterrupted_Interrupted(t *testing.T) {
	t.Parallel()
	msgs := []message.Message{
		{Role: message.User, Parts: []message.ContentPart{message.TextContent{Text: "hello"}}},
		{Role: message.Assistant, Parts: []message.ContentPart{
			message.TextContent{Text: "partial"},
			message.Finish{Reason: message.FinishReasonError, Message: "Interrupted"},
		}},
	}
	if !wasInterrupted(msgs) {
		t.Error("interrupted session should be detected")
	}
}

func TestWasInterrupted_ErrorButNotInterrupted(t *testing.T) {
	t.Parallel()
	msgs := []message.Message{
		{Role: message.Assistant, Parts: []message.ContentPart{
			message.Finish{Reason: message.FinishReasonError, Message: "API error"},
		}},
	}
	if wasInterrupted(msgs) {
		t.Error("non-interrupted errors should not be detected as interrupted")
	}
}

func TestWasInterrupted_LastMessageIsUser(t *testing.T) {
	t.Parallel()
	msgs := []message.Message{
		{Role: message.Assistant, Parts: []message.ContentPart{
			message.Finish{Reason: message.FinishReasonError, Message: "Interrupted"},
		}},
		{Role: message.User, Parts: []message.ContentPart{message.TextContent{Text: "new prompt"}}},
	}
	if wasInterrupted(msgs) {
		t.Error("user message after interrupted assistant means the session continued")
	}
}
