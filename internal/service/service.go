package service

import (
	"context"
	"errors"
	"strings"

	"qback/internal/models"
	"qback/internal/repository"
)

var (
	ErrInvalidID        = errors.New("invalid service id")
	ErrEmptySearchQuery = errors.New("search query cannot be empty")
)

// Service defines the business logic methods for services.
type Service interface {
	GetAll(ctx context.Context) ([]*models.Service, error)
	GetByID(ctx context.Context, id int) (*models.Service, error)
	Search(ctx context.Context, query string) ([]*models.Service, error)
	GetByVenueID(ctx context.Context, venueID int) ([]*models.Service, error)
}

type serviceImpl struct {
	repo repository.ServiceRepository
}

func NewService(repo repository.ServiceRepository) Service {
	return &serviceImpl{
		repo: repo,
	}
}

func (s *serviceImpl) GetAll(ctx context.Context) ([]*models.Service, error) {
	return s.repo.GetAll(ctx)
}

func (s *serviceImpl) GetByID(ctx context.Context, id int) (*models.Service, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	return s.repo.GetByID(ctx, id)
}

func (s *serviceImpl) Search(ctx context.Context, query string) ([]*models.Service, error) {
	trimmedQuery := strings.TrimSpace(query)
	if len(trimmedQuery) == 0 {
		return nil, ErrEmptySearchQuery
	}

	return s.repo.Search(ctx, trimmedQuery)
}

func (s *serviceImpl) GetByVenueID(ctx context.Context, venueID int) ([]*models.Service, error) {
	if venueID <= 0 {
		return nil, ErrInvalidID
	}
	return s.repo.GetByVenueID(ctx, venueID)
}
