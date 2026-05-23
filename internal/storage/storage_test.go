package storage

import (
	"testing"
	"time"

	"github.com/yourusername/search-trends/pkg/models"
)

func TestTrendStorage_AddEvent(t *testing.T) {
	storage := NewTrendStorage()

	event := models.SearchEvent{
		Query:     "iphone 15",
		Timestamp: time.Now().Unix(),
		UserID:    "user1",
		SessionID: "session1",
	}

	storage.AddEvent(event)

	top := storage.GetTopQueries(10)
	if len(top) != 1 {
		t.Errorf("Expected 1 query, got %d", len(top))
	}

	if top[0].Query != "iphone 15" {
		t.Errorf("Expected 'iphone 15', got '%s'", top[0].Query)
	}
}

func TestTrendStorage_StopList(t *testing.T) {
	storage := NewTrendStorage()

	storage.AddToStopList("blocked")

	event := models.SearchEvent{
		Query:     "blocked",
		Timestamp: time.Now().Unix(),
		UserID:    "user1",
		SessionID: "session1",
	}

	storage.AddEvent(event)

	top := storage.GetTopQueries(10)
	if len(top) != 0 {
		t.Errorf("Expected 0 queries (blocked), got %d", len(top))
	}
}

func TestTrendStorage_AnomalyDetection(t *testing.T) {
	storage := NewTrendStorage()

	for i := 0; i < 100; i++ {
		event := models.SearchEvent{
			Query:     "spam",
			Timestamp: time.Now().Unix(),
			UserID:    "user1",
			SessionID: "session1",
		}
		storage.AddEvent(event)
	}

	normalEvent := models.SearchEvent{
		Query:     "normal",
		Timestamp: time.Now().Unix(),
		UserID:    "user2",
		SessionID: "session2",
	}
	storage.AddEvent(normalEvent)

	top := storage.GetTopQueries(10)

	var spamScore, normalScore int64
	for _, q := range top {
		if q.Query == "spam" {
			spamScore = q.Count
		}
		if q.Query == "normal" {
			normalScore = q.Count
		}
	}

	if spamScore >= 100 {
		t.Errorf("Spam query should be penalized, got score %d", spamScore)
	}

	if normalScore != 1 {
		t.Errorf("Normal query should have score 1, got %d", normalScore)
	}
}

func BenchmarkTrendStorage_AddEvent(b *testing.B) {
	storage := NewTrendStorage()

	event := models.SearchEvent{
		Query:     "test query",
		Timestamp: time.Now().Unix(),
		UserID:    "user1",
		SessionID: "session1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.AddEvent(event)
	}
}

func BenchmarkTrendStorage_GetTopQueries(b *testing.B) {
	storage := NewTrendStorage()

	for i := 0; i < 1000; i++ {
		event := models.SearchEvent{
			Query:     "query" + string(rune(i%100)),
			Timestamp: time.Now().Unix(),
			UserID:    "user" + string(rune(i)),
			SessionID: "session" + string(rune(i)),
		}
		storage.AddEvent(event)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.GetTopQueries(10)
	}
}
