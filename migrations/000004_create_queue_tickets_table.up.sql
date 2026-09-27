CREATE TABLE IF NOT EXISTS queue_tickets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    venue_id INTEGER NOT NULL REFERENCES venues(id) ON DELETE CASCADE,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    ticket_number VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'waiting',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    served_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_status CHECK (status IN ('waiting', 'serving', 'served', 'cancelled', 'no_show'))
);


CREATE INDEX IF NOT EXISTS idx_queue_tickets_venue_status_joined ON queue_tickets(venue_id, status, joined_at);
CREATE INDEX IF NOT EXISTS idx_queue_tickets_user_id_status ON queue_tickets(user_id, status);
