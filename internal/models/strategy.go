package models

import (
	"time"
)

type StrategyStatus string

const (
	StrategyInitialized StrategyStatus = "INITIALIZED"
	StrategyArmed       StrategyStatus = "ARMED"
	StrategyTriggered   StrategyStatus = "TRIGGERED"
	StrategyExecuted    StrategyStatus = "EXECUTED"
	StrategyCompleted   StrategyStatus = "COMPLETED"
	StrategyFailed      StrategyStatus = "FAILED"
)

type PriceCondition string

const (
	PriceAbove PriceCondition = "ABOVE"
	PriceBelow PriceCondition = "BELOW"
)

type TradingStrategy struct {
	ID             string         `json:"id"`
	OrderRequest   OrderRequest   `json:"order_request"`
	TargetPrice    float64        `json:"target_price"`
	PriceCondition PriceCondition `json:"price_condition"`
	Status         StrategyStatus `json:"status"`
	CurrentPrice   float64        `json:"current_price"`
	ExecutedOrder  *Order         `json:"executed_order,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type StrategyRequest struct {
	OrderRequest   OrderRequest   `json:"order_request"`
	TargetPrice    float64        `json:"target_price"`
	PriceCondition PriceCondition `json:"price_condition"`
}

type PriceSignal struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}

type StrategyResponse struct {
	Strategy TradingStrategy `json:"strategy"`
}
