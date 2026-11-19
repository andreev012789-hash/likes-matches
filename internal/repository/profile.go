package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andreev012789-hash/likes-matches/internal/model"
)

var ErrNotFound = errors.New("Not found")

type ProfileRepository struct {
	pool *pgxpool.Pool
}

func NewProfileRepositore(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

func (r *ProfileRepository) CreateProfile(ctx context.Context, name string, age int, city string) (model.Profile, error) {
	var p model.Profile
	err := r.pool.QueryRow(ctx,
		`INSERT INTO profiles (name, age, city)
	 	VALUES ($1, $2, $3)
		RETURNING id, name, age, city, created_at`,
		name, age, city,
	).Scan(&p.ID, &p.Name, &p.Age, &p.City, &p.CreatedAt)
	if err != nil {
		return model.Profile{}, fmt.Errorf("insert profile: %w", err)
	}
	return p, nil
}

func (r *ProfileRepository) GetProfile(ctx context.Context, id int64) (model.Profile, error) {
	var p model.Profile
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, age, city, created_at
		 FROM profiles
		 WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.Name, &p.Age, &p.City, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Profile{}, ErrNotFound
	}
	if err != nil {
		return model.Profile{}, fmt.Errorf("select profile: %w", err)
	}
	return p, nil
}
