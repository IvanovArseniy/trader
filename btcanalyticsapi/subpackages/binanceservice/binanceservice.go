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
	analitycsapi "trader/btcanalyticsapi/root"

	"github.com/tkanos/gonfig"
)

type serverTime struct {
	Time int64 `json:"serverTime,int"`
}

type binanceOrder struct {
	symbol              string
	ID                  int64 `json:"orderId,int"`
	orderListID         int
	clientOrderID       string
	Price               float64 `json:",string"`
	OrigQty             float64 `json:",string"`
	ExecutedQty         float64 `json:",string"`
	CummulativeQuoteQty float64 `json:",string"`
	Status              string  `json:"status"`
	TimeInForce         string  `json:"timeInForce"`
	OrderType           string  `json:"type"`
	Side                string  `json:"side"`
	StopPrice           float64 `json:",string"`
	IcebergQty          float64 `json:",string"`
	time                int64
	updateTime          int64
	isWorking           bool
	OrigQuoteOrderQty   float64 `json:",string"`
}

type binanceOrders []binanceOrder

type binanceOrderList struct {
	ListOrderStatus string         `json:"listOrderStatus"`
	ListStatusType  string         `json:"listStatusType"`
	OrderListID     int64          `json:"orderListId"`
	Orders          []binanceOrder `json:"orders"`
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

func getSideByID(orderside analitycsapi.OrderSide) (side string) {
	if orderside == analitycsapi.BuySide {
		side = "BUY"
	} else if orderside == analitycsapi.SellSide {
		side = "SELL"
	}
	return
}

func getIDBySide(side string) (orderside analitycsapi.OrderSide) {
	if side == "BUY" {
		orderside = analitycsapi.BuySide
	}
	if side == "SELL" {
		orderside = analitycsapi.SellSide
	}
	return
}

func getIDByStatus(status string) (orderstatus analitycsapi.OrderStatus) {
	if status == "NEW" || status == "PARTIALLY_FILLED" || status == "PENDING_CANCEL" {
		orderstatus = analitycsapi.OpenedOrder
	}
	if status == "FILLED" {
		orderstatus = analitycsapi.ClosedOrder
	}
	if status == "CANCELED" || status == "REJECTED" || status == "EXPIRED" {
		orderstatus = analitycsapi.CanceledOrder
	}
	return
}

//GetOpenedOrders function get all opened orders
func GetOpenedOrders() (orders []analitycsapi.Order, err error) {
	configuration := analitycsapi.Configuration{}
	err = gonfig.GetConf("config/config.json", &configuration)
	if err != nil {
		return
	}

	url := "https://api.binance.com/api/v3/openOrders"
	timestamp, err := GetServerTime()
	if err != nil {
		return
	}
	queryString := fmt.Sprintf("symbol=BTCUSDT&timestamp=%v", timestamp)
	signature, err := createSignature(queryString, configuration.Secret)
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.Header.Set("X-MBX-APIKEY", configuration.APIKey)
	req.URL.RawQuery = fmt.Sprintf("symbol=BTCUSDT&timestamp=%v&signature=%v", timestamp, signature)
	binanceClient := http.Client{Timeout: 30 * time.Second}
	resp, err := binanceClient.Do(req)
	if err != nil {
		return
	}
	if resp != nil {
		defer resp.Body.Close()
	} else {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	binanceOrders := binanceOrders{}
	err = json.Unmarshal(body, &binanceOrders)
	if err != nil {
		return
	}
	orders = []analitycsapi.Order{}
	for _, o := range binanceOrders {
		orders = append(orders, analitycsapi.Order{Price: o.Price, Quantity: o.OrigQty, Status: getIDByStatus(o.Status), Side: getIDBySide(o.Side), ExternalID: o.ID})
	}
	return
}

//CreateOrder function create order at binance
func CreateOrder(order analitycsapi.Order) (orderID int64, err error) {
	configuration := analitycsapi.Configuration{}
	err = gonfig.GetConf("config/config.json", &configuration)
	if err != nil {
		return
	}

	url := "https://api.binance.com/api/v3/order"
	timestamp, err := GetServerTime()
	if err != nil {
		return
	}
	side := getSideByID(order.Side)
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
func CreateOcoOrder(order analitycsapi.Order) (orderID int64, err error) {
	configuration := analitycsapi.Configuration{}
	err = gonfig.GetConf("config/config.json", &configuration)
	if err != nil {
		return
	}

	url := "https://api.binance.com/api/v3/order/oco"
	timestamp, err := GetServerTime()
	if err != nil {
		return
	}
	side := getSideByID(order.Side)
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
	binanceOrderList := binanceOrderList{}
	err = json.Unmarshal(body, &binanceOrderList)
	if err != nil {
		return
	}
	if binanceOrderList.Orders != nil && len(binanceOrderList.Orders) > 0 {
		orderID = binanceOrderList.Orders[0].ID
	} else {
		err = errors.New("Oco order was not created")
	}
	return
}

//CloseOrder function close order by orderID
func CloseOrder(orderID int64) (result bool, err error) {
	configuration := analitycsapi.Configuration{}
	err = gonfig.GetConf("config/config.json", &configuration)
	if err != nil {
		return
	}

	url := "https://api.binance.com/api/v3/order"
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
	result = getIDByStatus(binanceOrder.Status) == analitycsapi.ClosedOrder || getIDByStatus(binanceOrder.Status) == analitycsapi.CanceledOrder
	return
}
