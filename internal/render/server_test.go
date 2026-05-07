package render

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerRenderAndFetch(t *testing.T) {
	t.Parallel()

	srv, err := NewServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = srv.Shutdown(context.Background())
	})

	healthResp, err := http.Get(srv.BaseURL() + "/healthz")
	require.NoError(t, err)
	defer healthResp.Body.Close()
	require.Equal(t, http.StatusOK, healthResp.StatusCode)

	result, err := srv.Render("session-1", "mermaid", "Flow", "graph TD\nA-->B", time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, result.ID)
	require.Contains(t, result.URL, "/render/")
	require.Equal(t, "session-1", result.SessionID)

	renderResp, err := http.Get(result.URL)
	require.NoError(t, err)
	defer renderResp.Body.Close()
	require.Equal(t, http.StatusOK, renderResp.StatusCode)
	require.Equal(t, "nosniff", renderResp.Header.Get("X-Content-Type-Options"))
	require.Equal(t, "no-referrer", renderResp.Header.Get("Referrer-Policy"))
	require.NotEmpty(t, renderResp.Header.Get("Content-Security-Policy"))
}

func TestServerRenderValidation(t *testing.T) {
	t.Parallel()

	srv, err := NewServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = srv.Shutdown(context.Background())
	})

	_, err = srv.Render("session", "plantuml", "x", "graph TD\nA-->B", time.Minute)
	require.ErrorIs(t, err, ErrUnsupportedFormat)

	_, err = srv.Render("", "mermaid", "x", "graph TD\nA-->B", time.Minute)
	require.ErrorIs(t, err, ErrEmptySessionID)

	_, err = srv.Render("session", "mermaid", "x", "", time.Minute)
	require.ErrorIs(t, err, ErrEmptyContent)

	invalid := string([]byte{0xff, 0xfe, 0xfd})
	_, err = srv.Render("session", "mermaid", invalid, "graph TD\nA-->B", time.Minute)
	require.ErrorIs(t, err, ErrInvalidUTF8)

	tooLarge := make([]byte, MaxDiagramContentLen+1)
	for i := range tooLarge {
		tooLarge[i] = 'a'
	}
	_, err = srv.Render("session", "mermaid", "x", string(tooLarge), time.Minute)
	require.ErrorIs(t, err, ErrContentTooLarge)
}

func TestServerRenderExpiry(t *testing.T) {
	t.Parallel()

	srv, err := NewServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = srv.Shutdown(context.Background())
	})

	result, err := srv.Render("session-1", "mermaid", "Flow", "graph TD\nA-->B", 5*time.Millisecond)
	require.NoError(t, err)

	time.Sleep(30 * time.Millisecond)

	resp, err := http.Get(result.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
