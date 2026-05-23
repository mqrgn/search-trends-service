package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/yourusername/search-trends/internal/metrics"
	"github.com/yourusername/search-trends/internal/storage"
	"github.com/yourusername/search-trends/pkg/models"
)

type Handler struct {
	storage *storage.TrendStorage
	metrics *metrics.Metrics
}

func NewHandler(storage *storage.TrendStorage, metrics *metrics.Metrics) *Handler {
	return &Handler{
		storage: storage,
		metrics: metrics,
	}
}

func (h *Handler) GetTopQueries(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.metrics.RecordRequest("get_top", time.Since(start).Seconds())
	}()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	queries := h.storage.GetTopQueries(limit)

	response := models.TopQueriesResponse{
		Queries:   queries,
		Timestamp: time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) AddToStopList(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.metrics.RecordRequest("add_stoplist", time.Since(start).Seconds())
	}()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	h.storage.AddToStopList(req.Query)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "added"})
}

func (h *Handler) RemoveFromStopList(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.metrics.RecordRequest("remove_stoplist", time.Since(start).Seconds())
	}()

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	h.storage.RemoveFromStopList(req.Query)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "removed"})
}

func (h *Handler) GetStopList(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.metrics.RecordRequest("get_stoplist", time.Since(start).Seconds())
	}()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stopList := h.storage.GetStopList()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"stoplist": stopList})
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
