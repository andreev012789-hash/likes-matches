package repository

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/andreev012789-hash/likes-matches/internal/event"
	"github.com/google/uuid"
)

type EventRepository struct {
	conn driver.Conn
}

func NewEventRepository(ctx context.Context, addr, user, password string) (*EventRepository, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: "likes",
			Username: user,
			Password: password,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}

	return &EventRepository{conn: conn}, nil
}

func (r *EventRepository) Close() error {
	return r.conn.Close()
}

func (r *EventRepository) InsertBatch(ctx context.Context, events []event.Event) error {
	if len(events) == 0 {
		return nil
	}

	batch, err := r.conn.PrepareBatch(ctx, "INSERT INTO events (event_id, type, from_id, to_id, created_at)")
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}
	defer func() { _ = batch.Abort() }()

	for _, e := range events {
		id, err := uuid.Parse(e.ID)
		if err != nil {
			return fmt.Errorf("parse event id %q: %w", e.ID, err)
		}

		if err := batch.Append(id, e.Type, e.FromID, e.ToID, e.CreatedAt); err != nil {
			return fmt.Errorf("append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("send batch: %w", err)
	}

	return nil
}
