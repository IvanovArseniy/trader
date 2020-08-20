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

	// order := orderer.Order{Price: 30000.02, Quantity: 0.001, Side: orderer.SellSide}
	// orderID, _ := binanceservice.CreateOrder(order)
	// binanceservice.CloseOrder(orderID)

	for {
		time.Sleep(10 * time.Second)
		openedOrders, err := postgresservice.GetOpenedOrders()
		if err != nil {
			fmt.Printf("Error occured %v", err)
		}
		if len(openedOrders) == 0 {
			ticket, err := binanceservice.GetTicket()
			level, err := postgresservice.GetLevel(ticket.Bid)
			if err != nil {
				fmt.Printf("Error occured %v\n", err)
			} else if level != (orderer.Level{}) {
				fmt.Printf("level ID=%v for bid %v\n", level.ID, ticket.Bid)
				order := orderer.Order{}
				orderID, createOrderErr := binanceservice.CreateOrder(order)
				if createOrderErr != nil {
					fmt.Printf("Error occured %v", createOrderErr)
				}
				order.ID = orderID
				postgresservice.CreateOrder(order)
			} else {
				fmt.Printf("No level for bid %v\n", ticket.Bid)
			}
		} else if len(openedOrders) == 1 {
			order, err := binanceservice.GetOrder(openedOrders[0].ID)
			if err != nil {
				fmt.Printf("Error occured: %v", err)
				continue

			}
			if order.Status != openedOrders[0].Status && order.Status == orderer.ClosedOrder {
				result, err := postgresservice.CloseOrder(order.ID)
				if err != nil {
					fmt.Printf("Error occured: %v", err)
					continue
				}
				if !result {
					fmt.Printf("Error occuder: Order do not closed in daatabase")
					continue
				}
				if openedOrders[0].ParentOrderID == 0 {
					order := orderer.Order{}
					orderID, err := binanceservice.CreateOcoOrder(order)
					if err != nil {
						fmt.Printf("Error occured: %v\n", err)
					}
					order.ID = orderID
					postgresservice.CreateOrder(order)
				}
			}
		} else {
			closeOpenedSellOrders()
		}
		defer closeOpenedSellOrders()
	}
}

func closeOpenedSellOrders() {
	orderIDs, err := postgresservice.CloseOpenedSellOrders()
	if err != nil {
		return
	}
	for orderID := range orderIDs {
		binanceservice.CloseOrder(int64(orderID))
	}
}
