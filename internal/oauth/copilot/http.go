package copilot

const (
	userAgent           = "GitHubCopilotChat/0.26.7"
	editorVersion       = "vscode/1.104.3"
	editorPluginVersion = "copilot-chat/0.26.7"
	integrationID       = "vscode-chat"
	apiVersion          = "2025-04-01"
)

func Headers() map[string]string {
	return map[string]string{
		"User-Agent":             userAgent,
		"Editor-Version":         editorVersion,
		"Editor-Plugin-Version":  editorPluginVersion,
		"Copilot-Integration-Id": integrationID,
		"X-Github-Api-Version":   apiVersion,
	}
}
