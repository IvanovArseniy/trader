package binanceservice

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
	tracker "trader/btctracker/root"
)

type binanceTicket struct {
	symbol             string
	PriceChange        float32 `json:",string"`
	PriceChangePercent float32 `json:",string"`
	WeightedAvgPrice   float32 `json:",string"`
	PrevClosePrice     float32 `json:",string"`
	LastPrice          float32 `json:",string"`
	LastQty            float32 `json:",string"`
	BidPrice           float32 `json:",string"`
	BidQty             float32 `json:",string"`
	AskPrice           float32 `json:",string"`
	AskQty             float32 `json:",string"`
	OpenPrice          float32 `json:",string"`
	HighPrice          float32 `json:",string"`
	LowPrice           float32 `json:",string"`
	Volume             float32 `json:",string"`
	QuoteVolume        float32 `json:",string"`
	openTime           int64
	closeTime          int64
	firstID            int32
	lastID             int32
	count              int32
}

//GetTicket requests ticket data from binance
func GetTicket() (ticket tracker.Ticket, funcErr error) {
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

	ticket = tracker.Ticket{Bid: binanceTicket.BidPrice, Ask: binanceTicket.AskPrice, CreatedOn: time.Now()}
	return
}
