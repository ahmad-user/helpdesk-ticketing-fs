package model

import (
	"time"

	"github.com/google/uuid"
)

// Comment represents the comments table domain entity.
type Comment struct {
	ID         uuid.UUID    `json:"id"`
	TicketID   uuid.UUID    `json:"ticket_id"`
	AuthorID   uuid.UUID    `json:"author_id"`
	Body       string       `json:"body"`
	IsInternal bool         `json:"is_internal"`
	CreatedAt  time.Time    `json:"created_at"`
	Author     *UserSummary `json:"author,omitempty"`
}

// CreateCommentRequest represents the payload for creating a comment or internal note.
type CreateCommentRequest struct {
	Body       string `json:"body" binding:"required,min=1,max=5000"`
	IsInternal *bool  `json:"is_internal,omitempty"`
}

// CommentResponse represents the external representation of a comment.
type CommentResponse struct {
	ID         uuid.UUID    `json:"id"`
	TicketID   uuid.UUID    `json:"ticket_id"`
	AuthorID   uuid.UUID    `json:"author_id"`
	Author     *UserSummary `json:"author,omitempty"`
	Body       string       `json:"body"`
	IsInternal bool         `json:"is_internal"`
	CreatedAt  time.Time    `json:"created_at"`
}

// ToResponse maps a domain Comment entity to CommentResponse DTO.
func (c *Comment) ToResponse() CommentResponse {
	return CommentResponse{
		ID:         c.ID,
		TicketID:   c.TicketID,
		AuthorID:   c.AuthorID,
		Author:     c.Author,
		Body:       c.Body,
		IsInternal: c.IsInternal,
		CreatedAt:  c.CreatedAt,
	}
}
