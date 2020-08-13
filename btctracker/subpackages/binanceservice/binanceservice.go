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
	priceChangePercent float32 `json:",string"`
	weightedAvgPrice   float32 `json:",string"`
	prevClosePrice     float32 `json:",string"`
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
func GetTicket() (ticket tracker.Ticket) {
	url := "https://api.binance.com/api/v3/ticker/24hr?symbol=BTCUSDT"
	binanceClient := http.Client{
		Timeout: time.Second * 30,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}

	res, httpErr := binanceClient.Do(req)
	if httpErr != nil {
		panic(httpErr)
	}
	defer res.Body.Close()

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		panic(readErr)
	}

	binanceTicket := binanceTicket{}
	jsonErr := json.Unmarshal(body, &binanceTicket)
	if jsonErr != nil {
		panic(jsonErr)
	}

	ticket = tracker.Ticket{Bid: binanceTicket.BidPrice, Ask: binanceTicket.AskPrice, CreatedOn: time.Now()}
	return
}
