package main

import (
	"fmt"
	"time"
	binanceservice "trader/btctracker/subpackages/binanceservice"
	postgresservice "trader/btctracker/subpackages/postgresservice"
)

func main() {
	fmt.Printf("Tracker started\n")
	for {
		time.Sleep(5 * time.Second)
		ticket, err := binanceservice.GetTicket()
		if err != nil {
			fmt.Printf("Error occured %v\n", err)
			continue
		}
		_, err = postgresservice.SaveTicket(ticket)
		if err != nil {
			fmt.Printf("Error occured %v\n", err)
		}
		err = postgresservice.CalculateCandle()
		if err != nil {
			fmt.Printf("Error occured %v\n", err)
		}
		err = postgresservice.CalculateLevel()
		if err != nil {
			fmt.Printf("Error occured %v\n", err)
		}

		fmt.Printf("ID: %v, Bid: %v, Ask: %v, CreatedOn: %v\n", ticket.ID, ticket.Bid, ticket.Ask, ticket.CreatedOn)
	}
}
