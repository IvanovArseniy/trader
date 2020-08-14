package main

import (
	"fmt"
	"time"
	orderer "trader/btcorderer/root"
	"trader/btcorderer/subpackages/binanceservice"
	"trader/btcorderer/subpackages/postgresservice"
)

func main() {
	fmt.Printf("Orderer started\n")
	for {
		time.Sleep(10 * time.Second)
		makeOrder := false
		ticket, err := binanceservice.GetTicket()
		// ticket := orderer.Ticket{ID: 200, Bid: 11750, Ask: 11751}
		level, err := postgresservice.GetLevel(ticket.Bid)
		if err != nil {
			fmt.Printf("Error occured %v\n", err)
		} else if level != (orderer.Level{}) {
			makeOrder = true
		}
		if makeOrder {
			fmt.Printf("level ID=%v for bid %v\n", level.ID, ticket.Bid)
			//make order
			//save opened order, close order and stoploss order to DB
		} else {
			fmt.Printf("No level for bid %v\n", ticket.Bid)
		}
	}
}
