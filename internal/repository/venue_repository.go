package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"qback/internal/database"
	"qback/internal/models"
)

var (
	ErrVenueNotFound = errors.New("venue not found")
)

type VenueRepository interface {
	GetAll(ctx context.Context) ([]*models.Venue, error)
	GetByID(ctx context.Context, id int) (*models.Venue, error)
	Search(ctx context.Context, query string) ([]*models.Venue, error)
}

type venueRepository struct {
	db *database.DB
}

func NewVenueRepository(db *database.DB) VenueRepository {
	return &venueRepository{db: db}
}

func (r *venueRepository) GetAll(ctx context.Context) ([]*models.Venue, error) {
	query := `
		SELECT id, owner_id, name, description, address, created_at, updated_at
		FROM venues
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var venues []*models.Venue
	for rows.Next() {
		venue := &models.Venue{}
		err := rows.Scan(
			&venue.ID,
			&venue.OwnerID,
			&venue.Name,
			&venue.Description,
			&venue.Address,
			&venue.CreatedAt,
			&venue.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		venues = append(venues, venue)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return venues, nil
}

func (r *venueRepository) GetByID(ctx context.Context, id int) (*models.Venue, error) {
	query := `
		SELECT id, owner_id, name, description, address, created_at, updated_at
		FROM venues
		WHERE id = $1
	`
	venue := &models.Venue{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&venue.ID,
		&venue.OwnerID,
		&venue.Name,
		&venue.Description,
		&venue.Address,
		&venue.CreatedAt,
		&venue.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVenueNotFound
		}
		return nil, err
	}
	return venue, nil
}

func (r *venueRepository) Search(ctx context.Context, searchQuery string) ([]*models.Venue, error) {
	query := `
		SELECT id, owner_id, name, description, address, created_at, updated_at
		FROM venues
		WHERE name ILIKE '%' || $1 || '%' 
		   OR description ILIKE '%' || $1 || '%'
		   OR address ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, searchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var venues []*models.Venue
	for rows.Next() {
		venue := &models.Venue{}
		err := rows.Scan(
			&venue.ID,
			&venue.OwnerID,
			&venue.Name,
			&venue.Description,
			&venue.Address,
			&venue.CreatedAt,
			&venue.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		venues = append(venues, venue)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return venues, nil
}
