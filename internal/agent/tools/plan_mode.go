package tools

import (
	"context"
	_ "embed"
	"strings"

	"charm.land/fantasy"
)

const PlanModeToolName = "plan_mode"

//go:embed plan_mode.md
var planModeDescription []byte

type PlanModeParams struct {
	Mode string `json:"mode" description:"Either 'plan' to enter plan mode or 'implement' to exit plan mode" enum:"plan,implement"`
	Plan string `json:"plan,omitempty" description:"When exiting plan mode, the finalized plan that was approved"`
}

type PlanModeResponseMetadata struct {
	Mode       string `json:"mode"`
	PlanActive bool   `json:"plan_active"`
	Plan       string `json:"plan,omitempty"`
}

func NewPlanModeTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		PlanModeToolName,
		string(planModeDescription),
		func(ctx context.Context, params PlanModeParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			mode := strings.TrimSpace(strings.ToLower(params.Mode))
			if mode != "plan" && mode != "implement" {
				return fantasy.NewTextErrorResponse("mode must be 'plan' or 'implement'"), nil
			}

			metadata := PlanModeResponseMetadata{
				Mode: mode,
			}

			if mode == "plan" {
				metadata.PlanActive = true
				return fantasy.WithResponseMetadata(
					fantasy.NewTextResponse("Plan mode activated. Use ONLY read-only tools (view, glob, grep, ls, agent, sourcegraph, web_search, fetch, diff) to explore the codebase and formulate your plan. Present the plan to the user before switching to implement mode."),
					metadata,
				), nil
			}

			// mode == "implement"
			metadata.PlanActive = false
			metadata.Plan = params.Plan

			response := "Implementation mode activated."
			if params.Plan != "" {
				response += " Proceeding with the approved plan."
			}
			response += " You may now use all available tools."

			return fantasy.WithResponseMetadata(
				fantasy.NewTextResponse(response),
				metadata,
			), nil
		})
}
