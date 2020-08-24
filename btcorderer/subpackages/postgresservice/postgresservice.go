package postgresservice

import (
	"database/sql"
	orderer "trader/btcorderer/root"

	//Database import
	_ "github.com/lib/pq"
	"github.com/tkanos/gonfig"
)

type growth struct {
	Price float64
}

//GetLevel try to find nearby levels for bid
func GetLevel(bid float64) (level orderer.Level, err error) {
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

	rows, err := db.Query("select \"Id\", \"bidfrom\", \"bidto\" from \"Level\" where (\"bidto\") > $1 and (\"bidfrom\" - 50) < $1 and \"active\" = 1 and \"deleted\" = 0", bid)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&level.ID, &level.BidFrom, &level.BidTo)
		if err != nil {
			continue
		}
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

	rows, err := db.Query("select \"Id\", \"parentId\", \"price\", \"quantity\", \"status\", \"side\", \"externalid\", \"buyprice\" from \"Order\" where \"status\" = 1")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		order := orderer.Order{}
		scanErr := rows.Scan(&order.ID, &order.ParentOrderID, &order.Price, &order.Quantity, &order.Status, &order.Side, &order.ExternalID, &order.BuyPrice)
		if scanErr != nil {
			err = scanErr
			continue
		}
		orders = append(orders, order)
	}
	return
}

//CreateOrder function saves order to database
func CreateOrder(order orderer.Order) (err error) {
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

	_, err = db.Exec("insert into \"Order\" (\"parentId\", \"price\", \"quantity\", \"status\",\"side\", \"externalid\", \"buyprice\") values ($1, $2, $3, $4, $5, $6, $7)", order.ParentOrderID, order.Price, order.Quantity, order.Status, order.Side, order.ExternalID, order.BuyPrice)
	if err != nil {
		return
	}
	return
}

//CloseOpenedSellOrders function close all opened sell orders
func CloseOpenedSellOrders() (orderIDs []int64, err error) {
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

	rows, err := db.Query("select \"externalid\" from \"Order\" where \"side\" = 1 and \"status\" = 1")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		order := orderer.Order{}
		err = rows.Scan(&order.ID)
		if err != nil {
			continue
		}
		orderIDs = append(orderIDs, order.ID)
	}

	res, err := db.Exec("delete from \"Order\" where \"side\" = 1 and \"status\" = 1")
	if err != nil {
		return
	}
	_, err = res.RowsAffected()
	return
}

//GetOrder funciton get order by ID from database
func GetOrder(orderID int64) (order orderer.Order, err error) {
	return
}

//CloseOrder function updates order status to CLOSED
func CloseOrder(orderID int64) (result bool, err error) {
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

	res, err := db.Exec("Update \"Order\" set \"status\"=2 where \"Id\"=$1", orderID)
	if err != nil {
		return
	}

	affRows, err := res.RowsAffected()
	if err != nil {
		return
	}
	result = affRows > 0
	return
}

//GetPriceGrowth get price growth for last 14 hours
func GetPriceGrowth() (priceGrowth float64, err error) {
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

	rows, err := db.Query("select max(\"endbid\") - min(\"startbid\") as \"priceGrowth\" from (select \"startbid\", \"endbid\" from \"Candle\" order by \"Id\" desc limit 14) t")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		growth := growth{}
		err = rows.Scan(&growth.Price)
		if err != nil {
			continue
		}
		priceGrowth = growth.Price
		break
	}
	return
}
