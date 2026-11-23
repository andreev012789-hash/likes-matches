package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andreev012789-hash/likes-matches/internal/model"
)

var ErrSelfLike = errors.New("cannot like yourself")

type LikeRepository struct {
	pool *pgxpool.Pool
}

func NewLikeRepository(pool *pgxpool.Pool) *LikeRepository {
	return &LikeRepository{pool: pool}
}

func (r *LikeRepository) Like(ctx context.Context, fromID, toID int64) (model.LikeResult, error) {
	if fromID == toID {
		return model.LikeResult{}, ErrSelfLike
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.LikeResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `INSERT INTO likes (from_id, to_id) VALUES ($1, $2) ON CONFLICT (from_id, to_id) DO NOTHING`, fromID, toID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return model.LikeResult{}, ErrNotFound
		}
		return model.LikeResult{}, fmt.Errorf("insert like: %w", err)
	}

	var mutual bool
	err = tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM likes WHERE from_id = $1 AND to_id = $2)`, toID, fromID).Scan(&mutual)
	if err != nil {
		return model.LikeResult{}, fmt.Errorf("check mutual like: %w", err)
	}

	var result model.LikeResult
	if mutual {
		a, b := min(fromID, toID), max(fromID, toID)

		_, err = tx.Exec(ctx, `INSERT INTO matches (user_a, user_b) VALUES ($1, $2) ON CONFLICT (user_a, user_b) DO NOTHING`, a, b)
		if err != nil {
			return model.LikeResult{}, fmt.Errorf("insert match: %w", err)
		}

		var createdAt time.Time
		err = tx.QueryRow(ctx, `SELECT created_at FROM matches WHERE user_a = $1 AND user_b = $2`, a, b).Scan(&createdAt)
		if err != nil {
			return model.LikeResult{}, fmt.Errorf("select match: %w", err)
		}

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

func (r *LikeRepository) Matches(ctx context.Context, userID int64) ([]model.Profile, error) {
	rows, err := r.pool.Query(ctx, `SELECT p.id, p.name, p.age, p.city, p.created_at
		FROM matches m
		JOIN profiles p ON p.id = CASE WHEN m.user_a = $1 THEN m.user_b ELSE m.user_a END
		WHERE m.user_a = $1 OR m.user_b = $1
		ORDER BY m.created_at DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("query matches: %w", err)
	}
	defer rows.Close()

	profiles := []model.Profile{}
	for rows.Next() {
		var p model.Profile
		if err := rows.Scan(&p.ID, &p.Name, &p.Age, &p.City, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan match: %w", err)
		}
		profiles = append(profiles, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate matches: %w", err)
	}
	return profiles, nil
}
