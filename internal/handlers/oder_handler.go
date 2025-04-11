package handlers

import (
	"log/slog"
	"time"

	"github.com/miltkall/go_playground/internal/models"
	restate "github.com/restatedev/sdk-go"
	"github.com/samber/oops"
)

// OrderService handles processing of trading orders
type OrderService struct{}

// NewOrderService creates a new order service
func NewOrderService() *OrderService {
	return &OrderService{}
}

// ProcessOrder is the main workflow that processes an order from validation to settlement
func (s *OrderService) ProcessOrder(ctx restate.Context, request models.OrderRequest) (*models.OrderResponse, error) {
	// Create error builder
	errorBuilder := oops.
		Code("order_processing_error").
		In("order_handler").
		With("account_id", request.AccountID).
		With("symbol", request.Symbol)

	// Generate a unique order ID
	orderId := restate.Rand(ctx).UUID().String()

	// Create order with initial status
	order := models.Order{
		ID:        orderId,
		AccountID: request.AccountID,
		Symbol:    request.Symbol,
		Quantity:  request.Quantity,
		Price:     request.Price,
		Type:      request.Type,
		Side:      request.Side,
		Status:    models.Created,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	slog.Info("Order created",
		"order_id", order.ID,
		"account_id", order.AccountID,
		"symbol", order.Symbol)

	// Step 1: Validate Order
	var err error
	order, err = s.validateOrder(ctx, order)
	if err != nil {
		return nil, errorBuilder.
			With("step", "validation").
			With("status", order.Status).
			Wrapf(err, "Order validation failed")
	}

	if order.Status == models.Rejected {
		slog.Info("Order rejected during validation",
			"order_id", order.ID,
			"account_id", order.AccountID)
		return &models.OrderResponse{Order: order}, nil
	}

	// Step 2: Execute Order
	order, err = s.executeOrder(ctx, order)
	if err != nil {
		return nil, errorBuilder.
			With("step", "execution").
			With("status", order.Status).
			Wrapf(err, "Order execution failed")
	}

	// Step 3: Settle Order
	order, err = s.settleOrder(ctx, order)
	if err != nil {
		return nil, errorBuilder.
			With("step", "settlement").
			With("status", order.Status).
			Wrapf(err, "Order settlement failed")
	}

	slog.Info("Order processed successfully",
		"order_id", order.ID,
		"account_id", order.AccountID,
		"status", order.Status)

	return &models.OrderResponse{Order: order}, nil
}

// validateOrder validates the order and updates its status
func (s *OrderService) validateOrder(ctx restate.Context, order models.Order) (models.Order, error) {
	// Use Restate.Run to ensure this operation is executed exactly once
	var validatedOrder models.Order
	_, err := restate.Run(ctx, func(runCtx restate.RunContext) (restate.Void, error) {
		slog.Info("Validating order", "order_id", order.ID)

		// Simulating validation logic
		// In a real implementation, this would check account balance, position limits, etc.

		// For demonstration, simulate some validation logic
		if order.Symbol == "INVALID" {
			order.Status = models.Rejected
			slog.Info("Order rejected: Invalid symbol", "order_id", order.ID)
		} else {
			order.Status = models.Validated
			slog.Info("Order validated successfully", "order_id", order.ID)
		}

		order.UpdatedAt = time.Now()
		validatedOrder = order

		return restate.Void{}, nil
	})

	if err != nil {
		return order, err
	}

	return validatedOrder, nil
}

// executeOrder executes the order on the market
func (s *OrderService) executeOrder(ctx restate.Context, order models.Order) (models.Order, error) {
	// Use Restate.Run to ensure this operation is executed exactly once
	var executedOrder models.Order
	_, err := restate.Run(ctx, func(runCtx restate.RunContext) (restate.Void, error) {
		slog.Info("Executing order", "order_id", order.ID)

		// Update order status
		order.Status = models.Executing
		order.UpdatedAt = time.Now()

		// Simulate a partial fill
		order.FilledQuantity = order.Quantity * 0.5
		order.Status = models.PartialFill
		order.UpdatedAt = time.Now()

		// Log the partial fill
		slog.Info("Order partially filled",
			"order_id", order.ID,
			"filled_quantity", order.FilledQuantity)

		// Simulate a delay between partial and complete fill
		err := restate.Sleep(ctx, 2*time.Second)
		if err != nil {
			return restate.Void{}, err
		}

		// Simulate a full fill
		order.FilledQuantity = order.Quantity
		order.Status = models.Filled
		order.UpdatedAt = time.Now()

		// Log the full fill
		slog.Info("Order fully filled",
			"order_id", order.ID,
			"filled_quantity", order.FilledQuantity)

		executedOrder = order
		return restate.Void{}, nil
	})

	if err != nil {
		return order, err
	}

	return executedOrder, nil
}

// settleOrder settles the executed order
func (s *OrderService) settleOrder(ctx restate.Context, order models.Order) (models.Order, error) {
	// Use Restate.Run to ensure this operation is executed exactly once
	var settledOrder models.Order
	_, err := restate.Run(ctx, func(runCtx restate.RunContext) (restate.Void, error) {
		slog.Info("Settling order", "order_id", order.ID)

		// Update order status
		order.Status = models.Settled
		order.UpdatedAt = time.Now()

		// Simulate settlement logic
		// In a real implementation, this would update account positions, cash balances, etc.

		slog.Info("Order settlement completed",
			"order_id", order.ID,
			"account_id", order.AccountID)

		settledOrder = order
		return restate.Void{}, nil
	})

	if err != nil {
		return order, err
	}

	return settledOrder, nil
}

