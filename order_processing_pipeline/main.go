package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/miltkall/components/component_email_service/internal/handler"
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
	emailService := handler.NewEmailService()
	batchEmailService := &handler.BatchEmailService{}

	// Create Restate server
	restateServer := server.NewRestate().
		// Register the email service
		Bind(restate.Reflect(emailService)).
		// Register the batch email service
		Bind(restate.Reflect(batchEmailService))

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

	logger.Info("Starting email service", "address", address)
	logger.Info("Email service endpoints available at",
		"sendEmail", fmt.Sprintf("http://localhost:%d/EmailService/SendEmail", port),
		"batchEmails", fmt.Sprintf("http://localhost:%d/BatchEmailService/ProcessBatch", port))

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
