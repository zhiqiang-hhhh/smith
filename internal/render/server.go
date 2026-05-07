package render

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	MaxDiagramContentLen = 200 * 1024
	defaultExpireAfter   = time.Hour
)

var (
	ErrUnsupportedFormat = errors.New("unsupported format: only mermaid is supported")
	ErrEmptySessionID    = errors.New("session ID is required")
	ErrEmptyContent      = errors.New("content cannot be empty")
	ErrInvalidUTF8       = errors.New("content and title must be valid UTF-8")
	ErrContentTooLarge   = fmt.Errorf("content exceeds maximum size of %d bytes", MaxDiagramContentLen)
)

type RenderResult struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	SessionID string    `json:"session_id"`
	Format    string    `json:"format"`
	Title     string    `json:"title"`
	ExpiresAt time.Time `json:"expires_at"`
}

type document struct {
	sessionID string
	format    string
	title     string
	content   string
	expiresAt time.Time
}

type Server struct {
	httpServer *http.Server
	listener   net.Listener
	baseURL    string

	mu   sync.RWMutex
	docs map[string]document
}

func NewServer() (*Server, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen render server: %w", err)
	}

	s := &Server{
		listener: ln,
		baseURL:  "http://" + ln.Addr().String(),
		docs:     make(map[string]document),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /render/{id}", s.handleRender)

	s.httpServer = &http.Server{Handler: mux}

	go func() {
		_ = s.httpServer.Serve(ln)
	}()

	return s, nil
}

func (s *Server) BaseURL() string {
	return s.baseURL
}

func (s *Server) URLFor(id string) string {
	return s.baseURL + "/render/" + id
}

func (s *Server) Render(sessionID, format, title, content string, expireAfter time.Duration) (RenderResult, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return RenderResult{}, ErrEmptySessionID
	}

	format = strings.ToLower(strings.TrimSpace(format))
	if format != "mermaid" {
		return RenderResult{}, ErrUnsupportedFormat
	}

	if !utf8.ValidString(content) || !utf8.ValidString(title) {
		return RenderResult{}, ErrInvalidUTF8
	}

	if strings.TrimSpace(content) == "" {
		return RenderResult{}, ErrEmptyContent
	}

	if len(content) > MaxDiagramContentLen {
		return RenderResult{}, ErrContentTooLarge
	}

	if strings.TrimSpace(title) == "" {
		title = "Mermaid Diagram"
	}

	if expireAfter <= 0 {
		expireAfter = defaultExpireAfter
	}

	id, err := generateID()
	if err != nil {
		return RenderResult{}, err
	}

	now := time.Now()
	expiresAt := now.Add(expireAfter)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupExpiredLocked(now)
	s.docs[id] = document{
		sessionID: sessionID,
		format:    format,
		title:     title,
		content:   content,
		expiresAt: expiresAt,
	}

	return RenderResult{
		ID:        id,
		URL:       s.URLFor(id),
		SessionID: sessionID,
		Format:    format,
		Title:     title,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	setSecurityHeaders(w)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleRender(w http.ResponseWriter, r *http.Request) {
	setSecurityHeaders(w)

	id := r.PathValue("id")
	if strings.TrimSpace(id) == "" {
		http.NotFound(w, r)
		return
	}

	now := time.Now()
	s.mu.Lock()
	doc, ok := s.docs[id]
	if ok && !now.Before(doc.expiresAt) {
		delete(s.docs, id)
		ok = false
	}
	s.cleanupExpiredLocked(now)
	s.mu.Unlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := renderPageTemplate.Execute(w, map[string]string{
		"Title":   doc.title,
		"Content": doc.content,
	}); err != nil {
		http.Error(w, "failed to render document", http.StatusInternalServerError)
		return
	}
}

func (s *Server) cleanupExpiredLocked(now time.Time) {
	for id, doc := range s.docs {
		if !now.Before(doc.expiresAt) {
			delete(s.docs, id)
		}
	}
}

func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate render ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src https://cdn.jsdelivr.net 'unsafe-inline'; style-src 'unsafe-inline'; img-src data:; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
}

var renderPageTemplate = template.Must(template.New("render").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <title>{{.Title}}</title>
  <style>
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      margin: 0;
      padding: 24px;
      background: #ffffff;
      color: #111827;
    }
    h1 {
      margin: 0 0 16px;
      font-size: 20px;
    }
    .mermaid {
      border: 1px solid #e5e7eb;
      border-radius: 8px;
      padding: 16px;
      background: #fff;
      overflow: auto;
    }
  </style>
</head>
<body>
  <h1>{{.Title}}</h1>
  <pre class="mermaid">{{.Content}}</pre>
  <script src="https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"></script>
  <script>
    mermaid.initialize({ startOnLoad: true, securityLevel: 'strict' });
  </script>
</body>
</html>
`))
