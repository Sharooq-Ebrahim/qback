package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"qback/internal/models"
	"qback/internal/repository"
)

var (
	ErrUnauthorizedTicket   = errors.New("unauthorized to access this ticket")
	ErrTicketNotWaiting     = errors.New("ticket is not in waiting state")
	ErrServiceVenueMismatch = errors.New("service does not belong to the specified venue")
)

type TicketDetails struct {
	*models.QueueTicket
	Position          int `json:"position"`
	PeopleAhead       int `json:"people_ahead"`
	EstimatedWaitMins int `json:"estimated_wait_mins"`
}

type QueueService interface {
	JoinQueue(ctx context.Context, userID, venueID, serviceID int) (*TicketDetails, error)
	GetTicketDetails(ctx context.Context, ticketID, userID int) (*TicketDetails, error)
	GetActiveTickets(ctx context.Context, userID int) ([]*TicketDetails, error)
	GetTicketHistory(ctx context.Context, userID int) ([]*models.QueueTicket, error)
	CancelTicket(ctx context.Context, ticketID, userID int) error
}

type queueServiceImpl struct {
	queueRepo   repository.QueueTicketRepository
	serviceRepo repository.ServiceRepository
}

func NewQueueService(queueRepo repository.QueueTicketRepository, serviceRepo repository.ServiceRepository) QueueService {
	return &queueServiceImpl{
		queueRepo:   queueRepo,
		serviceRepo: serviceRepo,
	}
}

func generateTicketNumber() string {
	b := make([]byte, 3)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *queueServiceImpl) calculateQueueStats(ctx context.Context, ticket *models.QueueTicket, svc *models.Service) (int, int, int, error) {
	if ticket.Status != "waiting" {
		return 0, 0, 0, nil
	}

	waitingTickets, err := s.queueRepo.GetWaitingByServiceID(ctx, ticket.ServiceID)
	if err != nil {
		return 0, 0, 0, err
	}

	position := 0
	peopleAhead := 0

	for i, t := range waitingTickets {
		if t.ID == ticket.ID {
			position = i + 1
			peopleAhead = i
			break
		}
	}

	if position == 0 {
		return 0, 0, 0, nil
	}

	estWait := peopleAhead * svc.DurationMinutes

	return position, peopleAhead, estWait, nil
}

func (s *queueServiceImpl) JoinQueue(ctx context.Context, userID, venueID, serviceID int) (*TicketDetails, error) {
	svc, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	if svc.VenueID != venueID {
		return nil, ErrServiceVenueMismatch
	}

	ticket := &models.QueueTicket{
		UserID:       userID,
		VenueID:      venueID,
		ServiceID:    serviceID,
		TicketNumber: generateTicketNumber(),
		Status:       "waiting",
	}

	if err := s.queueRepo.Create(ctx, ticket); err != nil {
		return nil, err
	}

	position, peopleAhead, estWait, err := s.calculateQueueStats(ctx, ticket, svc)
	if err != nil {
		return nil, err
	}

	return &TicketDetails{
		QueueTicket:       ticket,
		Position:          position,
		PeopleAhead:       peopleAhead,
		EstimatedWaitMins: estWait,
	}, nil
}

func (s *queueServiceImpl) GetTicketDetails(ctx context.Context, ticketID, userID int) (*TicketDetails, error) {
	ticket, err := s.queueRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.UserID != userID {
		return nil, ErrUnauthorizedTicket
	}

	svc, err := s.serviceRepo.GetByID(ctx, ticket.ServiceID)
	if err != nil {
		return nil, err
	}

	position, peopleAhead, estWait, err := s.calculateQueueStats(ctx, ticket, svc)
	if err != nil {
		return nil, err
	}

	return &TicketDetails{
		QueueTicket:       ticket,
		Position:          position,
		PeopleAhead:       peopleAhead,
		EstimatedWaitMins: estWait,
	}, nil
}

func (s *queueServiceImpl) GetActiveTickets(ctx context.Context, userID int) ([]*TicketDetails, error) {
	tickets, err := s.queueRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var details []*TicketDetails
	for _, t := range tickets {
		svc, err := s.serviceRepo.GetByID(ctx, t.ServiceID)
		if err != nil {
			return nil, err
		}

		pos, ahead, estWait, err := s.calculateQueueStats(ctx, t, svc)
		if err != nil {
			return nil, err
		}

		details = append(details, &TicketDetails{
			QueueTicket:       t,
			Position:          pos,
			PeopleAhead:       ahead,
			EstimatedWaitMins: estWait,
		})
	}

	return details, nil
}

func (s *queueServiceImpl) GetTicketHistory(ctx context.Context, userID int) ([]*models.QueueTicket, error) {
	return s.queueRepo.GetHistoryByUserID(ctx, userID)
}

func (s *queueServiceImpl) CancelTicket(ctx context.Context, ticketID, userID int) error {
	ticket, err := s.queueRepo.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}

	if ticket.UserID != userID {
		return ErrUnauthorizedTicket
	}

	if ticket.Status != "waiting" {
		return ErrTicketNotWaiting
	}

	return s.queueRepo.Cancel(ctx, ticketID)
}
