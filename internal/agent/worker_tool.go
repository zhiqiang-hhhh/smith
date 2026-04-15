package agent

import (
	"context"
	_ "embed"
	"errors"

	"charm.land/fantasy"

	"github.com/zhiqiang-hhhh/smith/internal/agent/prompt"
	"github.com/zhiqiang-hhhh/smith/internal/agent/tools"
	"github.com/zhiqiang-hhhh/smith/internal/config"
)

//go:embed templates/worker_tool.md
var workerToolDescription []byte

// WorkerParams defines the input parameters for the worker tool.
type WorkerParams struct {
	Prompt string `json:"prompt" description:"A detailed, self-contained task description for the worker to execute autonomously"`
}

const (
	WorkerToolName = "worker"
)

func (c *coordinator) workerTool(ctx context.Context) (fantasy.AgentTool, error) {
	agentCfg, ok := c.cfg.Config().Agents[config.AgentWorker]
	if !ok {
		return nil, errors.New("worker agent not configured")
	}
	p, err := workerPrompt(prompt.WithWorkingDir(c.cfg.WorkingDir()))
	if err != nil {
		return nil, err
	}

	agent, err := c.buildAgent(ctx, p, agentCfg, true)
	if err != nil {
		return nil, err
	}
	return fantasy.NewParallelAgentTool(
		WorkerToolName,
		string(workerToolDescription),
		func(ctx context.Context, params WorkerParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Prompt == "" {
				return fantasy.NewTextErrorResponse("prompt is required"), nil
			}

			sessionID := tools.GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, errors.New("session id missing from context")
			}

			agentMessageID := tools.GetMessageFromContext(ctx)
			if agentMessageID == "" {
				return fantasy.ToolResponse{}, errors.New("agent message id missing from context")
			}

			return c.runSubAgent(ctx, subAgentParams{
				Agent:          agent,
				SessionID:      sessionID,
				AgentMessageID: agentMessageID,
				ToolCallID:     call.ID,
				Prompt:         params.Prompt,
				SessionTitle:   "Worker Session",
				SessionSetup: func(sid string) {
					c.permissions.AutoApproveSession(sid)
				},
			})
		}), nil
}
