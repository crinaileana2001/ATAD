package entities

import "time"

type ClickEvent struct {
	ID         uint      `gorm:"primaryKey"`
	URLID      uint      `gorm:"index;not null"`
	CreatedAt  time.Time `gorm:"index;not null"`
	IPHash     string    `gorm:"size:64;index;not null"`
	Referrer   string    `gorm:"size:512"`
	UserAgent  string    `gorm:"size:512"`
	GeoCountry string    `gorm:"size:2;index"`
}
