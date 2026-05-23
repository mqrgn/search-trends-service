package storage

import (
	"sync"
	"time"

	"github.com/yourusername/search-trends/pkg/models"
)

const (
	WindowDuration = 5 * time.Minute
	BucketCount    = 5
	BucketDuration = WindowDuration / BucketCount
)

type QueryStats struct {
	TotalCount   int64
	UniqueUsers  map[string]struct{}
	UniqueSessions map[string]struct{}
}

type TimeBucket struct {
	Queries   map[string]*QueryStats
	Timestamp time.Time
}

type TrendStorage struct {
	mu           sync.RWMutex
	buckets      [BucketCount]*TimeBucket
	currentIndex int
	stopList     map[string]struct{}
	cachedTop    []models.TopQuery
	lastUpdate   time.Time
}

func NewTrendStorage() *TrendStorage {
	ts := &TrendStorage{
		stopList:  make(map[string]struct{}),
		cachedTop: make([]models.TopQuery, 0),
	}

	for i := 0; i < BucketCount; i++ {
		ts.buckets[i] = &TimeBucket{
			Queries:   make(map[string]*QueryStats),
			Timestamp: time.Now().Add(-time.Duration(BucketCount-i) * BucketDuration),
		}
	}

	return ts
}

func (ts *TrendStorage) AddEvent(event models.SearchEvent) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if _, blocked := ts.stopList[event.Query]; blocked {
		return
	}

	now := time.Now()
	ts.rotateBucketsIfNeeded(now)

	bucket := ts.buckets[ts.currentIndex]
	stats, exists := bucket.Queries[event.Query]
	if !exists {
		stats = &QueryStats{
			UniqueUsers:    make(map[string]struct{}),
			UniqueSessions: make(map[string]struct{}),
		}
		bucket.Queries[event.Query] = stats
	}

	stats.TotalCount++
	stats.UniqueUsers[event.UserID] = struct{}{}
	stats.UniqueSessions[event.SessionID] = struct{}{}
}

func (ts *TrendStorage) rotateBucketsIfNeeded(now time.Time) {
	bucket := ts.buckets[ts.currentIndex]
	if now.Sub(bucket.Timestamp) >= BucketDuration {
		ts.currentIndex = (ts.currentIndex + 1) % BucketCount
		ts.buckets[ts.currentIndex] = &TimeBucket{
			Queries:   make(map[string]*QueryStats),
			Timestamp: now,
		}
	}
}

func (ts *TrendStorage) GetTopQueries(limit int) []models.TopQuery {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	if time.Since(ts.lastUpdate) < time.Second && len(ts.cachedTop) > 0 {
		if limit > len(ts.cachedTop) {
			limit = len(ts.cachedTop)
		}
		return ts.cachedTop[:limit]
	}

	return ts.computeTop(limit)
}

func (ts *TrendStorage) computeTop(limit int) []models.TopQuery {
	aggregated := make(map[string]*QueryStats)
	cutoff := time.Now().Add(-WindowDuration)

	for _, bucket := range ts.buckets {
		if bucket.Timestamp.Before(cutoff) {
			continue
		}

		for query, stats := range bucket.Queries {
			if _, blocked := ts.stopList[query]; blocked {
				continue
			}

			agg, exists := aggregated[query]
			if !exists {
				agg = &QueryStats{
					UniqueUsers:    make(map[string]struct{}),
					UniqueSessions: make(map[string]struct{}),
				}
				aggregated[query] = agg
			}

			agg.TotalCount += stats.TotalCount
			for user := range stats.UniqueUsers {
				agg.UniqueUsers[user] = struct{}{}
			}
			for session := range stats.UniqueSessions {
				agg.UniqueSessions[session] = struct{}{}
			}
		}
	}

	result := make([]models.TopQuery, 0, len(aggregated))
	for query, stats := range aggregated {
		score := ts.calculateScore(stats)
		if score > 0 {
			result = append(result, models.TopQuery{
				Query: query,
				Count: score,
			})
		}
	}

	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Count > result[i].Count {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if limit > len(result) {
		limit = len(result)
	}

	return result[:limit]
}

func (ts *TrendStorage) calculateScore(stats *QueryStats) int64 {
	uniqueUsers := int64(len(stats.UniqueUsers))
	if uniqueUsers == 0 {
		return 0
	}

	ratio := float64(stats.TotalCount) / float64(uniqueUsers)

	const suspiciousThreshold = 10.0
	if ratio > suspiciousThreshold {
		penalty := ratio / suspiciousThreshold
		return int64(float64(stats.TotalCount) / penalty)
	}

	return stats.TotalCount
}

func (ts *TrendStorage) AddToStopList(query string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.stopList[query] = struct{}{}
}

func (ts *TrendStorage) RemoveFromStopList(query string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.stopList, query)
}

func (ts *TrendStorage) GetStopList() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	result := make([]string, 0, len(ts.stopList))
	for query := range ts.stopList {
		result = append(result, query)
	}
	return result
}
