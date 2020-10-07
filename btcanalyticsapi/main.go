package main

import (
	"log"
	"net/http"
	handler "trader/btcanalyticsapi/subpackages/handlers"

	"github.com/gorilla/mux"
)

func main() {
	handleRequest()
}

func handleRequest() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", handler.Index)
	router.HandleFunc("/candles/", handler.GetCandles)
	router.HandleFunc("/levels/", handler.GetLevels)
	router.HandleFunc("/ticket/", handler.GetTicket)
	router.HandleFunc("/orders/", handler.GetOrders).Methods("GET")
	router.HandleFunc("/orders/", handler.CancelOpenedOrders).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", router))
}
