package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	analyticsapi "trader/btcanalyticsapi/root"
	"trader/btcanalyticsapi/subpackages/binanceservice"
	"trader/btcanalyticsapi/subpackages/postgresservice"
)

//CancelOpenedOrders is hadler for cancel opened orders operation
func CancelOpenedOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := binanceservice.GetOpenedOrders()
	if err != nil {
		panic(err)
	}
	operationResult := analyticsapi.OperationResult{}
	for _, o := range orders {
		r, _ := binanceservice.CloseOrder(o.ExternalID)
		operationResult.Result = operationResult.Result && r
	}
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	json.NewEncoder(w).Encode(operationResult)
}

//Index is handler for index page
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to analytics API!")
}

//GetCandles is handler for getting candles
func GetCandles(w http.ResponseWriter, r *http.Request) {
	candles, err := postgresservice.GetCandles()
	if err != nil {
		{
			panic("Cant get candles")
		}
	}
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	json.NewEncoder(w).Encode(candles)
}

//GetLevels is handler for getting levels
func GetLevels(w http.ResponseWriter, r *http.Request) {
	levels, err := postgresservice.GetLevels()
	if err != nil {
		panic("Cant get levels")
	}
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	json.NewEncoder(w).Encode(levels)
}

//GetOrders is handler for getting orders
func GetOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := binanceservice.GetOpenedOrders()
	if err != nil {
		panic(err)
	}
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	json.NewEncoder(w).Encode(orders)
}

//GetTicket is handler for getting tickets
func GetTicket(w http.ResponseWriter, r *http.Request) {
	ticket, err := binanceservice.GetTicket()
	if err != nil {
		panic(err)
	}
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	json.NewEncoder(w).Encode(ticket)
}
