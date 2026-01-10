package dtos

import "time"

type URLListItem struct {
	Code           string     `json:"code"`
	ShortURL       string     `json:"short_url"`
	Original       string     `json:"original"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	Clicks         int64      `json:"clicks"`
	UniqueVisitors int64      `json:"unique_visitors"`
}
