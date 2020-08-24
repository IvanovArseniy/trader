package main

import (
	"fmt"
	"log"
	"os"
	"time"
	orderer "trader/btcorderer/root"
	"trader/btcorderer/subpackages/binanceservice"
	"trader/btcorderer/subpackages/orderservice"
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
			orderservice.CloseOpenedSellOrders()
		}
		if len(openedOrders) == 0 {
			if level != (orderer.Level{}) {
				err = orderservice.CreateOrder(level)
				if err != nil {
					log.Println(fmt.Sprintf("Error occured %v", err))
					fmt.Printf("Error occured %v\n", err)
					continue
				}
			}
		} else if len(openedOrders) == 1 {
			log.Println(fmt.Sprintf("Get info from binance for order binanceid=%v", openedOrders[0].ExternalID))
			fmt.Printf("Get info from binance for order binanceid=%v\n", openedOrders[0].ExternalID)
			order, err := orderservice.GetOrder(openedOrders[0].ExternalID)
			if err != nil {
				log.Println(fmt.Sprintf("Error occured %v", err))
				fmt.Printf("Error occured %v\n", err)
				continue
			}
			log.Println(fmt.Sprintf("Opened order was found, binanceID=%v status=%v", order.ExternalID, order.Status))
			fmt.Printf("Opened order was found, binanceID=%v status=%v\n", order.ExternalID, order.Status)
			if order.Status != openedOrders[0].Status && order.Status == orderer.CanceledOrder {
				_, err := orderservice.CloseOrder(openedOrders[0].ID)
				if err != nil {
					log.Println(fmt.Sprintf("Error occured %v", err))
					fmt.Printf("Error occured %v\n", err)
					continue
				}
			}
			if order.Status != openedOrders[0].Status && order.Status == orderer.ClosedOrder {
				_, err := orderservice.CloseOrder(openedOrders[0].ID)
				if err != nil {
					log.Println(fmt.Sprintf("Error occured %v", err))
					fmt.Printf("Error occured %v\n", err)
				}
				if openedOrders[0].ParentOrderID == 0 {
					closePrice, err := orderservice.GetClosePrice()
					if err != nil {
						log.Println(fmt.Sprintf("Error occured %v", err))
						fmt.Printf("Error occured %v\n", err)
						closePrice = openedOrders[0].Price
					}
					log.Println(fmt.Sprintf("Close price i %f", closePrice))
					fmt.Printf("Close price i %f", closePrice)

					priceGrowth, err := postgresservice.GetPriceGrowth()
					if err != nil {
						log.Println(fmt.Sprintf("Error occured %v", err))
						fmt.Printf("Error occured %v\n", err)
						priceGrowth = 200
					}
					log.Println(fmt.Sprintf("Price growth id %f", priceGrowth))
					fmt.Printf("Price growth id %f", priceGrowth)

					err = orderservice.CreateOrderWithSlopLoss(closePrice, priceGrowth, openedOrders[0].BuyPrice, openedOrders[0].ID)
					if err != nil {
						log.Println(fmt.Sprintf("Error occured %v", err))
						fmt.Printf("Error occured %v\n", err)
						continue
					}
				}
			}
		} else {
			log.Println(fmt.Sprintf("Something gone wrong, close all sell orders"))
			fmt.Printf("Something gone wrong, close all sell orders\n")
			orderservice.CloseOpenedSellOrders()
		}
	}

}
