package main

import (
	"fmt"
	"log"
	"os"
	"time"
	orderer "trader/btcorderer/root"
	"trader/btcorderer/subpackages/binanceservice"
	"trader/btcorderer/subpackages/postgresservice"
)

func main() {
	fmt.Printf("Orderer started\n")

	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	for {
		log.Println("------------------------------------------------")
		fmt.Printf("------------------------------------------------\n")
		time.Sleep(10 * time.Second)
		openedOrders, err := postgresservice.GetOpenedOrders()
		if err != nil {
			log.Println(fmt.Sprintf("Error occured %v", err))
			fmt.Printf("Error occured %v\n", err)
			continue
		}
		log.Println(fmt.Sprintf("There are %v opened orders", len(openedOrders)))
		fmt.Printf("There are %v opened orders\n", len(openedOrders))

		ticket, err := binanceservice.GetTicket()
		if err != nil {
			log.Println(fmt.Sprintf("Error occured %v", err))
			fmt.Printf("Error occured %v\n", err)
			continue
		}
		log.Println(fmt.Sprintf("Current ticket is bid:%f, ask:%f", ticket.Bid, ticket.Ask))
		fmt.Printf("Current ticket is bid:%f, ask:%f\n", ticket.Bid, ticket.Ask)

		level, err := postgresservice.GetLevel(ticket.Bid)
		if err != nil {
			log.Println(fmt.Sprintf("Error occured %v", err))
			fmt.Printf("Error occured %v\n", err)
			continue
		}
		if level != (orderer.Level{}) {
			log.Println(fmt.Sprintf("Level from %f to %f was found", level.BidFrom, level.BidTo))
			fmt.Printf("Level from %f to %f was found\n", level.BidFrom, level.BidTo)
		} else {
			log.Println(fmt.Sprintf("Cant find level for this ticket"))
			fmt.Printf("Cant find level for this ticket\n")
		}

		if level == (orderer.Level{}) {
			log.Println(fmt.Sprintf("Level wasnt found, close opened sell orders"))
			fmt.Printf("Level wasnt found, close opened sell orders\n")
			closeOpenedSellOrders()
		}
		if len(openedOrders) == 0 {
			if level != (orderer.Level{}) {
				order := orderer.Order{Price: (level.BidFrom - 30), Quantity: 0.001, Side: orderer.SellSide, ParentOrderID: 0, Status: orderer.OpenedOrder, BuyPrice: level.BidTo}
				log.Println(fmt.Sprintf("Level was found, create order price:%f quantity%f", order.Price, order.Quantity))
				fmt.Printf("Level was found, create order price:%f quantity%f\n", order.Price, order.Quantity)
				orderID, createOrderErr := binanceservice.CreateOrder(order)
				if createOrderErr != nil {
					log.Println(fmt.Sprintf("Error occured %v", err))
					fmt.Printf("Error occured %v\n", createOrderErr)
					continue
				}
				log.Println(fmt.Sprintf("Order was created on binance, id=%v", orderID))
				fmt.Printf("Order was created on binance, id=%v\n", orderID)
				order.ExternalID = orderID
				err = postgresservice.CreateOrder(order)
				if err != nil {
					log.Println(fmt.Sprintf("Error occured %v", err))
					fmt.Printf("Error occured %v\n", err)
					continue
				}
				log.Println(fmt.Sprintf("Order was created in database"))
				fmt.Printf("Order was created in database\n")
			}
		} else if len(openedOrders) == 1 {
			log.Println(fmt.Sprintf("Get info from binance for order binanceid=%v", openedOrders[0].ExternalID))
			fmt.Printf("Get info from binance for order binanceid=%v\n", openedOrders[0].ExternalID)
			order, err := binanceservice.GetOrder(openedOrders[0].ExternalID)
			if err != nil {
				log.Println(fmt.Sprintf("Error occured %v", err))
				fmt.Printf("Error occured %v\n", err)
				continue
			}
			log.Println(fmt.Sprintf("Opened order was found, binanceID=%v status=%v", order.ExternalID, order.Status))
			fmt.Printf("Opened order was found, binanceID=%v status=%v\n", order.ExternalID, order.Status)
			if order.Status != openedOrders[0].Status && order.Status == orderer.CanceledOrder {
				result, err := postgresservice.CloseOrder(openedOrders[0].ID)
				if err != nil {
					log.Println(fmt.Sprintf("Error occured %v", err))
					fmt.Printf("Error occured %v\n", err)
					continue
				}
				if !result {
					log.Println("Error occuder Order do not closed in database")
					fmt.Printf("Error occuder Order do not closed in database\n")
					continue
				}
				log.Println(fmt.Sprintf("Opened order binanceid=%v was cancelled", order.ExternalID))
				fmt.Printf("Opened order binanceid=%v was cancelled\n", order.ExternalID)
			}
			if order.Status != openedOrders[0].Status && order.Status == orderer.ClosedOrder {
				result, err := postgresservice.CloseOrder(openedOrders[0].ID)
				if err != nil {
					log.Println(fmt.Sprintf("Error occured %v", err))
					fmt.Printf("Error occured %v\n", err)
					continue
				}
				if !result {
					log.Println("Error occuder Order do not closed in database")
					fmt.Printf("Error occuder Order do not closed in database\n")
				} else {
					log.Println(fmt.Sprintf("Opened order binanceid=%v was closed", order.ExternalID))
					fmt.Printf("Opened order binanceid=%v was closed\n", order.ExternalID)
				}
				if openedOrders[0].ParentOrderID == 0 {
					stopPriceLimit := (openedOrders[0].BuyPrice + 30)
					s := int64(stopPriceLimit) % 100
					if s > 95 {
						stopPriceLimit = float64(int64(stopPriceLimit)/100*100 + 103)
					}

					buyPrice := (openedOrders[0].Price - 100)
					b := int64(buyPrice) % 100
					if b > 95 {
						buyPrice = float64(int64(buyPrice)/100*100 + 103)
					}
					order := orderer.Order{Price: buyPrice, Quantity: 0.001, Side: orderer.BuySide, StopPrice: (stopPriceLimit - 10), StopPriceLimit: stopPriceLimit, ParentOrderID: openedOrders[0].ID, Status: orderer.OpenedOrder}
					log.Println(fmt.Sprintf("It was an order to sell BTC. Create OCO order price:%f quantity:%f stopPrice:%f stopPriceLimit:%f", order.Price, order.Quantity, order.StopPrice, order.StopPriceLimit))
					fmt.Printf("It was an order to sell BTC. Create OCO order price:%f quantity:%f stopPrice:%f stopPriceLimit:%f\n", order.Price, order.Quantity, order.StopPrice, order.StopPriceLimit)
					orderID, err := binanceservice.CreateOcoOrder(order)
					if err != nil {
						log.Println(fmt.Sprintf("Error occured %v", err))
						fmt.Printf("Error occured %v\n", err)
						continue
					}
					log.Println(fmt.Sprintf("OCO order was created, binanceid=%v", orderID))
					fmt.Printf("OCO order was created, binanceid=%v\n", orderID)
					order.ExternalID = orderID
					err = postgresservice.CreateOrder(order)
					if err != nil {
						log.Println(fmt.Sprintf("Error occured %v", err))
						fmt.Printf("Error occured %v\n", err)
						continue
					}
					log.Println(fmt.Sprintf("OCO order was created in database"))
					fmt.Printf("OCO order was created in database\n")
				}
			}
		} else {
			log.Println(fmt.Sprintf("Something gone wrong, close all sell orders"))
			fmt.Printf("Something gone wrong, close all sell orders\n")
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
