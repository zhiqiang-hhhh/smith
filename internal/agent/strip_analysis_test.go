package agent

import (
	"testing"
)

func TestStripAnalysisBlock_RemovesBlock(t *testing.T) {
	t.Parallel()
	input := `<analysis>
This is the reasoning scratchpad.
Walking through the conversation...
</analysis>

## Current State
The task is to implement feature X.`

	want := `## Current State
The task is to implement feature X.`

	got := stripAnalysisBlock(input)
	if got != want {
		t.Errorf("stripAnalysisBlock:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestStripAnalysisBlock_NoBlock(t *testing.T) {
	t.Parallel()
	input := "## Current State\nThe task is done."
	got := stripAnalysisBlock(input)
	if got != input {
		t.Errorf("expected unchanged text, got %q", got)
	}
}

func TestStripAnalysisBlock_MultipleBlocks(t *testing.T) {
	t.Parallel()
	input := `<analysis>first</analysis>
text between
<analysis>second</analysis>
final text`

	want := `text between
final text`

	got := stripAnalysisBlock(input)
	if got != want {
		t.Errorf("stripAnalysisBlock:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestStripAnalysisBlock_PreservesKeyFacts(t *testing.T) {
	t.Parallel()
	input := `<analysis>reasoning here</analysis>

## Summary
Done.

<key_facts>
- files_modified: main.go
- current_task: feature X
</key_facts>`

	got := stripAnalysisBlock(input)
	if !keyFactsRegex.MatchString(got) {
		t.Error("key_facts block should be preserved after stripping analysis")
	}
	if analysisBlockRegex.MatchString(got) {
		t.Error("analysis block should have been removed")
	}
}
