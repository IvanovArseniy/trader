package binanceservice

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	orderer "trader/btcorderer/root"

	"github.com/tkanos/gonfig"
)

type binanceTicket struct {
	symbol             string
	PriceChange        float64 `json:",string"`
	PriceChangePercent float64 `json:",string"`
	WeightedAvgPrice   float64 `json:",string"`
	PrevClosePrice     float64 `json:",string"`
	LastPrice          float64 `json:",string"`
	LastQty            float64 `json:",string"`
	BidPrice           float64 `json:",string"`
	BidQty             float64 `json:",string"`
	AskPrice           float64 `json:",string"`
	AskQty             float64 `json:",string"`
	OpenPrice          float64 `json:",string"`
	HighPrice          float64 `json:",string"`
	LowPrice           float64 `json:",string"`
	Volume             float64 `json:",string"`
	QuoteVolume        float64 `json:",string"`
	openTime           int64
	closeTime          int64
	firstID            int32
	lastID             int32
	count              int32
}

type serverTime struct {
	Time int64 `json:"serverTime,int"`
}

type binanceOrder struct {
	Symbol              string  `json:",string"`
	ID                  int64   `json:"orderId,int"`
	OrderListID         int     `json:"orderListId,int"`
	ClientOrderID       string  `json:"clientOrderId,string"`
	TransactTime        int64   `json:",int"`
	Price               float64 `json:",string"`
	OrigQty             float64 `json:",string"`
	ExecutedQty         float64 `json:",string"`
	CummulativeQuoteQty float64 `json:",string"`
	Status              string  `json:",string"`
	TimeInForce         string  `json:",string"`
	OrderType           string  `json:"type,string"`
	Side                string  `json:",string"`
}

type binanceOrderStatus string

const (
	closeStatus binanceOrderStatus = "CANCELED"
)

//GetTicket requests ticket data from binance
func GetTicket() (ticket orderer.Ticket, funcErr error) {
	url := "https://api.binance.com/api/v3/ticker/24hr?symbol=BTCUSDT"
	binanceClient := http.Client{
		Timeout: time.Second * 30,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		funcErr = err
	}

	res, httpErr := binanceClient.Do(req)
	if httpErr != nil {
		funcErr = httpErr
	}
	if res != nil {
		defer res.Body.Close()
	} else {
		return
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		funcErr = readErr
	}

	binanceTicket := binanceTicket{}
	jsonErr := json.Unmarshal(body, &binanceTicket)
	if jsonErr != nil {
		funcErr = jsonErr
	}

	ticket = orderer.Ticket{Bid: binanceTicket.BidPrice, Ask: binanceTicket.AskPrice, CreatedOn: time.Now()}
	return
}

//GetServerTime requests server time from binance
func GetServerTime() (sTime int64, funcErr error) {
	url := "https://api.binance.com/api/v3/time"
	binanceClient := http.Client{
		Timeout: time.Second * 30,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		funcErr = err
	}

	res, httpErr := binanceClient.Do(req)
	if httpErr != nil {
		funcErr = httpErr
	}
	if res != nil {
		defer res.Body.Close()
	} else {
		return
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		funcErr = readErr
	}

	serverTime := serverTime{}
	jsonErr := json.Unmarshal(body, &serverTime)
	if jsonErr != nil {
		funcErr = jsonErr
	}

	sTime = serverTime.Time
	return
}

func createSignature(data string, secret string) (signature string, err error) {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	signature = hex.EncodeToString(h.Sum(nil))
	return
}

//GetOrder function get order from binance by ID
func GetOrder(orderOD int64) (order orderer.Order, err error) {
	// https://github.com/binance-exchange/binance-official-api-docs/blob/master/rest-api.md#order-book
	// https://academy.binance.com/tutorials/what-is-an-oco-order
	return
}

func getSide(orderside orderer.OrderSide) (side string) {
	if orderside == orderer.BuySide {
		side = "BUY"
	} else if orderside == orderer.SellSide {
		side = "SELL"
	}
	return
}

//CreateOrder function create order at binance
func CreateOrder(order orderer.Order) (orderID int64, err error) {
	configuration := orderer.Configuration{}
	err = gonfig.GetConf("config/config.json", &configuration)
	if err != nil {
		return
	}

	url := "https://api.binance.com/api/v3/order"
	timestamp, err := GetServerTime()
	if err != nil {
		return
	}
	side := getSide(order.Side)
	queryString := fmt.Sprintf("symbol=BTCUSDT&side=%v&type=LIMIT&timeInForce=GTC&quantity=%f&price=%.2f&timestamp=%v", side, order.Quantity, order.Price, timestamp)
	signature, err := createSignature(queryString, configuration.Secret)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return
	}
	req.Header.Set("X-MBX-APIKEY", configuration.APIKey)
	req.URL.RawQuery = fmt.Sprintf("symbol=BTCUSDT&side=%v&type=LIMIT&timeInForce=GTC&quantity=%f&price=%.2f&timestamp=%v&signature=%v", side, order.Quantity, order.Price, timestamp, signature)

	binanceClient := http.Client{Timeout: 30 * time.Second}
	response, err := binanceClient.Do(req)
	if err != nil {
		return
	}
	if response != nil {
		defer response.Body.Close()
	} else {
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	binanceOrder := binanceOrder{}
	err = json.Unmarshal(body, &binanceOrder)
	if err != nil {
		return
	}
	orderID = binanceOrder.ID
	return
}

//CreateOcoOrder function creates a pair of orders -TAKE_PROFIT and STOP_LOSS orders
func CreateOcoOrder(order orderer.Order) (orderID int64, err error) {
	configuration := orderer.Configuration{}
	err = gonfig.GetConf("config/config.json", &configuration)
	if err != nil {
		return
	}

	url := "https://api.binance.com/api/v3/order/oco"
	timestamp, err := GetServerTime()
	if err != nil {
		return
	}
	side := getSide(order.Side)
	queryString := fmt.Sprintf("symbol=BTCUSDT&side=%v&quantity=%f&price=%.2f&stopPrice=%.2f&stopLimitPrice=%.2f&stopLimitTimeInForce=GTC&timestamp=%v", side, order.Quantity, order.Price, order.StopPrice, order.StopPriceLimit, timestamp)
	signature, err := createSignature(queryString, configuration.Secret)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return
	}
	req.Header.Set("X-MBX-APIKEY", configuration.APIKey)
	req.URL.RawQuery = fmt.Sprintf("symbol=BTCUSDT&side=%v&quantity=%f&price=%.2f&stopPrice=%.2f&stopLimitPrice=%.2f&stopLimitTimeInForce=GTC&timestamp=%v&signature=%v", side, order.Quantity, order.Price, order.StopPrice, order.StopPriceLimit, timestamp, signature)

	binanceClient := http.Client{Timeout: 30 * time.Second}
	response, err := binanceClient.Do(req)
	if err != nil {
		return
	}
	if response != nil {
		defer response.Body.Close()
	} else {
		err = errors.New("Empty body")
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	binanceOrder := binanceOrder{}
	err = json.Unmarshal(body, &binanceOrder)
	if err != nil {
		return
	}
	orderID = binanceOrder.ID
	return
}

//CloseOrder function close order by orderID
func CloseOrder(orderID int64) (result bool, err error) {
	configuration := orderer.Configuration{}
	err = gonfig.GetConf("config/gonfig.json", &configuration)
	if err != nil {
		return
	}

	url := "https://api.binance.com//api/v3/order"
	timestamp, err := GetServerTime()
	if err != nil {
		return
	}
	queryString := fmt.Sprintf("symbol=BTCUSDT&orderId=%v&timestamp=%v", orderID, timestamp)
	signature, err := createSignature(queryString, configuration.Secret)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return
	}
	req.Header.Set("X-MBX-APIKEY", configuration.APIKey)
	req.URL.RawQuery = fmt.Sprintf("symbol=BTCUSDT&orderId=%v&timestamp=%v&signature=%v", orderID, timestamp, signature)
	binanceClient := http.Client{Timeout: 30 * time.Second}
	response, err := binanceClient.Do(req)
	if err != nil {
		return
	}
	if response != nil {
		defer response.Body.Close()
	} else {
		err = errors.New("Empty body")
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	binanceOrder := binanceOrder{}
	err = json.Unmarshal(body, &binanceOrder)
	if err != nil {
		return
	}
	result = binanceOrder.Status == string(closeStatus)
	return
}
