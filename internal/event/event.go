package event

import (
	"time"

	"github.com/google/uuid"
)

const (
	TypeLikeCreated  = "like.created"
	TypeMatchCreated = "match.created"
)

type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	FromID    int64     `json:"from_id"`
	ToID      int64     `json:"to_id"`
	CreatedAt time.Time `json:"created_at"`
}

func New(eventType string, fromID, toID int64) Event {
	return Event{
		ID:        uuid.NewString(),
		Type:      eventType,
		FromID:    fromID,
		ToID:      toID,
		CreatedAt: time.Now().UTC(),
	}
}
