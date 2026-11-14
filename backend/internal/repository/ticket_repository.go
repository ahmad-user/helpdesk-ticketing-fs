package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"helpdesk-ticketing-system/backend/internal/model"

	"github.com/google/uuid"
)

// Repository domain errors for tickets.
var (
	ErrTicketNotFound = errors.New("ticket not found")
)

// TicketRepository defines database persistence operations for the Ticket domain entity.
type TicketRepository interface {
	Create(ctx context.Context, ticket *model.Ticket) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Ticket, error)
	FindAll(ctx context.Context, filter model.TicketFilter) ([]*model.Ticket, int64, error)
	Update(ctx context.Context, ticket *model.Ticket) error
}

type ticketRepository struct {
	db *sql.DB
}

// NewTicketRepository constructs a new PostgreSQL TicketRepository instance.
func NewTicketRepository(db *sql.DB) TicketRepository {
	return &ticketRepository{db: db}
}

// Create persists a new ticket entity in PostgreSQL.
func (r *ticketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	if r.db == nil {
		return ErrDatabaseNotConnected
	}

	if ticket.ID == uuid.Nil {
		ticket.ID = uuid.New()
	}

	query := `
		INSERT INTO tickets (id, title, description, category, priority, status, requester_id, assignee_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at;
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		ticket.ID,
		ticket.Title,
		ticket.Description,
		ticket.Category,
		ticket.Priority,
		ticket.Status,
		ticket.RequesterID,
		ticket.AssigneeID,
		ticket.CreatedAt,
		ticket.UpdatedAt,
	).Scan(&ticket.CreatedAt, &ticket.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert ticket: %w", err)
	}

	return nil
}

// FindByID retrieves a single ticket by its UUID, including joined requester and assignee details.
func (r *ticketRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Ticket, error) {
	if r.db == nil {
		return nil, ErrDatabaseNotConnected
	}

	query := `
		SELECT 
			t.id, t.title, t.description, t.category, t.priority, t.status, 
			t.requester_id, t.assignee_id, t.created_at, t.updated_at,
			req.id, req.name, req.email, req.role,
			asg.id, asg.name, asg.email, asg.role
		FROM tickets t
		LEFT JOIN users req ON t.requester_id = req.id
		LEFT JOIN users asg ON t.assignee_id = asg.id
		WHERE t.id = $1;
	`

	ticket := &model.Ticket{}
	var reqID, reqName, reqEmail, reqRole sql.NullString
	var asgID, asgName, asgEmail, asgRole sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ticket.ID,
		&ticket.Title,
		&ticket.Description,
		&ticket.Category,
		&ticket.Priority,
		&ticket.Status,
		&ticket.RequesterID,
		&ticket.AssigneeID,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
		&reqID, &reqName, &reqEmail, &reqRole,
		&asgID, &asgName, &asgEmail, &asgRole,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTicketNotFound
		}
		return nil, fmt.Errorf("failed to find ticket by id: %w", err)
	}

	// Populate Requester summary
	if reqID.Valid {
		parsedReqID, _ := uuid.Parse(reqID.String)
		ticket.Requester = &model.UserSummary{
			ID:    parsedReqID,
			Name:  reqName.String,
			Email: reqEmail.String,
			Role:  reqRole.String,
		}
	}

	// Populate Assignee summary if assigned
	if asgID.Valid {
		parsedAsgID, _ := uuid.Parse(asgID.String)
		ticket.Assignee = &model.UserSummary{
			ID:    parsedAsgID,
			Name:  asgName.String,
			Email: asgEmail.String,
			Role:  asgRole.String,
		}
	}

	return ticket, nil
}

// FindAll queries tickets matching filter criteria, executing dynamic parameterized SQL to prevent injection.
func (r *ticketRepository) FindAll(ctx context.Context, filter model.TicketFilter) ([]*model.Ticket, int64, error) {
	if r.db == nil {
		return nil, 0, ErrDatabaseNotConnected
	}

	var conditions []string
	var args []interface{}
	paramIdx := 1

	if filter.RequesterID != nil {
		conditions = append(conditions, fmt.Sprintf("t.requester_id = $%d", paramIdx))
		args = append(args, *filter.RequesterID)
		paramIdx++
	}

	if filter.AssigneeID != nil {
		conditions = append(conditions, fmt.Sprintf("t.assignee_id = $%d", paramIdx))
		args = append(args, *filter.AssigneeID)
		paramIdx++
	}

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", paramIdx))
		args = append(args, filter.Status)
		paramIdx++
	}

	if filter.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("t.priority = $%d", paramIdx))
		args = append(args, filter.Priority)
		paramIdx++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("t.title ILIKE $%d", paramIdx))
		args = append(args, "%"+filter.Search+"%")
		paramIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 1. Count total matching rows
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tickets t %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets: %w", err)
	}

	// Whitelist sort column
	sortCol := "t.created_at"
	switch strings.ToLower(filter.SortBy) {
	case "updated_at":
		sortCol = "t.updated_at"
	case "priority":
		// Custom priority weighting: urgent > high > medium > low
		sortCol = `CASE t.priority 
			WHEN 'urgent' THEN 4 
			WHEN 'high' THEN 3 
			WHEN 'medium' THEN 2 
			WHEN 'low' THEN 1 
			ELSE 0 END`
	case "status":
		sortCol = "t.status"
	case "created_at":
		sortCol = "t.created_at"
	}

	// Whitelist sort direction
	sortDirection := "DESC"
	if strings.EqualFold(filter.SortOrder, "asc") {
		sortDirection = "ASC"
	}

	// Pagination defaults
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// 2. Fetch data rows with joined user details
	dataQuery := fmt.Sprintf(`
		SELECT 
			t.id, t.title, t.description, t.category, t.priority, t.status, 
			t.requester_id, t.assignee_id, t.created_at, t.updated_at,
			req.id, req.name, req.email, req.role,
			asg.id, asg.name, asg.email, asg.role
		FROM tickets t
		LEFT JOIN users req ON t.requester_id = req.id
		LEFT JOIN users asg ON t.assignee_id = asg.id
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d;
	`, whereClause, sortCol, sortDirection, paramIdx, paramIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query tickets: %w", err)
	}
	defer rows.Close()

	tickets := make([]*model.Ticket, 0)
	for rows.Next() {
		ticket := &model.Ticket{}
		var reqID, reqName, reqEmail, reqRole sql.NullString
		var asgID, asgName, asgEmail, asgRole sql.NullString

		if err := rows.Scan(
			&ticket.ID,
			&ticket.Title,
			&ticket.Description,
			&ticket.Category,
			&ticket.Priority,
			&ticket.Status,
			&ticket.RequesterID,
			&ticket.AssigneeID,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
			&reqID, &reqName, &reqEmail, &reqRole,
			&asgID, &asgName, &asgEmail, &asgRole,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan ticket row: %w", err)
		}

		if reqID.Valid {
			parsedReqID, _ := uuid.Parse(reqID.String)
			ticket.Requester = &model.UserSummary{
				ID:    parsedReqID,
				Name:  reqName.String,
				Email: reqEmail.String,
				Role:  reqRole.String,
			}
		}

		if asgID.Valid {
			parsedAsgID, _ := uuid.Parse(asgID.String)
			ticket.Assignee = &model.UserSummary{
				ID:    parsedAsgID,
				Name:  asgName.String,
				Email: asgEmail.String,
				Role:  asgRole.String,
			}
		}

		tickets = append(tickets, ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating ticket rows: %w", err)
	}

	return tickets, total, nil
}

// Update updates an existing ticket's status, priority, assignee_id, and updated_at.
func (r *ticketRepository) Update(ctx context.Context, ticket *model.Ticket) error {
	if r.db == nil {
		return ErrDatabaseNotConnected
	}

	query := `
		UPDATE tickets
		SET 
			status = $1,
			priority = $2,
			assignee_id = $3,
			updated_at = $4
		WHERE id = $5
		RETURNING updated_at;
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		ticket.Status,
		ticket.Priority,
		ticket.AssigneeID,
		ticket.UpdatedAt,
		ticket.ID,
	).Scan(&ticket.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTicketNotFound
		}
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	return nil
}
