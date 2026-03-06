package models

import "github.com/google/uuid"

type UserFeedback struct {
	DumpID uuid.UUID
	Text   string
}
