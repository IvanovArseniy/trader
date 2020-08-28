package orderservice

import (
	"errors"
	"fmt"
	"log"
	orderer "trader/btcorderer/root"
	"trader/btcorderer/subpackages/binanceservice"
	"trader/btcorderer/subpackages/postgresservice"

	"github.com/tkanos/gonfig"
)

//CreateOrder function creates orders in binance and database
func CreateOrder(level orderer.Level) (err error) {
	tradeConfig := orderer.TradeConfiguration{}
	err = gonfig.GetConf("config/tradeConfig.json", &tradeConfig)
	if err != nil {
		return
	}

	order := orderer.Order{Price: level.BidFrom, Quantity: tradeConfig.Quantity, Side: orderer.SellSide, ParentOrderID: 0, Status: orderer.OpenedOrder, BuyPrice: level.BidTo}
	p := int64(order.Price) % tradeConfig.RoundPriceBase
	if p < tradeConfig.RoundPriceLimiter {
		order.Price = float64(int64(order.Price)/100*100 - tradeConfig.RoundPriceAddition)
	}
	log.Println(fmt.Sprintf("Level was found, create order price:%f quantity%f", order.Price, order.Quantity))
	fmt.Printf("Level was found, create order price:%f quantity%f\n", order.Price, order.Quantity)
	orderID, err := binanceservice.CreateOrder(order)
	if err != nil {
		return
	}
	log.Println(fmt.Sprintf("Order was created on binance, id=%v", orderID))
	fmt.Printf("Order was created on binance, id=%v\n", orderID)
	order.ExternalID = orderID
	err = postgresservice.CreateOrder(order)
	if err != nil {
		binanceservice.CloseOrder(int64(orderID))
		log.Println(fmt.Sprintf("Forced to close order"))
		fmt.Printf("Forced to close order\n")
		return
	}
	log.Println(fmt.Sprintf("Order was created in database"))
	fmt.Printf("Order was created in database\n")
	return
}

//GetOrder function get order by externalId
func GetOrder(externalID int64) (order orderer.Order, err error) {
	order, err = binanceservice.GetOrder(externalID)
	return
}

//CloseOrder function close order
func CloseOrder(orderID int64) (result bool, err error) {
	result, err = postgresservice.CloseOrder(orderID)
	if err != nil {
		return
	}
	if !result {
		log.Println("Error occuder Order do not closed in database")
		fmt.Printf("Error occuder Order do not closed in database\n")
	} else {
		log.Println(fmt.Sprintf("Opened order orderID=%v was closed", orderID))
		fmt.Printf("Opened order orderID=%v was closed\n", orderID)
	}
	return
}

//CreateOrderWithSlopLoss function creates buy order and stoploss order
func CreateOrderWithSlopLoss(closePrice float64, priceGrowth float64, levelMaxPrice float64, parentOrderID int64) (err error) {
	tradeConfig := orderer.TradeConfiguration{}
	err = gonfig.GetConf("config/tradeConfig.json", &tradeConfig)
	if err != nil {
		return
	}

	stopPriceLimit := (levelMaxPrice + tradeConfig.StopPriceAddition)
	s := int64(stopPriceLimit) % tradeConfig.RoundPriceBase
	if s > (tradeConfig.RoundPriceBase - tradeConfig.RoundPriceLimiter) {
		stopPriceLimit = float64(int64(stopPriceLimit)/100*100 + (tradeConfig.RoundPriceBase + tradeConfig.RoundPriceAddition))
	}

	buyPrice := (closePrice - (priceGrowth / tradeConfig.PriceGrowthCoef))
	b := int64(buyPrice) % tradeConfig.RoundPriceBase
	if b > (tradeConfig.RoundPriceBase - tradeConfig.RoundPriceLimiter) {
		buyPrice = float64(int64(buyPrice)/100*100 + (tradeConfig.RoundPriceBase + tradeConfig.RoundPriceAddition))
	}
	order := orderer.Order{Price: buyPrice, Quantity: tradeConfig.Quantity, Side: orderer.BuySide, StopPrice: (stopPriceLimit - tradeConfig.StopPriceGapForOrder), StopPriceLimit: stopPriceLimit, ParentOrderID: parentOrderID, Status: orderer.OpenedOrder}
	log.Println(fmt.Sprintf("It was an order to sell BTC. Create OCO order price:%f quantity:%f stopPrice:%f stopPriceLimit:%f", order.Price, order.Quantity, order.StopPrice, order.StopPriceLimit))
	fmt.Printf("It was an order to sell BTC. Create OCO order price:%f quantity:%f stopPrice:%f stopPriceLimit:%f\n", order.Price, order.Quantity, order.StopPrice, order.StopPriceLimit)
	orderID, err := binanceservice.CreateOcoOrder(order)
	if err != nil {
		return
	}
	log.Println(fmt.Sprintf("OCO order was created, binanceid=%v", orderID))
	fmt.Printf("OCO order was created, binanceid=%v\n", orderID)
	order.ExternalID = orderID
	err = postgresservice.CreateOrder(order)
	if err != nil {
		panic(fmt.Sprintf("Critical error occured %v. Program terminated", err))
	}
	log.Println(fmt.Sprintf("OCO order was created in database"))
	fmt.Printf("OCO order was created in database\n")
	return
}

//CloseOpenedSellOrders function closes all opened sell orders
func CloseOpenedSellOrders() {
	orderIDs, err := postgresservice.CloseOpenedSellOrders()
	log.Println(fmt.Sprintf("%v orders was closed", len(orderIDs)))
	fmt.Printf("%v orders was closed\n", len(orderIDs))
	if err != nil {
		log.Println(fmt.Sprintf("Error occured %v", err))
		fmt.Printf("Error occured %v\n", err)
	}
	for _, orderID := range orderIDs {
		res, err := binanceservice.CloseOrder(int64(orderID))
		log.Println(fmt.Sprintf("Close binance order result=%v", res))
		fmt.Printf("Close binance order result=%v\n", res)
		if err != nil {
			log.Println(fmt.Sprintf("Error occured %v", err))
			fmt.Printf("Error occured %v\n", err)
		}
	}
	return
}

//GetClosePrice function get real close price of last sell order
func GetClosePrice() (closePrice float64, err error) {
	trades, err := binanceservice.GetTrades()
	if err != nil {
		return
	}

	closePrice = 0
	for i := (len(trades) - 1); i >= 0; i-- {
		t := trades[i]
		if !t.IsBuyer && t.Symbol == "BTCUSDT" {
			closePrice = t.Price
			break
		}
	}
	if closePrice == 0 {
		err = errors.New("ClosePrice wasnt calculated")
	}
	return
}
