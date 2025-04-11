package handlers

import (
	"log/slog"
	"time"

	"github.com/miltkall/go_playground/internal/models"
	restate "github.com/restatedev/sdk-go"
	"github.com/samber/oops"
)

// TradingStrategyService implements a price-based trading strategy as a virtual object
type TradingStrategyService struct{}

// NewTradingStrategyService creates a new trading strategy service
func NewTradingStrategyService() *TradingStrategyService {
	return &TradingStrategyService{}
}

// InitializeStrategy creates a new trading strategy with the given parameters
func (s *TradingStrategyService) InitializeStrategy(ctx restate.ObjectContext, request models.StrategyRequest) (*models.StrategyResponse, error) {
	// Generate a unique strategy ID using the object key
	strategyId := restate.Key(ctx)

	// Create strategy with initial status
	strategy := models.TradingStrategy{
		ID:             strategyId,
		OrderRequest:   request.OrderRequest,
		TargetPrice:    request.TargetPrice,
		PriceCondition: request.PriceCondition,
		Status:         models.StrategyInitialized,
		CurrentPrice:   0.0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Store the strategy in the virtual object's state
	restate.Set(ctx, "strategy", strategy)

	slog.Info("Strategy initialized",
		"strategy_id", strategy.ID,
		"symbol", strategy.OrderRequest.Symbol,
		"target_price", strategy.TargetPrice,
		"condition", strategy.PriceCondition)

	return &models.StrategyResponse{Strategy: strategy}, nil
}

// GetStrategy retrieves the current strategy state
func (s *TradingStrategyService) GetStrategy(ctx restate.ObjectSharedContext) (*models.StrategyResponse, error) {
	strategy, err := restate.Get[models.TradingStrategy](ctx, "strategy")
	if err != nil {
		return nil, err
	}

	return &models.StrategyResponse{Strategy: strategy}, nil
}

// ProcessPriceSignal handles a new price signal and updates the strategy state
func (s *TradingStrategyService) ProcessPriceSignal(ctx restate.ObjectContext, signal models.PriceSignal) (*models.StrategyResponse, error) {
	// Retrieve the current strategy
	strategy, err := restate.Get[models.TradingStrategy](ctx, "strategy")
	if err != nil {
		return nil, oops.Code("strategy_error").In("strategy_handler").
			Wrapf(err, "Failed to get strategy state")
	}

	// Verify signal is for the correct symbol
	if signal.Symbol != strategy.OrderRequest.Symbol {
		return nil, oops.Code("invalid_signal").In("strategy_handler").
			With("expected_symbol", strategy.OrderRequest.Symbol).
			With("actual_symbol", signal.Symbol).
			Errorf("Price signal symbol mismatch")
	}

	// Update current price
	strategy.CurrentPrice = signal.Price
	strategy.UpdatedAt = time.Now()

	// Check if the strategy should be armed
	if strategy.Status == models.StrategyInitialized {
		slog.Info("Strategy armed - waiting for price condition",
			"strategy_id", strategy.ID,
			"current_price", signal.Price,
			"target_price", strategy.TargetPrice,
			"condition", strategy.PriceCondition)
		strategy.Status = models.StrategyArmed
	}

	// Check if price condition is met
	conditionMet := false
	switch strategy.PriceCondition {
	case models.PriceAbove:
		conditionMet = signal.Price >= strategy.TargetPrice
	case models.PriceBelow:
		conditionMet = signal.Price <= strategy.TargetPrice
	}

	// If strategy is armed and condition is met, execute the order
	if strategy.Status == models.StrategyArmed && conditionMet {
		slog.Info("Price condition met - executing order",
			"strategy_id", strategy.ID,
			"symbol", strategy.OrderRequest.Symbol,
			"current_price", signal.Price,
			"target_price", strategy.TargetPrice)

		strategy.Status = models.StrategyTriggered

		// Store updated strategy state
		restate.Set(ctx, "strategy", strategy)

		// Execute the order
		orderResponse, err := restate.Service[*models.OrderResponse](ctx, "OrderService", "ProcessOrder").
			Request(strategy.OrderRequest)

		if err != nil {
			strategy.Status = models.StrategyFailed
			restate.Set(ctx, "strategy", strategy)

			return nil, oops.Code("order_execution_error").In("strategy_handler").
				With("strategy_id", strategy.ID).
				Wrapf(err, "Failed to execute order")
		}

		// Update strategy with executed order
		strategy.ExecutedOrder = &orderResponse.Order
		strategy.Status = models.StrategyExecuted

		if orderResponse.Order.Status == models.Settled {
			strategy.Status = models.StrategyCompleted
			slog.Info("Strategy completed successfully",
				"strategy_id", strategy.ID,
				"order_id", orderResponse.Order.ID,
				"execution_price", signal.Price)
		}
	}

	// Store updated strategy
	restate.Set(ctx, "strategy", strategy)

	return &models.StrategyResponse{Strategy: strategy}, nil
}
