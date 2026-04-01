package agent

import (
	"testing"
)

func TestSummarizeCircuitBreaker_InitialState(t *testing.T) {
	t.Parallel()
	cb := newSummarizeCircuitBreaker()
	if cb.isTripped("session-1") {
		t.Error("new circuit breaker should not be tripped")
	}
}

func TestSummarizeCircuitBreaker_TripsAfterMaxFailures(t *testing.T) {
	t.Parallel()
	cb := newSummarizeCircuitBreaker()

	for range maxConsecutiveSummarizeFailures - 1 {
		cb.recordFailure("session-1")
		if cb.isTripped("session-1") {
			t.Fatal("should not trip before max failures")
		}
	}

	cb.recordFailure("session-1")
	if !cb.isTripped("session-1") {
		t.Error("should trip after max consecutive failures")
	}
}

func TestSummarizeCircuitBreaker_SuccessResetsCount(t *testing.T) {
	t.Parallel()
	cb := newSummarizeCircuitBreaker()

	cb.recordFailure("session-1")
	cb.recordFailure("session-1")
	cb.recordSuccess("session-1")

	if cb.isTripped("session-1") {
		t.Error("success should reset failure count")
	}

	// Should need full maxConsecutiveSummarizeFailures failures again.
	for range maxConsecutiveSummarizeFailures {
		cb.recordFailure("session-1")
	}
	if !cb.isTripped("session-1") {
		t.Error("should trip again after max failures post-reset")
	}
}

func TestSummarizeCircuitBreaker_IsolatesSessions(t *testing.T) {
	t.Parallel()
	cb := newSummarizeCircuitBreaker()

	for range maxConsecutiveSummarizeFailures {
		cb.recordFailure("session-1")
	}

	if cb.isTripped("session-2") {
		t.Error("session-2 should not be affected by session-1 failures")
	}
	if !cb.isTripped("session-1") {
		t.Error("session-1 should be tripped")
	}
}
