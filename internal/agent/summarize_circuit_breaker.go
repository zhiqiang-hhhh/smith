package agent

import (
	"github.com/charmbracelet/crush/internal/csync"
)

// maxConsecutiveSummarizeFailures is the number of consecutive auto-summarize
// failures after which the circuit breaker trips and disables auto-summarize
// for that session.
const maxConsecutiveSummarizeFailures = 3

// summarizeCircuitBreaker tracks consecutive auto-summarize failures per
// session to avoid wasting API calls on sessions where summarization
// repeatedly fails.
type summarizeCircuitBreaker struct {
	failures *csync.Map[string, int]
}

func newSummarizeCircuitBreaker() *summarizeCircuitBreaker {
	return &summarizeCircuitBreaker{
		failures: csync.NewMap[string, int](),
	}
}

// recordFailure increments the consecutive failure count for a session.
func (b *summarizeCircuitBreaker) recordFailure(sessionID string) {
	b.failures.Update(sessionID, func(count int) int {
		return count + 1
	})
}

// recordSuccess resets the failure count for a session.
func (b *summarizeCircuitBreaker) recordSuccess(sessionID string) {
	b.failures.Del(sessionID)
}

// isTripped returns true if the circuit breaker has tripped for a session
// (too many consecutive failures).
func (b *summarizeCircuitBreaker) isTripped(sessionID string) bool {
	count, ok := b.failures.Get(sessionID)
	return ok && count >= maxConsecutiveSummarizeFailures
}
