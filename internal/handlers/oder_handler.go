package handlers

import (
	"fmt"
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
	// Create error builder for consistent error formatting
	errorBuilder := oops.
		Code("order_processing_error").
		In("order_handler").
		With("account_id", request.AccountID).
		With("symbol", request.Symbol)

	// Generate a unique order ID - this is deterministic and will be the same on retries
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

	slog.Info("Starting order processing workflow",
		"order_id", order.ID,
		"account_id", order.AccountID,
		"symbol", order.Symbol)

	// --------- STEP 1: Validate Order ---------
	slog.Info("====> Starting Step 1: Validating order", "order_id", order.ID)
	validatedOrder, err := restate.Run(ctx, func(runCtx restate.RunContext) (models.Order, error) {
		slog.Info("Executing validation for order", "order_id", order.ID)

		// Simulate validation work
		time.Sleep(1 * time.Second)

		order.Status = models.Validated
		order.UpdatedAt = time.Now()
		slog.Info("Order validated successfully", "order_id", order.ID)

		return order, nil
	})

	if err != nil {
		return nil, errorBuilder.
			With("step", "validation").
			With("status", order.Status).
			Wrapf(err, "Order validation failed")
	}

	// Update order reference for next step
	order = validatedOrder

	// --------- STEP 2: Execute Order ---------
	slog.Info("====> Starting Step 2: Executing order", "order_id", order.ID)
	executedOrder, err := restate.Run(ctx, func(runCtx restate.RunContext) (models.Order, error) {
		slog.Info("Executing placement for order", "order_id", order.ID)

		// Update order status
		order.Status = models.Executing
		order.UpdatedAt = time.Now()

		// Simulate partial fill with delay
		time.Sleep(2 * time.Second)

		order.FilledQuantity = order.Quantity * 0.5
		order.Status = models.PartialFill
		order.UpdatedAt = time.Now()
		slog.Info("Order partially filled",
			"order_id", order.ID,
			"filled_quantity", order.FilledQuantity)

		// Simulate completing the fill with another delay
		time.Sleep(2 * time.Second)

		order.FilledQuantity = order.Quantity
		order.Status = models.Filled
		order.UpdatedAt = time.Now()
		slog.Info("Order fully filled",
			"order_id", order.ID,
			"filled_quantity", order.FilledQuantity)

		return order, nil
	})

	if err != nil {
		return nil, errorBuilder.
			With("step", "execution").
			With("status", order.Status).
			Wrapf(err, "Order execution failed")
	}

	// Update order reference for next step
	order = executedOrder

	// --------- STEP 3: Settle Order ---------
	slog.Info("====> Starting Step 3: Settling order", "order_id", order.ID)
	settledOrder, err := restate.Run(ctx, func(runCtx restate.RunContext) (models.Order, error) {
		slog.Info("Executing settlement for order", "order_id", order.ID)

		// Simulate settlement work
		time.Sleep(1 * time.Second)

		// Simulate settlement failure for specific symbols
		if order.Symbol == "ERROR" {
			order.Status = models.Failed
			slog.Error(" Order settlement failed intentionally",
				"order_id", order.ID,
				"symbol", order.Symbol)
			return order, fmt.Errorf("settlement failed for symbol %s", order.Symbol)
		}

		order.Status = models.Settled
		order.UpdatedAt = time.Now()
		slog.Info("Order settlement completed",
			"order_id", order.ID,
			"account_id", order.AccountID)

		return order, nil
	})

	if err != nil {
		return nil, errorBuilder.
			With("step", "settlement").
			With("status", order.Status).
			Wrapf(err, "Order settlement failed")
	}

	slog.Info("====> Order workflow completed successfully",
		"order_id", settledOrder.ID,
		"account_id", settledOrder.AccountID,
		"status", settledOrder.Status)

	return &models.OrderResponse{Order: settledOrder}, nil
}
