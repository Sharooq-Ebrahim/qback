package sse

import (
	"encoding/json"
	"sync"

	"qback/internal/models"
)

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type subscriber struct {
	ownerID   int
	serviceID *int
	ch        chan Event
}

type Broker struct {
	mu          sync.RWMutex
	subscribers map[int64]*subscriber
	nextID      int64
}

func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[int64]*subscriber),
	}
}

func (b *Broker) Subscribe(ownerID int, serviceID *int) (<-chan Event, func()) {
	ch := make(chan Event, 16)

	b.mu.Lock()
	id := b.nextID
	b.nextID++
	b.subscribers[id] = &subscriber{
		ownerID:   ownerID,
		serviceID: serviceID,
		ch:        ch,
	}
	b.mu.Unlock()

	unsub := func() {
		b.mu.Lock()
		delete(b.subscribers, id)
		b.mu.Unlock()
		close(ch)
	}

	return ch, unsub
}

func (b *Broker) PublishTicketUpdate(ownerID int, ticket *models.QueueTicket) {
	payload, err := json.Marshal(ticket)
	if err != nil {
		return
	}

	event := Event{
		Type: "ticket_update",
		Data: json.RawMessage(payload),
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, sub := range b.subscribers {
		if sub.ownerID != ownerID {
			continue
		}
		if sub.serviceID != nil && *sub.serviceID != ticket.ServiceID {
			continue
		}
		select {
		case sub.ch <- event:
		default:
		}
	}
}
