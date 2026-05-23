package models

type TopQuery struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
}

type TopQueriesResponse struct {
	Queries   []TopQuery `json:"queries"`
	Timestamp int64      `json:"timestamp"`
}
