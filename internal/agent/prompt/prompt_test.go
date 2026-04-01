package prompt

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/stretchr/testify/require"
)

func TestPromptData_LoadsGlobalContextFiles(t *testing.T) {

	globalDir := t.TempDir()
	workingDir := t.TempDir()
	t.Setenv("CRUSH_GLOBAL_CONFIG", globalDir)

	os.WriteFile(filepath.Join(globalDir, "AGENTS.md"), []byte("global instructions"), 0o644)

	cfg := &config.Config{Options: &config.Options{
		DataDirectory: filepath.Join(workingDir, ".crush"),
	}}
	store := config.NewConfigStoreForTesting(cfg, workingDir)
	p, err := NewPrompt("test", "{{range .ContextFiles}}[{{.Path}}:{{.Content}}]{{end}}")
	require.NoError(t, err)

	data, err := p.promptData(context.Background(), "test", "test-model", store)
	require.NoError(t, err)

	found := false
	for _, cf := range data.ContextFiles {
		if strings.Contains(cf.Content, "global instructions") {
			found = true
			break
		}
	}
	require.True(t, found, "global AGENTS.md should be loaded")
}

func TestPromptData_ProjectOverridesGlobal(t *testing.T) {

	globalDir := t.TempDir()
	workingDir := t.TempDir()
	t.Setenv("CRUSH_GLOBAL_CONFIG", globalDir)

	os.WriteFile(filepath.Join(globalDir, "AGENTS.md"), []byte("global instructions"), 0o644)
	os.WriteFile(filepath.Join(workingDir, "AGENTS.md"), []byte("project instructions"), 0o644)

	cfg := &config.Config{Options: &config.Options{
		DataDirectory: filepath.Join(workingDir, ".crush"),
	}}
	store := config.NewConfigStoreForTesting(cfg, workingDir)
	p, err := NewPrompt("test", "{{range .ContextFiles}}[{{.Content}}]{{end}}")
	require.NoError(t, err)

	data, err := p.promptData(context.Background(), "test", "test-model", store)
	require.NoError(t, err)

	globalFound := false
	projectFound := false
	for _, cf := range data.ContextFiles {
		if strings.Contains(cf.Content, "global instructions") {
			globalFound = true
		}
		if strings.Contains(cf.Content, "project instructions") {
			projectFound = true
		}
	}
	require.True(t, projectFound, "project AGENTS.md should be loaded")
	require.False(t, globalFound, "global AGENTS.md should be overridden by project")
}

func TestPromptData_GlobalDirNotExist(t *testing.T) {

	t.Setenv("CRUSH_GLOBAL_CONFIG", "/nonexistent/path/to/config")
	workingDir := t.TempDir()

	cfg := &config.Config{Options: &config.Options{
		DataDirectory: filepath.Join(workingDir, ".crush"),
	}}
	store := config.NewConfigStoreForTesting(cfg, workingDir)
	p, err := NewPrompt("test", "{{range .ContextFiles}}[{{.Content}}]{{end}}")
	require.NoError(t, err)

	data, err := p.promptData(context.Background(), "test", "test-model", store)
	require.NoError(t, err)
	require.Empty(t, data.ContextFiles)
}
