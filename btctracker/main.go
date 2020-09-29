package main

import (
	"fmt"
	"time"
	tracker "trader/btctracker/root"
	binanceservice "trader/btctracker/subpackages/binanceservice"
	postgresservice "trader/btctracker/subpackages/postgresservice"
)

func main() {
	fmt.Printf("Tracker started\n")
	ticketChannel := make(chan tracker.Ticket)

	i := 0
	j := 0
	for {
		i++
		j++
		time.Sleep(5 * time.Second)
		go func() {
			ticket, err := binanceservice.GetTicket()
			if err != nil {
				fmt.Printf("Error occured %v\n", err)
			}
			fmt.Printf("ID: %v, Bid: %v, Ask: %v, CreatedOn: %v\n", ticket.ID, ticket.Bid, ticket.Ask, ticket.CreatedOn)
			ticketChannel <- ticket
		}()

		go func() {
			ticket := <-ticketChannel
			_, err := postgresservice.SaveTicket(ticket)
			if err != nil {
				fmt.Printf("Error occured %v\n", err)
			}
		}()

		if i == 12 {
			i = 0
			go func() {
				err := postgresservice.CalculateCandle()
				if err != nil {
					fmt.Printf("Error occured %v\n", err)
				}
			}()
		}
		if j == 720 {
			j = 0
			go func() {
				err := postgresservice.CalculateLevel()
				if err != nil {
					fmt.Printf("Error occured %v\n", err)
				}
			}()
		}

	}
}
