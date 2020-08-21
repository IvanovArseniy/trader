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
		openedOrders, err := postgresservice.GetOpenedOrders()
		if err != nil {
			fmt.Printf("Error occured %v", err)
			continue
		}
		ticket, err := binanceservice.GetTicket()
		if err != nil {
			fmt.Printf("Error occured %v\n", err)
			continue
		}
		level, err := postgresservice.GetLevel(ticket.Bid)
		if err != nil {
			fmt.Printf("Error occured %v\n", err)
			continue
		}
		if level == (orderer.Level{}) {
			closeOpenedSellOrders()
		}
		if len(openedOrders) == 0 {
			if level != (orderer.Level{}) {
				fmt.Printf("level ID=%v for bid %v\n", level.ID, ticket.Bid)
				order := orderer.Order{Price: level.BidFrom, Quantity: 0.001, Side: orderer.SellSide, ParentOrderID: 0, Status: orderer.OpenedOrder}
				orderID, createOrderErr := binanceservice.CreateOrder(order)
				if createOrderErr != nil {
					fmt.Printf("Error occured %v", createOrderErr)
				}
				order.ExternalID = orderID
				postgresservice.CreateOrder(order)
			} else {
				fmt.Printf("No level for bid %v\n", ticket.Bid)
			}
		} else if len(openedOrders) == 1 {
			order, err := binanceservice.GetOrder(openedOrders[0].ExternalID)
			if err != nil {
				fmt.Printf("Error occured: %v", err)
				continue

			}
			if order.Status != openedOrders[0].Status && order.Status == orderer.ClosedOrder {
				result, err := postgresservice.CloseOrder(openedOrders[0].ID)
				if err != nil {
					fmt.Printf("Error occured: %v", err)
					continue
				}
				if !result {
					fmt.Printf("Error occuder: Order do not closed in database")
					continue
				}
				if openedOrders[0].ParentOrderID == 0 && order.Status == orderer.ClosedOrder {
					order := orderer.Order{Price: (openedOrders[0].Price - 20), Quantity: 0.001, Side: orderer.BuySide, StopPrice: (openedOrders[0].Price + 30), StopPriceLimit: (openedOrders[0].Price + 25), ParentOrderID: openedOrders[0].ID, Status: orderer.OpenedOrder}
					orderID, err := binanceservice.CreateOcoOrder(order)
					if err != nil {
						fmt.Printf("Error occured: %v\n", err)
					}
					order.ExternalID = orderID
					postgresservice.CreateOrder(order)
				}
			}
		} else {
			closeOpenedSellOrders()
		}
	}
}

func closeOpenedSellOrders() {
	orderIDs, err := postgresservice.CloseOpenedSellOrders()
	if err != nil {
		return
	}
	for _, orderID := range orderIDs {
		binanceservice.CloseOrder(int64(orderID))
	}
}
