package trace

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zhiqiang-hhhh/smith/internal/db"
)

type Record struct {
	ID         string
	SessionID  string
	StartedAt  int64
	StoppedAt  int64
	EventCount int64
	DataJSONL  string
	CreatedAt  int64
}

type Service interface {
	Save(ctx context.Context, sessionID string, snapshot Snapshot) (Record, error)
	Get(ctx context.Context, traceID string) (Record, error)
	ListBySession(ctx context.Context, sessionID string) ([]Record, error)
}

type service struct {
	q db.Querier
}

func NewService(q db.Querier) Service {
	return &service{q: q}
}

func (s *service) Save(ctx context.Context, sessionID string, snapshot Snapshot) (Record, error) {
	traceID := "trc_" + uuid.NewString()
	dbTrace, err := s.q.CreateTrace(ctx, db.CreateTraceParams{
		ID:         traceID,
		SessionID:  sessionID,
		StartedAt:  snapshot.StartedAt,
		StoppedAt:  snapshot.StoppedAt,
		EventCount: int64(snapshot.EventCount),
		DataJsonl:  snapshot.DataJSONL,
	})
	if err != nil {
		return Record{}, err
	}
	return fromDBRecord(dbTrace), nil
}

func (s *service) Get(ctx context.Context, traceID string) (Record, error) {
	dbTrace, err := s.q.GetTraceByID(ctx, traceID)
	if err != nil {
		return Record{}, err
	}
	return fromDBRecord(dbTrace), nil
}

func (s *service) ListBySession(ctx context.Context, sessionID string) ([]Record, error) {
	dbTraces, err := s.q.ListTracesBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	records := make([]Record, len(dbTraces))
	for i, dbTrace := range dbTraces {
		records[i] = fromDBRecord(dbTrace)
	}
	return records, nil
}

func fromDBRecord(t db.Trace) Record {
	return Record{
		ID:         t.ID,
		SessionID:  t.SessionID,
		StartedAt:  t.StartedAt,
		StoppedAt:  t.StoppedAt,
		EventCount: t.EventCount,
		DataJSONL:  t.DataJsonl,
		CreatedAt:  t.CreatedAt,
	}
}

func FormatAttachment(record Record) (string, []byte, error) {
	if record.ID == "" {
		return "", nil, fmt.Errorf("trace id is empty")
	}
	return fmt.Sprintf("%s.jsonl", record.ID), []byte(record.DataJSONL), nil
}
