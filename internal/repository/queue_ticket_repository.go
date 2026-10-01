package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"qback/internal/database"
	"qback/internal/models"
)

var (
	ErrQueueTicketNotFound = errors.New("queue ticket not found")
)

type QueueTicketRepository interface {
	Create(ctx context.Context, ticket *models.QueueTicket) error
	GetByID(ctx context.Context, id int) (*models.QueueTicket, error)
	GetActiveByUserID(ctx context.Context, userID int) ([]*models.QueueTicket, error)
	GetHistoryByUserID(ctx context.Context, userID int) ([]*models.QueueTicket, error)
	UpdateStatus(ctx context.Context, id int, status string) error
	Cancel(ctx context.Context, id int) error
	GetWaitingByServiceID(ctx context.Context, serviceID int) ([]*models.QueueTicket, error)
	GetStaffQueues(ctx context.Context, ownerID int, venueID, serviceID *int, status *string) ([]*models.QueueTicket, error)
	CallNext(ctx context.Context, ownerID, serviceID int) (*models.QueueTicket, error)
	Serve(ctx context.Context, ownerID, ticketID int) (*models.QueueTicket, error)
	NoShow(ctx context.Context, ownerID, ticketID int) (*models.QueueTicket, error)
	CancelByStaff(ctx context.Context, ownerID, ticketID int) (*models.QueueTicket, error)
}

type queueTicketRepository struct {
	db *database.DB
}

func NewQueueTicketRepository(db *database.DB) QueueTicketRepository {
	return &queueTicketRepository{db: db}
}

func scanTickets(rows pgx.Rows) ([]*models.QueueTicket, error) {
	var tickets []*models.QueueTicket
	for rows.Next() {
		t := &models.QueueTicket{}
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.VenueID,
			&t.ServiceID,
			&t.TicketNumber,
			&t.Status,
			&t.JoinedAt,
			&t.ServedAt,
			&t.CancelledAt,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *queueTicketRepository) Create(ctx context.Context, ticket *models.QueueTicket) error {
	query := `
		INSERT INTO queue_tickets (user_id, venue_id, service_id, ticket_number, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, joined_at, created_at, updated_at
	`

	status := ticket.Status
	if status == "" {
		status = "waiting"
	}
	ticket.Status = status

	err := r.db.Pool.QueryRow(ctx, query,
		ticket.UserID,
		ticket.VenueID,
		ticket.ServiceID,
		ticket.TicketNumber,
		ticket.Status,
	).Scan(
		&ticket.ID,
		&ticket.JoinedAt,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)
	return err
}

func (r *queueTicketRepository) GetByID(ctx context.Context, id int) (*models.QueueTicket, error) {
	query := `
		SELECT id, user_id, venue_id, service_id, ticket_number, status, joined_at, served_at, cancelled_at, created_at, updated_at
		FROM queue_tickets
		WHERE id = $1
	`
	ticket := &models.QueueTicket{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&ticket.ID,
		&ticket.UserID,
		&ticket.VenueID,
		&ticket.ServiceID,
		&ticket.TicketNumber,
		&ticket.Status,
		&ticket.JoinedAt,
		&ticket.ServedAt,
		&ticket.CancelledAt,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrQueueTicketNotFound
		}
		return nil, err
	}
	return ticket, nil
}

func (r *queueTicketRepository) GetActiveByUserID(ctx context.Context, userID int) ([]*models.QueueTicket, error) {
	query := `
		SELECT id, user_id, venue_id, service_id, ticket_number, status, joined_at, served_at, cancelled_at, created_at, updated_at
		FROM queue_tickets
		WHERE user_id = $1 AND status IN ('waiting', 'serving')
		ORDER BY joined_at DESC, id DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(rows)
}

func (r *queueTicketRepository) GetHistoryByUserID(ctx context.Context, userID int) ([]*models.QueueTicket, error) {
	query := `
		SELECT id, user_id, venue_id, service_id, ticket_number, status, joined_at, served_at, cancelled_at, created_at, updated_at
		FROM queue_tickets
		WHERE user_id = $1 AND status IN ('served', 'cancelled', 'no_show')
		ORDER BY created_at DESC, id DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(rows)
}

func (r *queueTicketRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE queue_tickets
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	cmdTag, err := r.db.Pool.Exec(ctx, query, status, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrQueueTicketNotFound
	}
	return nil
}

func (r *queueTicketRepository) Cancel(ctx context.Context, id int) error {
	query := `
		UPDATE queue_tickets
		SET status = 'cancelled', cancelled_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	cmdTag, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrQueueTicketNotFound
	}
	return nil
}

func (r *queueTicketRepository) GetWaitingByServiceID(ctx context.Context, serviceID int) ([]*models.QueueTicket, error) {
	query := `
		SELECT id, user_id, venue_id, service_id, ticket_number, status, joined_at, served_at, cancelled_at, created_at, updated_at
		FROM queue_tickets
		WHERE service_id = $1 AND status = 'waiting'
		ORDER BY joined_at ASC, id ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(rows)
}

func (r *queueTicketRepository) GetStaffQueues(ctx context.Context, ownerID int, venueID, serviceID *int, status *string) ([]*models.QueueTicket, error) {
	query := `
		SELECT qt.id, qt.user_id, qt.venue_id, qt.service_id, qt.ticket_number, qt.status, qt.joined_at, qt.served_at, qt.cancelled_at, qt.created_at, qt.updated_at
		FROM queue_tickets qt
		JOIN services s ON qt.service_id = s.id
		JOIN venues v ON s.venue_id = v.id
		WHERE v.owner_id = $1
	`
	args := []any{ownerID}
	argID := 2

	if venueID != nil {
		query += fmt.Sprintf(" AND qt.venue_id = $%d", argID)
		args = append(args, *venueID)
		argID++
	}
	if serviceID != nil {
		query += fmt.Sprintf(" AND qt.service_id = $%d", argID)
		args = append(args, *serviceID)
		argID++
	}
	if status != nil {
		query += fmt.Sprintf(" AND qt.status = $%d", argID)
		args = append(args, *status)
		argID++
	}

	query += " ORDER BY qt.joined_at ASC, qt.id ASC"

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(rows)
}

func (r *queueTicketRepository) CallNext(ctx context.Context, ownerID, serviceID int) (*models.QueueTicket, error) {
	query := `
		UPDATE queue_tickets
		SET status = 'serving', updated_at = CURRENT_TIMESTAMP
		WHERE id = (
			SELECT qt.id
			FROM queue_tickets qt
			JOIN services s ON qt.service_id = s.id
			JOIN venues v ON s.venue_id = v.id
			WHERE qt.service_id = $1 
			  AND qt.status = 'waiting'
			  AND v.owner_id = $2
			ORDER BY qt.joined_at ASC, qt.id ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, user_id, venue_id, service_id, ticket_number, status, joined_at, served_at, cancelled_at, created_at, updated_at
	`
	ticket := &models.QueueTicket{}
	err := r.db.Pool.QueryRow(ctx, query, serviceID, ownerID).Scan(
		&ticket.ID, &ticket.UserID, &ticket.VenueID, &ticket.ServiceID,
		&ticket.TicketNumber, &ticket.Status, &ticket.JoinedAt,
		&ticket.ServedAt, &ticket.CancelledAt, &ticket.CreatedAt, &ticket.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrQueueTicketNotFound
		}
		return nil, err
	}
	return ticket, nil
}

func (r *queueTicketRepository) Serve(ctx context.Context, ownerID, ticketID int) (*models.QueueTicket, error) {
	query := `
		UPDATE queue_tickets qt
		SET status = 'served', served_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		FROM services s, venues v
		WHERE qt.id = $1
		  AND qt.status = 'serving'
		  AND qt.service_id = s.id
		  AND s.venue_id = v.id
		  AND v.owner_id = $2
		RETURNING qt.id, qt.user_id, qt.venue_id, qt.service_id, qt.ticket_number, qt.status, qt.joined_at, qt.served_at, qt.cancelled_at, qt.created_at, qt.updated_at
	`
	ticket := &models.QueueTicket{}
	err := r.db.Pool.QueryRow(ctx, query, ticketID, ownerID).Scan(
		&ticket.ID, &ticket.UserID, &ticket.VenueID, &ticket.ServiceID,
		&ticket.TicketNumber, &ticket.Status, &ticket.JoinedAt,
		&ticket.ServedAt, &ticket.CancelledAt, &ticket.CreatedAt, &ticket.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrQueueTicketNotFound
		}
		return nil, err
	}
	return ticket, nil
}

func (r *queueTicketRepository) NoShow(ctx context.Context, ownerID, ticketID int) (*models.QueueTicket, error) {
	query := `
		UPDATE queue_tickets qt
		SET status = 'no_show', updated_at = CURRENT_TIMESTAMP
		FROM services s, venues v
		WHERE qt.id = $1
		  AND qt.status = 'serving'
		  AND qt.service_id = s.id
		  AND s.venue_id = v.id
		  AND v.owner_id = $2
		RETURNING qt.id, qt.user_id, qt.venue_id, qt.service_id, qt.ticket_number, qt.status, qt.joined_at, qt.served_at, qt.cancelled_at, qt.created_at, qt.updated_at
	`
	ticket := &models.QueueTicket{}
	err := r.db.Pool.QueryRow(ctx, query, ticketID, ownerID).Scan(
		&ticket.ID, &ticket.UserID, &ticket.VenueID, &ticket.ServiceID,
		&ticket.TicketNumber, &ticket.Status, &ticket.JoinedAt,
		&ticket.ServedAt, &ticket.CancelledAt, &ticket.CreatedAt, &ticket.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrQueueTicketNotFound
		}
		return nil, err
	}
	return ticket, nil
}

func (r *queueTicketRepository) CancelByStaff(ctx context.Context, ownerID, ticketID int) (*models.QueueTicket, error) {
	query := `
		UPDATE queue_tickets qt
		SET status = 'cancelled', cancelled_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		FROM services s, venues v
		WHERE qt.id = $1
		  AND qt.status IN ('waiting', 'serving')
		  AND qt.service_id = s.id
		  AND s.venue_id = v.id
		  AND v.owner_id = $2
		RETURNING qt.id, qt.user_id, qt.venue_id, qt.service_id, qt.ticket_number, qt.status, qt.joined_at, qt.served_at, qt.cancelled_at, qt.created_at, qt.updated_at
	`
	ticket := &models.QueueTicket{}
	err := r.db.Pool.QueryRow(ctx, query, ticketID, ownerID).Scan(
		&ticket.ID, &ticket.UserID, &ticket.VenueID, &ticket.ServiceID,
		&ticket.TicketNumber, &ticket.Status, &ticket.JoinedAt,
		&ticket.ServedAt, &ticket.CancelledAt, &ticket.CreatedAt, &ticket.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrQueueTicketNotFound
		}
		return nil, err
	}
	return ticket, nil
}
