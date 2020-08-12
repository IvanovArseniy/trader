package binanceservice

import (
	"time"
	tracker "trader/btctracker/root"
)

//GetTicket requests ticket data from binance
func GetTicket() (ticket tracker.Ticket) {
	ticket = tracker.Ticket{ID: 10, Bid: 16.03, Ask: 15.02, CreatedOn: time.Now()}
	return
}
