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

	"github.com/tkanos/gonfig"
)

func main() {
	fmt.Printf("Orderer started\n")

	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	tradeConfig := orderer.TradeConfiguration{}
	err = gonfig.GetConf("config/tradeConfig.json", &tradeConfig)
	if err != nil {
		return
	}

	for {
		log.Println("------------------------------------------------")
		fmt.Printf("------------------------------------------------\n")
		time.Sleep(time.Duration(tradeConfig.RunInterval) * time.Second)
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

		level, err := postgresservice.GetLevelAtTop(ticket.Ask)
		if err != nil {
			log.Println(fmt.Sprintf("Error occured %v", err))
			fmt.Printf("Error occured %v\n", err)
			continue
		}
		if level != (orderer.Level{}) {
			log.Println(fmt.Sprintf("Level %f was found", level.BidTo))
			fmt.Printf("Level %f was found\n", level.BidTo)
		} else {
			log.Println(fmt.Sprintf("Cant find level for this ticket"))
			fmt.Printf("Cant find level for this ticket\n")
		}

		if len(openedOrders) == 0 {
			if level != (orderer.Level{}) {
				candles, err := postgresservice.GetTwoLastCandles()
				if err != nil {
					log.Println(fmt.Sprintf("Error occured %v", err))
					fmt.Printf("Error occured %v\n", err)
					continue
				}
				log.Println(fmt.Sprintf("Candle startbid=%f, endbid=%f, minbid=%f, maxbid=%f", candles[0].StartBid, candles[0].EndBid, candles[0].MinBid, candles[0].MaxBid))
				fmt.Printf("Candle startbid=%f, endbid=%f, minbid=%f, maxbid=%f\n", candles[0].StartBid, candles[0].EndBid, candles[0].MinBid, candles[0].MaxBid)
				log.Println(fmt.Sprintf("Ticket is %f", ticket.Ask))
				fmt.Printf("Ticket is %f\n", ticket.Ask)
				if ((candles[0].StartBid + 20) < ticket.Ask) || (candles[1].StartBid < ticket.Ask) {
					risks, err := orderservice.GetPriceByRisks(level.BidTo)
					if err != nil {
						log.Println(fmt.Sprintf("Error occured %v", err))
						fmt.Printf("Error occured %v\n", err)
					}

					makeSellOrder, err := orderservice.CheckRisksByLevel(risks, level.BidTo)
					if err != nil {
						log.Println(fmt.Sprintf("Error occured %v", err))
						fmt.Printf("Error occured %v\n", err)
					}
					log.Println(fmt.Sprintf("CheckRisksByLevel result is %v", makeSellOrder))
					fmt.Printf("CheckRisksByLevel result is %v\n", makeSellOrder)

					if makeSellOrder {
						err = orderservice.CreateOrder(level.BidTo)
						if err != nil {
							log.Println(fmt.Sprintf("Error occured %v", err))
							fmt.Printf("Error occured %v\n", err)
							continue
						}
					}
				} else {
					log.Println(fmt.Sprintf("No order because ticket comes from top"))
					fmt.Printf("No order because ticket comes from top\n")
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
			} else if order.Status != openedOrders[0].Status && order.Status == orderer.ClosedOrder {
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
					fmt.Printf("Close price %f\n", closePrice)

					risks, err := orderservice.GetPriceByRisks(closePrice)
					if err != nil {
						log.Println(fmt.Sprintf("Error occured %v", err))
						fmt.Printf("Error occured %v\n", err)
					}

					buyPrice, stopLossPrice, err := orderservice.GetConfirmedRisk(risks)
					if err != nil {
						log.Println(fmt.Sprintf("Error occured %v", err))
						fmt.Printf("Error occured %v\n", err)
					}

					err = orderservice.CreateOrderWithSlopLoss(buyPrice, stopLossPrice, openedOrders[0].ID)
					if err != nil {
						log.Println(fmt.Sprintf("Error occured %v", err))
						fmt.Printf("Error occured %v\n", err)
						continue
					}
				}
			} else {
				if level == (orderer.Level{}) && openedOrders[0].ParentOrderID == 0 {
					log.Println(fmt.Sprintf("Level wasnt found, close opened sell orders"))
					fmt.Printf("Level wasnt found, close opened sell orders\n")
					res, err := binanceservice.CloseOrder(openedOrders[0].ExternalID)
					if err != nil {
						log.Println(fmt.Sprintf("Error occured %v", err))
						fmt.Printf("Error occured %v\n", err)
						continue
					}
					if res {
						_, err := orderservice.CloseOrder(openedOrders[0].ID)
						if err != nil {
							log.Println(fmt.Sprintf("Error occured %v", err))
							fmt.Printf("Error occured %v\n", err)
							continue
						}
					} else {
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
