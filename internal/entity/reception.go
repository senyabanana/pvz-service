package entity

import (
	"time"

	"github.com/google/uuid"
)

type ReceptionStatus string

const (
	StatusInProgress ReceptionStatus = "in_progress"
	StatusClosed     ReceptionStatus = "close"
)

type Reception struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	DateTime  time.Time       `json:"dateTime" db:"date_time"`
	PVZID     uuid.UUID       `json:"pvzId" db:"pvz_id"`
	Status    ReceptionStatus `json:"status" db:"status"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	ClosedAt  *time.Time      `json:"closed_at,omitempty" db:"closed_at"`
}
