package tools

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/askuser"
)

const AskUserToolName = "ask_user"

//go:embed ask_user.md
var askUserDescription []byte

type AskUserOption struct {
	Label       string `json:"label" description:"Display text (1-5 words, concise)"`
	Description string `json:"description,omitempty" description:"Explanation of the choice"`
}

type AskUserParams struct {
	Question  string         `json:"question" description:"The full question to ask the user. Be specific and provide context."`
	Header    string         `json:"header,omitempty" description:"Short label for the dialog title (max 30 chars)"`
	Options   []AskUserOption `json:"options,omitempty" description:"Available choices. If empty, a text input is shown."`
	Multi     bool           `json:"multi,omitempty" description:"Allow selecting multiple choices (default: false)"`
	AllowText bool           `json:"allow_text,omitempty" description:"Allow typing a custom answer even when options are provided (default: true when options are provided)"`
}

type AskUserResponseMetadata struct {
	Question string   `json:"question"`
	Answers  []string `json:"answers"`
}

func NewAskUserTool(svc askuser.Service) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		AskUserToolName,
		string(askUserDescription),
		func(ctx context.Context, params AskUserParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if strings.TrimSpace(params.Question) == "" {
				return fantasy.NewTextErrorResponse("question is required"), nil
			}

			sessionID := GetSessionFromContext(ctx)

			options := make([]askuser.Option, len(params.Options))
			for i, o := range params.Options {
				options[i] = askuser.Option{
					Label:       o.Label,
					Description: o.Description,
				}
			}

			allowText := params.AllowText
			if len(params.Options) > 0 && !params.AllowText {
				allowText = true
			}

			req := askuser.QuestionRequest{
				SessionID:  sessionID,
				ToolCallID: call.ID,
				Question:   params.Question,
				Header:     params.Header,
				Options:    options,
				Multi:      params.Multi,
				AllowText:  allowText,
			}

			answers, err := svc.Ask(ctx, req)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to get user response: %s", err)), nil
			}

			if len(answers) == 0 {
				return fantasy.NewTextResponse("User dismissed the question without answering."), nil
			}

			metadata := AskUserResponseMetadata{
				Question: params.Question,
				Answers:  answers,
			}

			display := askuser.MarshalAnswers(answers)
			return fantasy.WithResponseMetadata(
				fantasy.NewTextResponse(fmt.Sprintf("User answered: %s", display)),
				metadata,
			), nil
		})
}
