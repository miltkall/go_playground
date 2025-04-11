package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/miltkall/go_playground/internal/handlers"
	restate "github.com/restatedev/sdk-go"
	"github.com/restatedev/sdk-go/server"
	"github.com/samber/oops"
)

func main() {
	// Configure structured logging
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	// Create service handlers
	orderService := handlers.NewOrderService()
	orderSagaService := handlers.NewOrderSagaService()
	tradingStrategyService := handlers.NewTradingStrategyService() // New service

	// Create Restate server
	restateServer := server.NewRestate().
		// Register the services
		Bind(restate.Reflect(orderService)).
		Bind(restate.Reflect(orderSagaService)).
		Bind(restate.Reflect(tradingStrategyService)) // Add trading strategy service

	// Configure Restate authentication if key is provided
	if key := os.Getenv("RESTATE_PUBLIC_KEY"); key != "" {
		logger.Info("Configuring Restate with identity key")
		restateServer = restateServer.WithIdentityV1(key)
	}

	// Configure server port
	port := 9080
	if portStr := os.Getenv("RESTATE_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		} else {
			logger.Warn("Failed to parse RESTATE_PORT",
				"error", err.Error())
		}
	}
	address := fmt.Sprintf(":%d", port)

	logger.Info("Starting trading system", "address", address)
	logger.Info("Endpoints available at",
		"processOrder", fmt.Sprintf("http://localhost:%d/OrderService/ProcessOrder", port),
		"sagaOrderProcessing", fmt.Sprintf("http://localhost:%d/OrderSagaService/ProcessOrderWithSaga", port),
		"strategyInit", fmt.Sprintf("http://localhost:%d/TradingStrategyService/{strategyId}/InitializeStrategy", port),
		"priceSignal", fmt.Sprintf("http://localhost:%d/TradingStrategyService/{strategyId}/ProcessPriceSignal", port))

	// Start the server with error handling
	err := restateServer.Start(context.Background(), address)
	if err != nil {
		appError := oops.
			Code("server_startup_error").
			In("main").
			With("address", address).
			Wrapf(err, "Application exited unexpectedly")

		logger.Error(appError.Error(),
			"error", err,
			"error_context", appError)
		os.Exit(1)
	}
}
