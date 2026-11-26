package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProcessedRepository struct {
	pool *pgxpool.Pool
}

func NewProcessedRepository(pool *pgxpool.Pool) *ProcessedRepository {
	return &ProcessedRepository{pool: pool}
}

func (r *ProcessedRepository) MarkProcessed(ctx context.Context, eventID string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `INSERT INTO processed_events (event_id) VALUES ($1) ON CONFLICT (event_id) DO NOTHING`, eventID)
	if err != nil {
		return false, fmt.Errorf("mark event processed: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}
