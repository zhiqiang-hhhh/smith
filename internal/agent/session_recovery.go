package agent

import "github.com/charmbracelet/crush/internal/message"

// wasInterrupted checks if the last assistant message in the conversation was
// interrupted (crashed or killed mid-response). This is detected by looking
// for the synthetic error finish that RepairUnfinished adds. Returns false if
// any user message was sent after the interrupted assistant message (meaning
// the session was already continued).
func wasInterrupted(msgs []message.Message) bool {
	for i := len(msgs) - 1; i >= 0; i-- {
		switch msgs[i].Role {
		case message.User:
			return false
		case message.Assistant:
			fp := msgs[i].FinishPart()
			if fp == nil {
				return false
			}
			return fp.Reason == message.FinishReasonError && fp.Message == "Interrupted"
		}
	}
	return false
}
