package main

import (
	"fmt"
	"time"
	binanceservice "trader/btctracker/subpackages/binanceservice"
	postgresservice "trader/btctracker/subpackages/postgresservice"
)

func main() {
	fmt.Printf("Trader started\n")
	for {
		ticket := binanceservice.GetTicket()
		postgresservice.SaveTicket(ticket)
		fmt.Printf("ID: %v, Bid: %v, Ask: %v, CreatedOn: %v\n", ticket.ID, ticket.Bid, ticket.Ask, ticket.CreatedOn)
		time.Sleep(5 * time.Second)
	}
}
