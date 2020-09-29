package postgresservice

import (
	"database/sql"
	tracker "trader/btctracker/root"

	_ "github.com/lib/pq"
)

// SaveTicket saves ticket data to database
func SaveTicket(ticket tracker.Ticket) (ticketID int64, funcErr error) {
	connStr := "user=postgres password=postgres dbname=btcanalytics sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		funcErr = err
		return
	}
	defer db.Close()

	result, err := db.Exec("insert into \"RateHistory\" (bid, ask, createdOn) values ($1, $2,  $3)", ticket.Bid, ticket.Ask, ticket.CreatedOn)
	if err != nil {
		funcErr = err
		return
	}
	ticketID, err = result.LastInsertId()
	return
}

// CalculateCandle call stored procedure to calculate candles
func CalculateCandle() (funcErr error) {
	connStr := "user=postgres password=postgres dbname=btcanalytics sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		funcErr = err
		return
	}
	defer db.Close()

	_, err = db.Exec("call \"CalculateCandle\"()")
	if err != nil {
		funcErr = err
		return
	}
	return
}

// CalculateLevel call storep procedure to calculate levels
func CalculateLevel() (funcErr error) {
	connStr := "user=postgres password=postgres dbname=btcanalytics sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		funcErr = err
		return
	}
	defer db.Close()

	_, err = db.Exec("call \"CalculateAnalyticsLevel\"()")
	if err != nil {
		funcErr = err
		return
	}
	return
}
