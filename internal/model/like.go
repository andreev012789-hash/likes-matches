package model

import "time"

type Like struct {
	FromID    int64     `json:"from_id"`
	ToID      int64     `json:"to_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Match struct {
	UserA     int64     `json:"user_a"`
	UserB     int64     `json:"user_b"`
	CreatedAt time.Time `json:"created_at"`
}

type LikeResult struct {
	Matched bool   `json:"matched"`
	Match   *Match `json:"match,omitempty"`
}
