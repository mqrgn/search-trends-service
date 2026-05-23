package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/yourusername/search-trends/pkg/models"
)

var queries = []string{
	"iphone 15",
	"samsung galaxy",
	"macbook pro",
	"airpods",
	"playstation 5",
	"nike shoes",
	"adidas sneakers",
	"laptop",
	"headphones",
	"smartwatch",
}

var users = []string{
	"user1", "user2", "user3", "user4", "user5",
	"user6", "user7", "user8", "user9", "user10",
}

func main() {
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	log.Println("Producer started, sending events...")

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		event := models.SearchEvent{
			Query:     queries[rand.Intn(len(queries))],
			Timestamp: time.Now().Unix(),
			UserID:    users[rand.Intn(len(users))],
			SessionID: fmt.Sprintf("session-%d", rand.Intn(20)),
		}

		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
			continue
		}

		if err := nc.Publish("search.events", data); err != nil {
			log.Printf("Failed to publish event: %v", err)
		}
	}
}
