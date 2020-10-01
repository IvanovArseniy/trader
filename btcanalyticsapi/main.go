package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"trader/btcanalyticsapi/subpackages/binanceservice"
	"trader/btcanalyticsapi/subpackages/postgresservice"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", index)
	router.HandleFunc("/candles/", getCandles)
	router.HandleFunc("/levels/", getLevels)
	router.HandleFunc("/orders/", getOrders)
	router.HandleFunc("/ticket/", getTicket)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to analytics API!")
}

func getCandles(w http.ResponseWriter, r *http.Request) {
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

func getLevels(w http.ResponseWriter, r *http.Request) {
	levels, err := postgresservice.GetLevels()
	if err != nil {
		panic("Cant get levels")
	}
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	json.NewEncoder(w).Encode(levels)
}

func getOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := binanceservice.GetOpenedOrders()
	if err != nil {
		panic(err)
	}
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	json.NewEncoder(w).Encode(orders)
}

func getTicket(w http.ResponseWriter, r *http.Request) {
	ticket, err := binanceservice.GetTicket()
	if err != nil {
		panic(err)
	}
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	json.NewEncoder(w).Encode(ticket)
}
