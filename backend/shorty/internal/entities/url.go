package entities

import "time"

type URL struct {
	ID        uint       `gorm:"primaryKey"`
	Code      string     `gorm:"uniqueIndex;size:16;not null"`
	Original  string     `gorm:"size:2048;not null"`
	CreatedAt time.Time  `gorm:"not null"`
	ExpiresAt *time.Time `gorm:"index"`
}
