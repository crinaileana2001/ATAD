package dtos

import "time"

type StatsResponse struct {
	Original       string           `json:"original"`
	Clicks         int64            `json:"clicks"`
	UniqueVisitors int64            `json:"unique_visitors"`
	ExpiresAt      *time.Time       `json:"expires_at,omitempty"`
	Countries      map[string]int64 `json:"countries,omitempty"`
}
