package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;size:64" json:"username"`
	IsScammer bool      `json:"is_scammer"`
	CreatedAt time.Time `json:"created_at"`
}

type Ad struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    int64          `json:"user_id"`
	Username  string         `gorm:"size:64" json:"username"`
	Title     string         `gorm:"size:128" json:"title"`
	Desc      string         `gorm:"size:1024" json:"desc"`
	PhotoID   string         `gorm:"size:256" json:"photo_id"` // Telegram file_id
	Category  string         `gorm:"size:32" json:"category"`  // services, trade, other
	Subcat    string         `gorm:"size:64" json:"subcat"`
	Filter1   string         `gorm:"size:64" json:"filter1"`
	IsPremium bool           `json:"is_premium"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
