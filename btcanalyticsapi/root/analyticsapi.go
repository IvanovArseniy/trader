package analyticsapi

import "time"

//Configuration is a configuration element from config.json
type Configuration struct {
	APIKey                   string
	Secret                   string
	PostgresConnectionString string
}

//TradeConfiguration is a configuration element from tradeConfig.json
type TradeConfiguration struct {
	Quantity float64
}

//OperationResult is a rusult of any operation
type OperationResult struct {
	Result bool
}

// Ticket is an element with ask and bid in some moment
type Ticket struct {
	ID        int64
	Bid       float64
	Ask       float64
	CreatedOn time.Time
}

//Level is support or resistance interval with start and end points
type Level struct {
	ID      int64
	BidFrom float64
	BidTo   float64
}

//Levels is an array of levels
type Levels []Level

//Candle is a candlestick
type Candle struct {
	StartBid float64
	MinBid   float64
	MaxBid   float64
	EndBid   float64
}

//Candles is an array of candles
type Candles []Candle

//Order is a real order
type Order struct {
	ID             int64
	Status         OrderStatus
	ParentOrderID  int64
	Price          float64
	StopPrice      float64
	StopPriceLimit float64
	Quantity       float64
	Side           OrderSide
	ExternalID     int64
	BuyPrice       float64
}

//OrderStatus is Opened and Closed
type OrderStatus int

//Trade is a trade
type Trade struct {
	Symbol      string
	Price       float64
	Qty         float64
	ExecutedQty float64
	QuoteQty    float64
	IsBuyer     bool
}

const (
	//OpenedOrder orderstatus value
	OpenedOrder OrderStatus = 1
	//ClosedOrder orderstatus 2
	ClosedOrder OrderStatus = 2
	//CanceledOrder orderstatus 2
	CanceledOrder OrderStatus = 3
)

//OrderSide is a side of an order
type OrderSide int

const (
	//SellSide is a SELL order side
	SellSide OrderSide = 1
	//BuySide is a BUY order side
	BuySide OrderSide = 2
)
