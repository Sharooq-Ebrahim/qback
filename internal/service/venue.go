package service

import (
	"context"
	"errors"
	"strings"

	"qback/internal/models"
	"qback/internal/repository"
)

var (
	ErrInvalidVenueID = errors.New("invalid venue id")
)

type VenueService interface {
	GetAll(ctx context.Context) ([]*models.Venue, error)
	GetByID(ctx context.Context, id int) (*models.Venue, error)
	Search(ctx context.Context, query string) ([]*models.Venue, error)
}

type venueServiceImpl struct {
	repo repository.VenueRepository
}

func NewVenueService(repo repository.VenueRepository) VenueService {
	return &venueServiceImpl{
		repo: repo,
	}
}

func (s *venueServiceImpl) GetAll(ctx context.Context) ([]*models.Venue, error) {
	return s.repo.GetAll(ctx)
}

func (s *venueServiceImpl) GetByID(ctx context.Context, id int) (*models.Venue, error) {
	if id <= 0 {
		return nil, ErrInvalidVenueID
	}

	return s.repo.GetByID(ctx, id)
}

func (s *venueServiceImpl) Search(ctx context.Context, query string) ([]*models.Venue, error) {
	trimmedQuery := strings.TrimSpace(query)
	if len(trimmedQuery) == 0 {
		return nil, ErrEmptySearchQuery
	}

	return s.repo.Search(ctx, trimmedQuery)
}
