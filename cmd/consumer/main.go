package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/andreev012789-hash/likes-matches/internal/config"
	"github.com/andreev012789-hash/likes-matches/internal/event"
	"github.com/segmentio/kafka-go"
)

const workerCount = 3

func commit(reader *kafka.Reader, msg kafka.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := reader.CommitMessages(ctx, msg); err != nil {
		slog.Error("commit offset", "partition", msg.Partition, "offset", msg.Offset, "error", err)
		return
	}

	slog.Info("offset committed", "partition", msg.Partition, "offset", msg.Offset)
}

func worker(id int, reader *kafka.Reader, jobs <-chan kafka.Message, wg *sync.WaitGroup) {
	defer wg.Done()

	for msg := range jobs {

		var e event.Event
		if err := json.Unmarshal(msg.Value, &e); err != nil {
			slog.Error("decode event", "partition", msg.Partition, "offset", msg.Offset, "error", err)
			commit(reader, msg)
			continue
		}

		slog.Info("event received",
			"worker", id,
			"type", e.Type,
			"id", e.ID,
			"from_id", e.FromID,
			"to_id", e.ToID,
			"partition", msg.Partition,
			"offset", msg.Offset,
		)
		commit(reader, msg)
	}
	slog.Info("worker stopped", "worker", id)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaBrokers},
		GroupID: "likes-consumer",
		Topic:   "events",
	})
	defer func() {
		if err := reader.Close(); err != nil {
			slog.Error("close kafka reader", "error", err)
		}
	}()

	slog.Info("consumer started", "group", "likes-consumer", "topic", "events")

	var wg sync.WaitGroup
	channels := make([]chan kafka.Message, workerCount)

	for i := range channels {
		channels[i] = make(chan kafka.Message, 10)
		wg.Add(1)
		go worker(i, reader, channels[i], &wg)
	}

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				slog.Error("fetch message", "error", err)
			}
			break
		}
		channels[msg.Partition%workerCount] <- msg
	}

	for _, ch := range channels {
		close(ch)
	}
	wg.Wait()
	slog.Info("consumer stopped")
}
