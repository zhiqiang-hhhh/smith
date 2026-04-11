package tools

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"log/slog"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/shell"
)

const (
	JobOutputToolName = "job_output"

	jobOutputPollInterval = 200 * time.Millisecond
	jobOutputMinWait      = 700 * time.Millisecond
	jobOutputQuietPeriod  = 1200 * time.Millisecond
	jobOutputMaxWait      = 5 * time.Second
	jobOutputWaitRounds   = 3
)

//go:embed job_output.md
var jobOutputDescription []byte

type JobOutputParams struct {
	ShellID string `json:"shell_id" description:"The ID of the background shell to retrieve output from"`
	Wait    bool   `json:"wait" description:"If true, block until the background shell completes before returning output"`
}

type JobOutputResponseMetadata struct {
	ShellID             string `json:"shell_id"`
	Command             string `json:"command"`
	Description         string `json:"description"`
	Done                bool   `json:"done"`
	WorkingDirectory    string `json:"working_directory"`
	LikelyLongRunning   bool   `json:"likely_long_running"`
	ObservedAnyProgress bool   `json:"observed_any_progress"`
}

type waitOutcome int

const (
	waitOutcomeCompleted waitOutcome = iota
	waitOutcomeQuiet
	waitOutcomeMaxWait
	waitOutcomeInterrupted
)

func waitForCompletionOrProgress(ctx context.Context, bgShell *shell.BackgroundShell) (outcome waitOutcome, observedProgress bool) {
	stdout, stderr, done, _ := bgShell.GetOutput()
	if done {
		return waitOutcomeCompleted, false
	}

	start := time.Now()
	lastChangeAt := start
	lastSize := len(stdout) + len(stderr)
	ticker := time.NewTicker(jobOutputPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return waitOutcomeInterrupted, observedProgress
		case <-ticker.C:
			stdout, stderr, done, _ = bgShell.GetOutput()
			if done {
				return waitOutcomeCompleted, observedProgress
			}

			now := time.Now()
			size := len(stdout) + len(stderr)
			if size != lastSize {
				lastSize = size
				lastChangeAt = now
				observedProgress = true
			}

			elapsed := now.Sub(start)
			if elapsed >= jobOutputMaxWait {
				return waitOutcomeMaxWait, observedProgress
			}
			if elapsed >= jobOutputMinWait && now.Sub(lastChangeAt) >= jobOutputQuietPeriod {
				return waitOutcomeQuiet, observedProgress
			}
		}
	}
}

func likelyLongRunningCommand(command string) bool {
	cmd := strings.ToLower(command)
	longRunningHints := []string{
		"server",
		"serve",
		"watch",
		"tail -f",
		"npm start",
		"npm run dev",
		"pnpm dev",
		"yarn dev",
		"python -m http.server",
		"uvicorn",
		"gunicorn",
		"flask run",
		"django",
		"manage.py runserver",
		"java -jar",
		"clickhouse server",
		"docker run",
		"docker compose up",
		"cargo run",
		"go run",
	}
	for _, hint := range longRunningHints {
		if strings.Contains(cmd, hint) {
			return true
		}
	}
	return false
}

func NewJobOutputTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		JobOutputToolName,
		string(jobOutputDescription),
		func(ctx context.Context, params JobOutputParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.ShellID == "" {
				return fantasy.NewTextErrorResponse("missing shell_id"), nil
			}

			bgManager := shell.GetBackgroundShellManager()
			bgShell, ok := bgManager.Get(params.ShellID)
			if !ok {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("background shell not found: %s", params.ShellID)), nil
			}

			observedAnyProgress := false
			lastWaitOutcome := waitOutcomeCompleted
			if params.Wait {
				for range jobOutputWaitRounds {
					waitOutcome, observedProgress := waitForCompletionOrProgress(ctx, bgShell)
					observedAnyProgress = observedAnyProgress || observedProgress
					lastWaitOutcome = waitOutcome
					if waitOutcome == waitOutcomeInterrupted {
						slog.Warn("job_output: wait interrupted by context cancellation",
							"shell_id", params.ShellID,
							"command", bgShell.Command)
						break
					}
					if waitOutcome == waitOutcomeCompleted {
						break
					}
				}
			}

			stdout, stderr, done, err := bgShell.GetOutput()

			var outputParts []string
			if stdout != "" {
				outputParts = append(outputParts, stdout)
			}
			if stderr != "" {
				outputParts = append(outputParts, stderr)
			}

			status := "running"
			if done {
				status = "completed"
				if err != nil {
					exitCode := shell.ExitCode(err)
					if exitCode != 0 {
						outputParts = append(outputParts, fmt.Sprintf("Exit code %d", exitCode))
					}
				}
				bgManager.Remove(params.ShellID)
			}

			output := strings.Join(outputParts, "\n")
			longRunning := likelyLongRunningCommand(bgShell.Command)
			if !done {
				switch {
				case longRunning:
					output += "\n\nAgent hint: likely long-running service/watch task; keep polling with wait=false and stop via job_kill when done."
				case observedAnyProgress:
					output += "\n\nAgent hint: job is still making progress; continue polling until status becomes completed."
				case lastWaitOutcome == waitOutcomeQuiet:
					output += "\n\nAgent hint: job is currently quiet; it may be waiting on a child process or blocked I/O."
				default:
					output += "\n\nAgent hint: job is still running; call job_output with wait=true to wait for completion, or job_kill to stop it."
				}
			}

			metadata := JobOutputResponseMetadata{
				ShellID:             params.ShellID,
				Command:             bgShell.Command,
				Description:         bgShell.Description,
				Done:                done,
				WorkingDirectory:    bgShell.WorkingDir,
				LikelyLongRunning:   longRunning,
				ObservedAnyProgress: observedAnyProgress,
			}

			if output == "" {
				output = BashNoOutput
			}

			result := fmt.Sprintf("Status: %s\n\n%s", status, output)
			result = TruncateString(result, MaxOutputLength)
			return fantasy.WithResponseMetadata(fantasy.NewTextResponse(result), metadata), nil
		})
}
