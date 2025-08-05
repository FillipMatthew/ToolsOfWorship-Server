package feed

import (
	"time"

	"github.com/google/uuid"
)

type ListRequest struct {
	Limit  *int       `json:"limit"`
	Before *time.Time `json:"before"`
	After  *time.Time `json:"after"`
}

type PostRequest struct {
	FellowshipId uuid.UUID `json:"fellowshipId"`
	CircleId     uuid.UUID `json:"circleId"`
	Heading      string    `json:"heading"`
	Article      string    `json:"article"`
}
