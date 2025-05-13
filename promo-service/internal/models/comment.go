package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID          uuid.UUID
	PromoID     uuid.UUID
	ClientID    uuid.UUID
	CommentText string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
