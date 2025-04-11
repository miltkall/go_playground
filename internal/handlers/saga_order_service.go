package handlers

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/miltkall/go_playground/internal/models"
	restate "github.com/restatedev/sdk-go"
	"github.com/samber/oops"
)

// OrderSagaService demonstrates the saga pattern with compensation in a single function
type OrderSagaService struct{}

// NewOrderSagaService creates a new order saga service
func NewOrderSagaService() *OrderSagaService {
	return &OrderSagaService{}
}

// ProcessOrderWithSaga handles the entire order processing workflow with compensation actions
// If a terminal error occurs, earlier successful steps are compensated (rolled back)
func (s *OrderSagaService) ProcessOrderWithSaga(ctx restate.Context, request models.OrderRequest) (*models.OrderResponse, error) {
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

	slog.Info("Starting saga order workflow", "order_id", order.ID, "symbol", order.Symbol)

	// Track compensation functions in a slice
	compensations := make([]func() error, 0, 3)

	// Handle errors by running compensations if it's a terminal error
	handleError := func(err error) error {
		if restate.IsTerminalError(err) {
			slog.Error("Terminal error occurred, running compensations", "error", err.Error())
			// Run compensations in reverse order (LIFO)
			for i := len(compensations) - 1; i >= 0; i-- {
				if compErr := compensations[i](); compErr != nil {
					slog.Error("Compensation failed", "index", i, "error", compErr.Error())
				}
			}
		}
		return err
	}

	// Single function with entire workflow
	result, err := restate.Run(ctx, func(runCtx restate.RunContext) (models.Order, error) {
		// var err error

		// ----- STEP 1: Validate Order -----
		slog.Info("===> Step 1: Validating order", "order_id", order.ID)

		// Simulate validation work
		time.Sleep(500 * time.Millisecond)

		// Create dummy reservation ID
		reservationId := "res_" + order.ID
		slog.Info("Reserved validation capacity", "order_id", order.ID, "reservation_id", reservationId)

		order.Status = models.Validated
		order.UpdatedAt = time.Now()

		// Add compensation for validation
		compensations = append(compensations, func() error {
			slog.Info("Compensation: Releasing validation reservation", "reservation_id", reservationId)
			return nil
		})

		// ----- STEP 2: Execute Order -----
		slog.Info("===> Step 2: Executing order", "order_id", order.ID)

		order.Status = models.Executing

		// Create dummy market order ID
		marketOrderId := "mkt_" + order.ID
		slog.Info("Created market order", "order_id", order.ID, "market_order_id", marketOrderId)

		// Simulate partial fill
		order.FilledQuantity = order.Quantity * 0.5
		order.Status = models.PartialFill
		order.UpdatedAt = time.Now()
		slog.Info("Order partially filled", "order_id", order.ID, "filled_quantity", order.FilledQuantity)

		// Simulate work that might fail
		time.Sleep(1 * time.Second)

		// Test failure scenario
		if order.Symbol == "FAIL_DURING_EXECUTION" {
			return order, restate.TerminalError(fmt.Errorf("market execution failed"))
		}

		// Complete the fill
		order.FilledQuantity = order.Quantity
		order.Status = models.Filled
		order.UpdatedAt = time.Now()

		// Add compensation for execution
		compensations = append(compensations, func() error {
			slog.Info("Compensation: Cancelling market order", "market_order_id", marketOrderId)
			return nil
		})

		// ----- STEP 3: Settle Order -----
		slog.Info("===> Step 3: Settling order", "order_id", order.ID)

		// Create dummy settlement ID
		settlementId := "stl_" + order.ID
		slog.Info("Created settlement record", "order_id", order.ID, "settlement_id", settlementId)

		// Simulate settlement work
		time.Sleep(700 * time.Millisecond)

		// Test failure scenario
		if order.Symbol == "FAIL_DURING_SETTLEMENT" {
			return order, restate.TerminalError(fmt.Errorf("settlement failed"))
		}

		order.Status = models.Settled
		order.UpdatedAt = time.Now()

		// Add compensation for settlement
		compensations = append(compensations, func() error {
			slog.Info("Compensation: Reversing settlement", "settlement_id", settlementId)
			return nil
		})

		slog.Info("Order processing completed successfully", "order_id", order.ID)
		return order, nil
	})

	// Handle errors at the workflow level
	if err != nil {
		errorBuilder := oops.
			Code("order_saga_error").
			In("saga_handler").
			With("order_id", order.ID).
			With("status", order.Status)

		return nil, handleError(errorBuilder.Wrapf(err, "Order saga failed"))
	}

	return &models.OrderResponse{Order: result}, nil
}
