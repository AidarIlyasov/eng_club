package models

import (
	"time"
)

type Member struct {
	ID       int     `json:"member_id"`
	Name     string  `json:"name"`
	Telegram *string `json:"telegram"`
}

// Place is a venue the club can meet at.
type Place struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	MetroArea string    `json:"metro_area"`
	MapURL    string    `json:"map_url"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Event is a scheduled club meeting before or after a session is run.
type Event struct {
	ID          int64      `json:"id"`
	PlaceID     int64      `json:"place_id"`
	ScheduledAt time.Time  `json:"scheduled_at"`
	Topic       string     `json:"topic"`
	NotifyAt    *time.Time `json:"notify_at,omitempty"`
	Status      string     `json:"status"`
	Type        string     `json:"type"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`

	// Place information (populated when joined with places table)
	Place *Place `json:"place,omitempty"`
}

type Participant struct {
	ID             int     `json:"member_id"`
	Name           string  `json:"name"`
	Telegram       *string `json:"telegram"`
	TelegramChatID *string `json:"telegram_chat_id,omitempty"`
	Phone          *string `json:"phone,omitempty"`
	Attended       bool    `json:"attended,omitempty"`
	TableID        *int64  `json:"table_id,omitempty"`
}

type Activity struct {
	ID        int
	Name      string
	Duration  int
	GroupSize int
}
