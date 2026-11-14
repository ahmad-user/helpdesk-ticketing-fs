package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"helpdesk-ticketing-system/backend/internal/model"

	"github.com/google/uuid"
)

// CommentRepository defines database persistence operations for the Comment entity.
type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	FindByTicketID(ctx context.Context, ticketID uuid.UUID, includeInternal bool) ([]*model.Comment, error)
}

type commentRepository struct {
	db *sql.DB
}

// NewCommentRepository constructs a new PostgreSQL CommentRepository instance.
func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{db: db}
}

// Create inserts a new comment or internal note into the database.
func (r *commentRepository) Create(ctx context.Context, comment *model.Comment) error {
	if r.db == nil {
		return ErrDatabaseNotConnected
	}

	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO comments (id, ticket_id, author_id, body, is_internal, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at;
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		comment.ID,
		comment.TicketID,
		comment.AuthorID,
		comment.Body,
		comment.IsInternal,
		comment.CreatedAt,
	).Scan(&comment.ID, &comment.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert comment: %w", err)
	}

	return nil
}

// FindByTicketID retrieves all comments associated with a ticket.
// If includeInternal is false (for requester callers), internal notes are filtered out at the DB level.
// Joins with the users table to prevent N+1 queries.
func (r *commentRepository) FindByTicketID(ctx context.Context, ticketID uuid.UUID, includeInternal bool) ([]*model.Comment, error) {
	if r.db == nil {
		return nil, ErrDatabaseNotConnected
	}

	query := `
		SELECT 
			c.id, c.ticket_id, c.author_id, c.body, c.is_internal, c.created_at,
			u.id, u.name, u.email, u.role
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.ticket_id = $1
		  AND ($2 = TRUE OR c.is_internal = FALSE)
		ORDER BY c.created_at ASC;
	`

	rows, err := r.db.QueryContext(ctx, query, ticketID, includeInternal)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		c := &model.Comment{
			Author: &model.UserSummary{},
		}
		err := rows.Scan(
			&c.ID,
			&c.TicketID,
			&c.AuthorID,
			&c.Body,
			&c.IsInternal,
			&c.CreatedAt,
			&c.Author.ID,
			&c.Author.Name,
			&c.Author.Email,
			&c.Author.Role,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("comments row error: %w", err)
	}

	return comments, nil
}
