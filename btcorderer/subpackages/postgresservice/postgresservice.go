package postgresservice

import (
	"database/sql"
	orderer "trader/btcorderer/root"

	_ "github.com/lib/pq"
)

//GetLevel try to find nearby levels for bid
func GetLevel(bid float32) (level orderer.Level, funcErr error) {
	connStr := "user=postgres password=postgres dbname=btcanalytics sslmode=disable"
	db, openErr := sql.Open("postgres", connStr)
	if openErr != nil {
		funcErr = openErr
	}
	defer db.Close()

	rows, queryErr := db.Query("select \"Id\", \"bidfrom\", \"bidto\" from \"Level\" where \"bidfrom\" > $1 and (\"bidfrom\" - 20) < $1", bid)
	if queryErr != nil {
		funcErr = queryErr
		return
	}
	defer rows.Close()

	for rows.Next() {
		scanErr := rows.Scan(&level.ID, &level.BidFrom, &level.BidTo)
		if scanErr != nil {
			funcErr = scanErr
			continue
		}
	}
	return
}
