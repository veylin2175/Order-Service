package main

import (
	"L0/internal/config"
	"L0/internal/models"
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

func main() {
	cfg := config.MustLoad()

	writer := kafka.Writer{
		Addr:  kafka.TCP(cfg.Kafka.Brokers[0]),
		Topic: cfg.Kafka.Topic,
	}

	order := models.Order{
		OrderUID:    "test123",
		TrackNumber: "123456",
		Entry:       "123456",
		Delivery: models.Delivery{
			Name:    "Vasya",
			Phone:   "89123456789",
			Zip:     "111",
			City:    "Moscow",
			Address: "Russia",
			Region:  "VAO",
			Email:   "email@gmail.com",
		},
		Payment: models.Payment{
			Transaction:  "tra",
			RequestID:    "reqID",
			Currency:     "USD",
			Provider:     "WB",
			Amount:       560,
			PaymentDT:    23456,
			Bank:         "wb",
			DeliveryCost: 145654,
			GoodsTotal:   111,
			CustomFee:    222,
		},
		Items: []models.Item{
			models.Item{
				ChrtID:      12345,
				TrackNumber: "1212",
				Price:       12345,
				RID:         "fff",
				Name:        "adfa",
				Sale:        10,
				Size:        "10",
				TotalPrice:  1999,
				NmID:        234324232,
				Brand:       "channel",
				Status:      1,
			},
		},
		Locale:            "locale",
		InternalSignature: "a1",
		CustomerID:        "12345fs1212",
		DeliveryService:   "samokat",
		Shardkey:          "22222",
		SmID:              1234323231,
		DateCreated:       time.Now(),
		OofShard:          "oof",
	}

	value, _ := json.Marshal(order)
	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: value,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("message sent")
}
