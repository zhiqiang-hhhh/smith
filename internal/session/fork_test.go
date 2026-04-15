package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/zhiqiang-hhhh/smith/internal/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type testEnv struct {
	svc  Service
	q    *db.Queries
	conn *sql.DB
}

func setupForkTest(t *testing.T) *testEnv {
	t.Helper()
	conn, err := db.Connect(t.Context(), t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	q := db.New(conn)
	return &testEnv{
		svc:  NewService(q, conn),
		q:    q,
		conn: conn,
	}
}

func createSession(t *testing.T, ctx context.Context, q *db.Queries, title string) db.Session {
	t.Helper()
	s, err := q.CreateSession(ctx, db.CreateSessionParams{
		ID:               uuid.New().String(),
		Title:            title,
		PromptTokens:     100,
		CompletionTokens: 200,
		Cost:             0.05,
	})
	require.NoError(t, err)
	return s
}

func makeParts(t *testing.T, finishReason string) string {
	t.Helper()
	type partData struct {
		Reason string `json:"reason"`
	}
	type part struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	var parts []part
	parts = append(parts, part{
		Type: "text",
		Data: json.RawMessage(`{"text":"hello"}`),
	})
	if finishReason != "" {
		d, err := json.Marshal(partData{Reason: finishReason})
		require.NoError(t, err)
		parts = append(parts, part{
			Type: "finish",
			Data: d,
		})
	}
	b, err := json.Marshal(parts)
	require.NoError(t, err)
	return string(b)
}

func addMessage(t *testing.T, ctx context.Context, q *db.Queries, sessionID, role, finishReason string) db.Message {
	t.Helper()
	msg, err := q.CreateMessage(ctx, db.CreateMessageParams{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      role,
		Parts:     makeParts(t, finishReason),
	})
	require.NoError(t, err)
	return msg
}

func addSummaryMessage(t *testing.T, ctx context.Context, q *db.Queries, sessionID string) db.Message {
	t.Helper()
	msg, err := q.CreateMessage(ctx, db.CreateMessageParams{
		ID:               uuid.New().String(),
		SessionID:        sessionID,
		Role:             "assistant",
		Parts:            makeParts(t, "end_turn"),
		IsSummaryMessage: 1,
	})
	require.NoError(t, err)
	return msg
}

func TestFork_BasicCopy(t *testing.T) {
	env := setupForkTest(t)
	ctx := t.Context()

	src := createSession(t, ctx, env.q, "Original")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "end_turn")

	forked, err := env.svc.Fork(ctx, src.ID)
	require.NoError(t, err)
	require.NotEqual(t, src.ID, forked.ID)
	require.Contains(t, forked.Title, "Original (fork ")
	require.Equal(t, int64(100), forked.PromptTokens)
	require.Equal(t, int64(200), forked.CompletionTokens)
	require.Equal(t, 0.05, forked.Cost)

	msgs, err := env.q.ListMessagesBySession(ctx, forked.ID)
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	require.Equal(t, "user", msgs[0].Role)
	require.Equal(t, "assistant", msgs[1].Role)
}

func TestFork_TrimsIncompleteLoop(t *testing.T) {
	env := setupForkTest(t)
	ctx := t.Context()

	src := createSession(t, ctx, env.q, "With partial")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "end_turn")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "") // no end_turn = incomplete

	forked, err := env.svc.Fork(ctx, src.ID)
	require.NoError(t, err)

	msgs, err := env.q.ListMessagesBySession(ctx, forked.ID)
	require.NoError(t, err)
	require.Len(t, msgs, 2, "should trim to last complete agent loop")
	require.Equal(t, "user", msgs[0].Role)
	require.Equal(t, "assistant", msgs[1].Role)
}

func TestFork_TrimsErrorFinish(t *testing.T) {
	env := setupForkTest(t)
	ctx := t.Context()

	src := createSession(t, ctx, env.q, "Error finish")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "end_turn")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "error") // error finish

	forked, err := env.svc.Fork(ctx, src.ID)
	require.NoError(t, err)

	msgs, err := env.q.ListMessagesBySession(ctx, forked.ID)
	require.NoError(t, err)
	require.Len(t, msgs, 2, "error-finished loop should be trimmed")
}

func TestFork_NoCompletedLoop_ClearsAll(t *testing.T) {
	env := setupForkTest(t)
	ctx := t.Context()

	src := createSession(t, ctx, env.q, "No complete loop")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "") // never completed

	forked, err := env.svc.Fork(ctx, src.ID)
	require.NoError(t, err)

	msgs, err := env.q.ListMessagesBySession(ctx, forked.ID)
	require.NoError(t, err)
	require.Empty(t, msgs, "no complete loop means empty session")
}

func TestFork_AlreadyCompleteNoTrim(t *testing.T) {
	env := setupForkTest(t)
	ctx := t.Context()

	src := createSession(t, ctx, env.q, "Already complete")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "end_turn")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "end_turn")

	forked, err := env.svc.Fork(ctx, src.ID)
	require.NoError(t, err)

	msgs, err := env.q.ListMessagesBySession(ctx, forked.ID)
	require.NoError(t, err)
	require.Len(t, msgs, 4, "all messages kept when last is end_turn")
}

func TestFork_EmptySession(t *testing.T) {
	env := setupForkTest(t)
	ctx := t.Context()

	src := createSession(t, ctx, env.q, "Empty")

	forked, err := env.svc.Fork(ctx, src.ID)
	require.NoError(t, err)
	require.Contains(t, forked.Title, "Empty (fork ")

	msgs, err := env.q.ListMessagesBySession(ctx, forked.ID)
	require.NoError(t, err)
	require.Empty(t, msgs)
}

func TestFork_CopiesFiles(t *testing.T) {
	env := setupForkTest(t)
	ctx := t.Context()

	src := createSession(t, ctx, env.q, "With files")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "end_turn")

	_, err := env.q.CreateFile(ctx, db.CreateFileParams{
		ID:        uuid.New().String(),
		SessionID: src.ID,
		Path:      "/tmp/test.go",
		Content:   "package main",
		Version:   1,
	})
	require.NoError(t, err)

	forked, err := env.svc.Fork(ctx, src.ID)
	require.NoError(t, err)

	files, err := env.q.ListFilesBySession(ctx, forked.ID)
	require.NoError(t, err)
	require.Len(t, files, 1)
	require.Equal(t, "/tmp/test.go", files[0].Path)
	require.Equal(t, "package main", files[0].Content)
	require.Equal(t, forked.ID, files[0].SessionID)
}

func TestFork_CopiesReadFiles(t *testing.T) {
	env := setupForkTest(t)
	ctx := t.Context()

	src := createSession(t, ctx, env.q, "With read files")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "end_turn")

	err := env.q.RecordFileRead(ctx, db.RecordFileReadParams{
		SessionID: src.ID,
		Path:      "/tmp/read.go",
	})
	require.NoError(t, err)

	forked, err := env.svc.Fork(ctx, src.ID)
	require.NoError(t, err)

	readFiles, err := env.q.ListSessionReadFiles(ctx, forked.ID)
	require.NoError(t, err)
	require.Len(t, readFiles, 1)
	require.Equal(t, "/tmp/read.go", readFiles[0].Path)
}

func TestFork_CopiesSummaryMessageID(t *testing.T) {
	env := setupForkTest(t)
	ctx := t.Context()

	src := createSession(t, ctx, env.q, "With summary")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	summaryMsg := addSummaryMessage(t, ctx, env.q, src.ID)

	_, err := env.q.UpdateSession(ctx, db.UpdateSessionParams{
		ID:               src.ID,
		Title:            src.Title,
		PromptTokens:     src.PromptTokens,
		CompletionTokens: src.CompletionTokens,
		SummaryMessageID: sql.NullString{String: summaryMsg.ID, Valid: true},
		Cost:             src.Cost,
	})
	require.NoError(t, err)

	forked, err := env.svc.Fork(ctx, src.ID)
	require.NoError(t, err)
	require.NotEmpty(t, forked.SummaryMessageID)
	require.NotEqual(t, summaryMsg.ID, forked.SummaryMessageID,
		"forked session should have its own summary message ID (not the original)")
}

func TestFork_DoesNotMutateSource(t *testing.T) {
	env := setupForkTest(t)
	ctx := t.Context()

	src := createSession(t, ctx, env.q, "Source")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "end_turn")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "") // incomplete

	_, err := env.svc.Fork(ctx, src.ID)
	require.NoError(t, err)

	srcMsgs, err := env.q.ListMessagesBySession(ctx, src.ID)
	require.NoError(t, err)
	require.Len(t, srcMsgs, 4, "source session should be untouched")
}

func TestFork_NonExistentSession(t *testing.T) {
	env := setupForkTest(t)
	ctx := t.Context()

	_, err := env.svc.Fork(ctx, "nonexistent-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "getting source session")
}

func TestFork_MultipleForksIndependent(t *testing.T) {
	env := setupForkTest(t)
	ctx := t.Context()

	src := createSession(t, ctx, env.q, "Source")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "end_turn")

	fork1, err := env.svc.Fork(ctx, src.ID)
	require.NoError(t, err)
	fork2, err := env.svc.Fork(ctx, src.ID)
	require.NoError(t, err)

	require.NotEqual(t, fork1.ID, fork2.ID)
	require.NotEqual(t, fork1.ID, src.ID)
}

func TestForkTitle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain title", "My Session", `My Session (fork `},
		{"already forked", "My Session (fork 2026-04-07 12:00:00)", `My Session (fork `},
		{"double forked", "My Session (fork 2026-01-01 00:00:00) (fork 2026-04-07 12:00:00)", `My Session (fork 2026-01-01 00:00:00) (fork `},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := forkTitle(tt.input)
			require.Contains(t, got, tt.want)
			require.NotContains(t, got, "(fork 2026-04-07 12:00:00)")
		})
	}
}

func TestFork_ForkOfForkReplacesTitle(t *testing.T) {
	env := setupForkTest(t)
	ctx := t.Context()

	src := createSession(t, ctx, env.q, "Original")
	addMessage(t, ctx, env.q, src.ID, "user", "")
	addMessage(t, ctx, env.q, src.ID, "assistant", "end_turn")

	fork1, err := env.svc.Fork(ctx, src.ID)
	require.NoError(t, err)
	require.Contains(t, fork1.Title, "Original (fork ")

	fork2, err := env.svc.Fork(ctx, fork1.ID)
	require.NoError(t, err)
	require.Contains(t, fork2.Title, "Original (fork ")
	count := strings.Count(fork2.Title, "(fork ")
	require.Equal(t, 1, count, "title should have only one (fork ...) suffix, got: %s", fork2.Title)
}

func TestIsEndTurnFinish(t *testing.T) {
	require.True(t, isEndTurnFinish(makeParts(t, "end_turn")))
	require.False(t, isEndTurnFinish(makeParts(t, "error")))
	require.False(t, isEndTurnFinish(makeParts(t, "")))
	require.False(t, isEndTurnFinish("invalid json"))
}

func TestIsErrorFinish(t *testing.T) {
	require.True(t, isErrorFinish(makeParts(t, "error")))
	require.False(t, isErrorFinish(makeParts(t, "end_turn")))
	require.False(t, isErrorFinish(makeParts(t, "")))
}
