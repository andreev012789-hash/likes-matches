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
	"github.com/andreev012789-hash/likes-matches/internal/repository"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

const (
	workerCount   = 3
	batchSize     = 100
	flushInterval = time.Second

	clickhouseAddr     = "localhost:9001"
	clickhouseUser     = "likes"
	clickhousePassword = "likes"
)

func commit(reader *kafka.Reader, msgs ...kafka.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := reader.CommitMessages(ctx, msgs...); err != nil {
		slog.Error("commit offsets", "count", len(msgs), "error", err)
		return
	}

	slog.Info("offsets committed", "count", len(msgs))
}

func flush(ctx context.Context, reader *kafka.Reader, store *repository.EventRepository, msgs []kafka.Message, events []event.Event) bool {
	if len(msgs) == 0 {
		return true
	}

	for {
		dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := store.InsertBatch(dbCtx, events)
		cancel()

		if err == nil {
			break
		}
		slog.Error("insert batch", "size", len(events), "error", err)

		select {
		case <-ctx.Done():
			return false
		case <-time.After(time.Second):
		}
	}

	commit(reader, msgs...)
	return true
}

func worker(ctx context.Context, id int, reader *kafka.Reader, store *repository.EventRepository, jobs <-chan kafka.Message, wg *sync.WaitGroup) {
	defer wg.Done()
	defer slog.Info("worker stopped", "worker", id)

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	var msgs []kafka.Message
	var events []event.Event

	for {
		select {
		case msg, ok := <-jobs:
			if !ok {
				flush(ctx, reader, store, msgs, events)
				return
			}

			msgs = append(msgs, msg)

			var e event.Event
			if err := json.Unmarshal(msg.Value, &e); err != nil {
				slog.Error("decode event", "partition", msg.Partition, "offset", msg.Offset, "error", err)
			} else if _, err := uuid.Parse(e.ID); err != nil {
				slog.Error("invalid event id", "partition", msg.Partition, "offset", msg.Offset, "error", err)
			} else {
				events = append(events, e)
			}

			if len(msgs) >= batchSize {
				if !flush(ctx, reader, store, msgs, events) {
					for range jobs {
					}
					return
				}
				msgs, events = nil, nil
			}

		case <-ticker.C:
			if !flush(ctx, reader, store, msgs, events) {
				for range jobs {
				}
				return
			}
			msgs, events = nil, nil
		}
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startCtx, startCancel := context.WithTimeout(ctx, 5*time.Second)
	store, err := repository.NewEventRepository(startCtx, clickhouseAddr, clickhouseUser, clickhousePassword)
	startCancel()
	if err != nil {
		slog.Error("connect clickhouse", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := store.Close(); err != nil {
			slog.Error("close clickhouse", "error", err)
		}
	}()

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
		go worker(ctx, i, reader, store, channels[i], &wg)
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
	stop()

	for _, ch := range channels {
		close(ch)
	}
	wg.Wait()
	slog.Info("consumer stopped")
}
