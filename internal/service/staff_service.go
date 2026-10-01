package service

import (
	"context"

	"qback/internal/models"
	"qback/internal/repository"
)

type StaffService interface {
	GetStaffQueues(ctx context.Context, ownerID int, venueID, serviceID *int, status *string) ([]*models.QueueTicket, error)
	CallNext(ctx context.Context, ownerID, serviceID int) (*models.QueueTicket, error)
	Serve(ctx context.Context, ownerID, ticketID int) (*models.QueueTicket, error)
	NoShow(ctx context.Context, ownerID, ticketID int) (*models.QueueTicket, error)
	Cancel(ctx context.Context, ownerID, ticketID int) (*models.QueueTicket, error)
}

type staffServiceImpl struct {
	queueRepo repository.QueueTicketRepository
}

func NewStaffService(queueRepo repository.QueueTicketRepository) StaffService {
	return &staffServiceImpl{
		queueRepo: queueRepo,
	}
}

func (s *staffServiceImpl) GetStaffQueues(ctx context.Context, ownerID int, venueID, serviceID *int, status *string) ([]*models.QueueTicket, error) {
	return s.queueRepo.GetStaffQueues(ctx, ownerID, venueID, serviceID, status)
}

func (s *staffServiceImpl) CallNext(ctx context.Context, ownerID, serviceID int) (*models.QueueTicket, error) {
	return s.queueRepo.CallNext(ctx, ownerID, serviceID)
}

func (s *staffServiceImpl) Serve(ctx context.Context, ownerID, ticketID int) (*models.QueueTicket, error) {
	return s.queueRepo.Serve(ctx, ownerID, ticketID)
}

func (s *staffServiceImpl) NoShow(ctx context.Context, ownerID, ticketID int) (*models.QueueTicket, error) {
	return s.queueRepo.NoShow(ctx, ownerID, ticketID)
}

func (s *staffServiceImpl) Cancel(ctx context.Context, ownerID, ticketID int) (*models.QueueTicket, error) {
	return s.queueRepo.CancelByStaff(ctx, ownerID, ticketID)
}
