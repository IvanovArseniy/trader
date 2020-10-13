package orderservice

import (
	"errors"
	"fmt"
	"log"
	"math"
	orderer "trader/btcorderer/root"
	"trader/btcorderer/subpackages/binanceservice"
	"trader/btcorderer/subpackages/postgresservice"

	"github.com/tkanos/gonfig"
)

//CreateOrder function creates orders in binance and database
func CreateOrder(price float64) (err error) {
	tradeConfig := orderer.TradeConfiguration{}
	err = gonfig.GetConf("config/tradeConfig.json", &tradeConfig)
	if err != nil {
		return
	}

	order := orderer.Order{Price: price, Quantity: tradeConfig.Quantity, Side: orderer.SellSide, ParentOrderID: 0, Status: orderer.OpenedOrder}
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
func CreateOrderWithSlopLoss(buyPrice float64, stopLossPrice float64, parentOrderID int64) (err error) {
	tradeConfig := orderer.TradeConfiguration{}
	err = gonfig.GetConf("config/tradeConfig.json", &tradeConfig)
	if err != nil {
		return
	}

	order := orderer.Order{Price: buyPrice, Quantity: tradeConfig.Quantity, Side: orderer.BuySide, StopPrice: stopLossPrice, StopPriceLimit: stopLossPrice, ParentOrderID: parentOrderID, Status: orderer.OpenedOrder}
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

//GetPriceByRisks function calculate prices by trade risks
func GetPriceByRisks(price float64) (risks orderer.Risks, err error) {
	commission := float64(0.00075)
	priceRisk := price * commission
	risk := float64(2.5)
	for i := float64(0); i < 60; i = i + float64(2.5) {
		stopLoss := (price + i)
		stopLossRisk := (price+i)*(1+commission) - price
		buy := price - (stopLossRisk+priceRisk)*risk
		risks = append(risks, orderer.Risk{Buy: buy, StopLoss: stopLoss})
	}
	return
}

//GetConfirmedRisk functions calculates best risk by nearest bottom level
func GetConfirmedRisk(risks orderer.Risks) (buyPrice float64, stopLossPrice float64, err error) {
	levels, err := postgresservice.GetLevels()
	if err != nil {
		log.Println(fmt.Sprintf("Error occured %v", err))
		fmt.Printf("Error occured %v\n", err)
		return
	}
	buyPrice = 0
	stopLossPrice = 0
	for _, l := range levels {
		for i, r := range risks {
			if (i+1) < len(risks) && (r.Buy < l.BidTo && risks[i+1].Buy > l.BidTo) {
				buyPrice = risks[i+1].Buy
				stopLossPrice = risks[i+1].StopLoss
			} else if (i+1) < len(risks) && (r.Buy > l.BidTo && risks[i+1].Buy < l.BidTo) {
				buyPrice = r.Buy
				stopLossPrice = r.StopLoss
			}
		}
	}

	if buyPrice == 0 || stopLossPrice == 0 {
		buyPrice = risks[0].Buy
		stopLossPrice = risks[0].StopLoss
		for _, r := range risks {
			if buyPrice > r.Buy {
				buyPrice = r.Buy
				stopLossPrice = r.StopLoss
			}
		}
	}
	buyPrice = math.Floor(buyPrice*100) / 100
	stopLossPrice = math.Floor(stopLossPrice*100) / 100
	return
}

//CheckRisksByLevel functions calculates best risk by nearest bottom level
func CheckRisksByLevel(risks orderer.Risks, closePrice float64) (makeSellOrder bool, err error) {
	levels, err := postgresservice.GetLevels()
	if err != nil {
		log.Println(fmt.Sprintf("Error occured %v", err))
		fmt.Printf("Error occured %v\n", err)
		return
	}
	makeSellOrder = true
	for _, l := range levels {
		if l.BidTo > risks[0].Buy && l.BidTo < closePrice {
			makeSellOrder = false
			return
		}
	}
	return
}
