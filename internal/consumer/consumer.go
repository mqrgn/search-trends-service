package consumer

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/yourusername/search-trends/internal/metrics"
	"github.com/yourusername/search-trends/internal/storage"
	"github.com/yourusername/search-trends/pkg/models"
)

type Consumer struct {
	nc      *nats.Conn
	storage *storage.TrendStorage
	metrics *metrics.Metrics
	subject string
}

func NewConsumer(natsURL, subject string, storage *storage.TrendStorage, metrics *metrics.Metrics) (*Consumer, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		nc:      nc,
		storage: storage,
		metrics: metrics,
		subject: subject,
	}, nil
}

func (c *Consumer) Start() error {
	_, err := c.nc.Subscribe(c.subject, func(msg *nats.Msg) {
		var event models.SearchEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			return
		}

		if event.Timestamp == 0 {
			event.Timestamp = time.Now().Unix()
		}

		c.storage.AddEvent(event)
		c.metrics.IncrementEventsProcessed()
	})

	if err != nil {
		return err
	}

	log.Printf("Consumer started, listening on subject: %s", c.subject)
	return nil
}

func (c *Consumer) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}
