package dialog

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/catwalk/pkg/catwalk"
	"github.com/zhiqiang-hhhh/smith/internal/config"
	openaiauth "github.com/zhiqiang-hhhh/smith/internal/oauth/openai"
	"github.com/zhiqiang-hhhh/smith/internal/ui/common"
)

func NewOAuthOpenAI(
	com *common.Common,
	isOnboarding bool,
	provider catwalk.Provider,
	model config.SelectedModel,
	modelType config.SelectedModelType,
) (*OAuth, tea.Cmd) {
	return newOAuth(com, isOnboarding, provider, model, modelType, &OAuthOpenAI{})
}

type OAuthOpenAI struct {
	deviceCode *openaiauth.DeviceCode
	cancelFunc func()
}

var _ OAuthProvider = (*OAuthOpenAI)(nil)

func (m *OAuthOpenAI) name() string {
	return "OpenAI"
}

func (m *OAuthOpenAI) initiateAuth() tea.Msg {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	deviceCode, err := openaiauth.RequestDeviceCode(ctx)
	if err != nil {
		return ActionOAuthErrored{Error: fmt.Errorf("failed to initiate device auth: %w", err)}
	}

	m.deviceCode = deviceCode

	return ActionInitiateOAuth{
		DeviceCode:      deviceCode.DeviceAuthID,
		UserCode:        deviceCode.UserCode,
		VerificationURL: deviceCode.VerificationURL,
		ExpiresIn:       int(openaiauth.MaxPollTimeout.Seconds()),
		Interval:        deviceCode.Interval,
	}
}

func (m *OAuthOpenAI) startPolling(_ string, _ int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		m.cancelFunc = cancel

		token, err := openaiauth.PollForToken(ctx, m.deviceCode)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return ActionOAuthErrored{Error: err}
		}

		return ActionCompleteOAuth{Token: token}
	}
}

func (m *OAuthOpenAI) stopPolling() tea.Msg {
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
	return nil
}
