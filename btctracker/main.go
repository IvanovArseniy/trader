package main

import (
	"fmt"
	binanceservice "trader/btctracker/subpackages/binanceservice"
	postgresservice "trader/btctracker/subpackages/postgresservice"
)

func main() {
	fmt.Printf("Trader started\n")
	ticket := binanceservice.GetTicket()
	postgresservice.SaveTicket(ticket)
	fmt.Printf("ID: %v, Bid: %v, Ask: %v, CreatedOn: %v", ticket.ID, ticket.Bid, ticket.Ask, ticket.CreatedOn)
}
