package model

import "time"

type Profile struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	City      string    `json:"city"`
	CreatedAt time.Time `json:"created_at"`
}
