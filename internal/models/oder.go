package models

import (
	"time"
)

type OrderType string
type OrderStatus string
type OrderSide string

const (
	Market OrderType = "MARKET"
	Limit  OrderType = "LIMIT"
)

const (
	Created     OrderStatus = "CREATED"
	Validated   OrderStatus = "VALIDATED"
	Rejected    OrderStatus = "REJECTED"
	Executing   OrderStatus = "EXECUTING"
	PartialFill OrderStatus = "PARTIAL_FILL"
	Filled      OrderStatus = "FILLED"
	Settled     OrderStatus = "SETTLED"
	Failed      OrderStatus = "FAILED"
)

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

type Order struct {
	ID             string      `json:"id"`
	AccountID      string      `json:"account_id"`
	Symbol         string      `json:"symbol"`
	Quantity       float64     `json:"quantity"`
	Price          float64     `json:"price,omitempty"`
	Type           OrderType   `json:"type"`
	Side           OrderSide   `json:"side"`
	Status         OrderStatus `json:"status"`
	FilledQuantity float64     `json:"filled_quantity"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type OrderRequest struct {
	AccountID string    `json:"account_id"`
	Symbol    string    `json:"symbol"`
	Quantity  float64   `json:"quantity"`
	Price     float64   `json:"price,omitempty"`
	Type      OrderType `json:"type"`
	Side      OrderSide `json:"side"`
}

type OrderResponse struct {
	Order Order `json:"order"`
}
