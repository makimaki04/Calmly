package models

import (
	"time"

	"github.com/google/uuid"
)

type DumpStatus string

const (
	DumpStatusNew             DumpStatus = "new"
	DumpStatusWaitingAnalysis DumpStatus = "waiting_analysis"
	DumpStatusWaitingAnswers  DumpStatus = "waiting_answers"
	DumpStatusPlanned         DumpStatus = "planned"
	DumpStatusAbandoned       DumpStatus = "abandoned"
)

type Dump struct {
	ID           uuid.UUID
	UserID       *uuid.UUID
	GuestID      *uuid.UUID
	Status       DumpStatus
	RawText      *string
	RawDeletedAt *time.Time
	RawExpiresAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
