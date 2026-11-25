package producer

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/andreev012789-hash/likes-matches/internal/event"
)

type Producer struct {
	writer *kafka.Writer
}

func New(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			BatchTimeout: 10 * time.Millisecond,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, e event.Event) error {
	body, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(e.FromID, 10)),
		Value: body,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
