package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"qback/internal/database"
	"qback/internal/models"
)

var (
	ErrServiceNotFound = errors.New("service not found")
)

type ServiceRepository interface {
	GetAll(ctx context.Context) ([]*models.Service, error)
	GetByID(ctx context.Context, id int) (*models.Service, error)
	Search(ctx context.Context, query string) ([]*models.Service, error)
}

type serviceRepository struct {
	db *database.DB
}

func NewServiceRepository(db *database.DB) ServiceRepository {
	return &serviceRepository{db: db}
}

func (r *serviceRepository) GetAll(ctx context.Context) ([]*models.Service, error) {
	query := `
		SELECT id, venue_id, name, description, duration_minutes, price, created_at, updated_at
		FROM services
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*models.Service
	for rows.Next() {
		service := &models.Service{}
		err := rows.Scan(
			&service.ID,
			&service.VenueID,
			&service.Name,
			&service.Description,
			&service.DurationMinutes,
			&service.Price,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (r *serviceRepository) GetByID(ctx context.Context, id int) (*models.Service, error) {
	query := `
		SELECT id, venue_id, name, description, duration_minutes, price, created_at, updated_at
		FROM services
		WHERE id = $1
	`
	service := &models.Service{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&service.ID,
		&service.VenueID,
		&service.Name,
		&service.Description,
		&service.DurationMinutes,
		&service.Price,
		&service.CreatedAt,
		&service.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}
	return service, nil
}

func (r *serviceRepository) Search(ctx context.Context, searchQuery string) ([]*models.Service, error) {
	query := `
		SELECT id, venue_id, name, description, duration_minutes, price, created_at, updated_at
		FROM services
		WHERE name ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, searchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*models.Service
	for rows.Next() {
		service := &models.Service{}
		err := rows.Scan(
			&service.ID,
			&service.VenueID,
			&service.Name,
			&service.Description,
			&service.DurationMinutes,
			&service.Price,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}
