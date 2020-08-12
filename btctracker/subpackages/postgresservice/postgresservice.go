package postgresservice

import (
	"database/sql"
	tracker "trader/btctracker/root"

	_ "github.com/lib/pq"
)

// SaveTicket saves ticket data to database
func SaveTicket(ticket tracker.Ticket) (ticketID int64) {
	connStr := "user=postgres password=postgres dbname=btcanalytics sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	result, err := db.Exec("insert into \"RateHistory\" (bid, ask, createdOn) values ($1, $2,  $3)", ticket.Bid, ticket.Ask, ticket.CreatedOn)
	if err != nil {
		panic(err)
	}
	ticketID, err = result.LastInsertId()
	return
}
