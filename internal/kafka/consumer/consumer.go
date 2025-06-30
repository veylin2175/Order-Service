package consumer

import (
	"L0/internal/config"
	"L0/internal/models"
	"L0/internal/service"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"sync"
)

type Consumer struct {
	reader  *kafka.Reader
	service *service.OrderService
}

func NewConsumer(cfg config.Kafka, orderService *service.OrderService) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     cfg.Brokers,
			Topic:       cfg.Topic,
			GroupID:     cfg.GroupID,
			StartOffset: kafka.LastOffset,
			MaxBytes:    10e6,
		}),
		service: orderService,
	}
}

func (c *Consumer) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("stopping kafka consumer...")
			if err := c.reader.Close(); err != nil {
				log.Printf("failed to close kafka reader: %v", err)
			}
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("kafka read error: %v", err)
				continue
			}

			var order models.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("failed to unmarshal order: %v", err)
				continue
			}

			if err := c.service.SaveOrder(&order); err != nil {
				log.Printf("failed to save order: %v", err)
			} else {
				log.Printf("processed order: %s", order.OrderUID)
			}
		}
	}
}
