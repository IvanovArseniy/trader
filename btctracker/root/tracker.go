package tracker

import "time"

// Ticket is element with ask and bid in some moment
type Ticket struct {
	ID        int
	Bid       float32
	Ask       float32
	CreatedOn time.Time
}
