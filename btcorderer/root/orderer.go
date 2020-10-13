package orderer

import "time"

//Configuration is a configuration element from config.json
type Configuration struct {
	APIKey                   string
	Secret                   string
	PostgresConnectionString string
}

//TradeConfiguration is a configuration element from tradeConfig.json
type TradeConfiguration struct {
	RunInterval          int64
	RoundPriceLimiter    int64
	RoundPriceAddition   int64
	RoundPriceBase       int64
	Quantity             float64
	StopPriceAddition    float64
	PriceGrowthCoef      float64
	StopPriceGapForOrder float64
	GetLevelBottomGap    int64
	GetPriceGrowthLimit  int64
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

//Risk is a risk
type Risk struct {
	Buy      float64
	StopLoss float64
}

//Risks is an array of Risk
type Risks []Risk

//Candle is a market candlestick
type Candle struct {
	ID       int64
	StartBid float64
	EndBid   float64
	MinBid   float64
	MaxBid   float64
}
