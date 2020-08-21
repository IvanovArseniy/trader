package orderer

import "time"

//Configuration is a configuration element from config.json
type Configuration struct {
	APIKey                   string
	Secret                   string
	PostgresConnectionString string
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
}

//OrderStatus is Opened and Closed
type OrderStatus int

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
