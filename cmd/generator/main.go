package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/andreev012789-hash/likes-matches/internal/event"
	"github.com/segmentio/kafka-go"
)

const (
	kafkaAddr   = "localhost:9094"
	topic       = "events"
	totalEvents = 10000
	chunkSize   = 500
	maxUserID   = 50
	spread      = 30 * 24 * time.Hour
)

func main() {
	if err := run(); err != nil {
		slog.Error("generator failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	writer := &kafka.Writer{
		Addr:         kafka.TCP(kafkaAddr),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		BatchTimeout: 10 * time.Millisecond,
	}
	defer func() {
		if err := writer.Close(); err != nil {
			slog.Error("close writer", "error", err)
		}
	}()

	var likes, matches int
	chunk := make([]kafka.Message, 0, chunkSize)

	for i := 0; i < totalEvents; i++ {
		fromID := int64(rand.Intn(maxUserID) + 1)
		toID := int64(rand.Intn(maxUserID-1) + 1)
		if toID >= fromID {
			toID++
		}

		eventType := event.TypeLikeCreated
		if rand.Intn(10) == 0 {
			eventType = event.TypeMatchCreated
			matches++
		} else {
			likes++
		}

		e := event.New(eventType, fromID, toID)
		e.CreatedAt = time.Now().UTC().Add(-time.Duration(rand.Int63n(int64(spread))))

		body, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("marshal event: %w", err)
		}

		chunk = append(chunk, kafka.Message{
			Key:   []byte(strconv.FormatInt(fromID, 10)),
			Value: body,
		})

		if len(chunk) == chunkSize {
			if err := writer.WriteMessages(ctx, chunk...); err != nil {
				return fmt.Errorf("write chunk: %w", err)
			}
			chunk = chunk[:0]
		}
	}

	bad := []kafka.Message{
		{Key: []byte("bad"), Value: []byte("not json")},
		{Key: []byte("bad"), Value: []byte(`{"id":"not-a-uuid","type":"like.created","from_id":1,"to_id":2,"created_at":"2026-11-27T00:00:00Z"}`)},
	}
	chunk = append(chunk, bad...)

	if err := writer.WriteMessages(ctx, chunk...); err != nil {
		return fmt.Errorf("write last chunk: %w", err)
	}

	slog.Info("generated", "likes", likes, "matches", matches, "bad", len(bad))
	return nil
}
