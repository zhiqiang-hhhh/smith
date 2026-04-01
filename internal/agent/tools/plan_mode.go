package tools

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/askuser"
)

const PlanModeToolName = "plan_mode"

//go:embed plan_mode.md
var planModeDescription []byte

type PlanModeParams struct {
	Mode string `json:"mode" description:"Either 'plan' to enter plan mode or 'implement' to exit plan mode" enum:"plan,implement"`
	Plan string `json:"plan,omitempty" description:"When exiting plan mode, include the finalized plan for user approval"`
}

type PlanModeResponseMetadata struct {
	Mode         string `json:"mode"`
	PlanActive   bool   `json:"plan_active"`
	Plan         string `json:"plan,omitempty"`
	ClearContext bool   `json:"clear_context,omitempty"`
}

func NewPlanModeTool(askSvc askuser.Service) fantasy.AgentTool {
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
					fantasy.NewTextResponse("Plan mode activated. You MUST NOT make any edits or run non-readonly tools. Focus on exploring the codebase, designing a plan, and presenting it for approval."),
					metadata,
				), nil
			}

			// mode == "implement": present the plan for user approval.
			sessionID := GetSessionFromContext(ctx)
			plan := strings.TrimSpace(params.Plan)

			question := "Exit plan mode and begin implementation?"
			if plan != "" {
				question = "Ready to implement this plan?"
			}

			options := []askuser.Option{
				{Label: "Approve", Description: "Exit plan mode and start implementing"},
				{Label: "Approve and clear context", Description: "Clear conversation history and start fresh with the plan"},
				{Label: "Reject", Description: "Stay in plan mode and revise the plan"},
			}
			if plan == "" {
				options = []askuser.Option{
					{Label: "Approve", Description: "Exit plan mode"},
					{Label: "Reject", Description: "Stay in plan mode"},
				}
			}

			req := askuser.QuestionRequest{
				SessionID:  sessionID,
				ToolCallID: call.ID,
				ToolName:   PlanModeToolName,
				Question:   question,
				Header:     "Plan Mode",
				Body:       plan,
				Options:    options,
				AllowText:  true,
			}

			answers, err := askSvc.Ask(ctx, req)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to get user confirmation: %s", err)), nil
			}

			if len(answers) == 0 {
				metadata.PlanActive = true
				metadata.Mode = "plan"
				return fantasy.WithResponseMetadata(
					fantasy.NewTextResponse("User dismissed the dialog. Stay in plan mode."),
					metadata,
				), nil
			}

			answer := answers[0]

			if strings.EqualFold(answer, "Reject") {
				metadata.PlanActive = true
				metadata.Mode = "plan"
				response := "User rejected the plan."
				if len(answers) > 1 {
					response += fmt.Sprintf(" User feedback: %s", strings.Join(answers[1:], " "))
				}
				response += " Stay in plan mode and revise your plan."
				return fantasy.WithResponseMetadata(
					fantasy.NewTextResponse(response),
					metadata,
				), nil
			}

			// User approved (either "Approve" or "Approve and clear context").
			metadata.PlanActive = false
			metadata.Plan = plan
			metadata.ClearContext = strings.EqualFold(answer, "Approve and clear context")

			var response string
			if plan != "" {
				response = fmt.Sprintf("User has approved your plan. You can now start coding.\n\n## Approved Plan:\n%s", plan)
			} else {
				response = "User has approved exiting plan mode. You can now proceed."
			}

			return fantasy.WithResponseMetadata(
				fantasy.NewTextResponse(response),
				metadata,
			), nil
		})
}
