package postgresservice

import (
	"database/sql"
	analitycsapi "trader/btcanalyticsapi/root"
	orderer "trader/btcorderer/root"

	//Database import
	_ "github.com/lib/pq"
	"github.com/tkanos/gonfig"
)

//GetLevels get all active levels from database
func GetLevels() (levels analitycsapi.Levels, err error) {
	configuration := analitycsapi.Configuration{}
	err = gonfig.GetConf("config/config.json", &configuration)
	if err != nil {
		return
	}

	db, err := sql.Open("postgres", configuration.PostgresConnectionString)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("select \"Id\", \"bidfrom\", \"bidto\" from \"Level\" where \"active\" = 1 and \"deleted\" = 0")
	if err != nil {
		return
	}
	defer rows.Close()

	levels = analitycsapi.Levels{}
	for rows.Next() {
		level := analitycsapi.Level{}
		err := rows.Scan(&level.ID, &level.BidFrom, &level.BidTo)
		if err != nil {
			continue
		}
		levels = append(levels, level)
	}
	return
}

//GetCandles get all candles from databases
func GetCandles() (candles analitycsapi.Candles, err error) {
	configuration := analitycsapi.Configuration{}
	err = gonfig.GetConf("config/config.json", &configuration)
	if err != nil {
		return
	}

	db, err := sql.Open("postgres", configuration.PostgresConnectionString)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("select \"startbid\", \"minbid\", \"maxbid\", \"endbid\" from \"Candle\" order by \"Id\" asc")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		candle := analitycsapi.Candle{}
		err := rows.Scan(&candle.StartBid, &candle.MinBid, &candle.MaxBid, &candle.EndBid)
		if err != nil {
			continue
		}
		candles = append(candles, candle)
	}
	return
}

//GetOpenedOrders function returns all opened orders from database
func GetOpenedOrders() (orders []orderer.Order, err error) {
	configuration := orderer.Configuration{}
	err = gonfig.GetConf("config/config.json", &configuration)
	if err != nil {
		return
	}

	db, err := sql.Open("postgres", configuration.PostgresConnectionString)
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("select \"Id\", \"price\" from \"Order\" where \"status\" = 1")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		order := orderer.Order{}
		scanErr := rows.Scan(&order.ID, &order.Price)
		if scanErr != nil {
			err = scanErr
			continue
		}
		orders = append(orders, order)
	}
	return
}
