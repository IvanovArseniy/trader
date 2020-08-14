package tracker

import "time"

// Ticket is element with ask and bid in some moment
type Ticket struct {
	ID        int64
	Bid       float32
	Ask       float32
	CreatedOn time.Time
}

//Level is support or resistance interval with start and end points
type Level struct {
	ID      int64
	BidFrom float32
	BidTo   float32
}
