package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andreev012789-hash/likes-matches/internal/model"
)

var ErrSelfLike = errors.New("cannot like yourself")

type LikeRepositore struct {
	pool *pgxpool.Pool
}

func NewLikeRepository(pool *pgxpool.Pool) *LikeRepositore {
	return &LikeRepositore{pool: pool}
}

func (r *LikeRepositore) Like(ctx context.Context, FromID, ToID int64) (model.LikeResult, error) {
	if FromID == ToID {
		return model.LikeResult{}, ErrSelfLike
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.LikeResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `INSERT INTO likes (from_id, to_id) VALUES ($1, $2) ON CONFLICT (from_id, to_id) DO NOTHING`, FromID, ToID)
	if err != nil {
		return model.LikeResult{}, fmt.Errorf("insert like: %w", err)
	}

	var mutual bool
	err = tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM likes WHERE from_id = $1 AND to_id = $2)`, ToID, FromID).Scan(&mutual)
	if err != nil {
		return model.LikeResult{}, fmt.Errorf("check mutual like: %w", err)
	}

	var result model.LikeResult
	if mutual {
		a, b := min(FromID, ToID), max(FromID, ToID)

		_, err = tx.Exec(ctx, `INSERT INTO matches (user_a, user_b) VALUES ($1, $2) ON CONFLICT (user_a, user_b) DO NOTHING`, a, b)
		if err != nil {
			return model.LikeResult{}, fmt.Errorf("insert match: %w", err)
		}

		var createdAt time.Time
		err = tx.QueryRow(ctx, `SELECT created_at FROM matches WHERE user_a = $1 AND user_b = $2`, a, b).Scan(&createdAt)

		result = model.LikeResult{
			Matched: true,
			Match:   &model.Match{UserA: a, UserB: b, CreatedAt: createdAt},
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.LikeResult{}, fmt.Errorf("commit: %w", err)
	}
	return result, nil
}
