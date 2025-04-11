# Default recipe that displays available commands, restate and fly status
[private]
default:
    @just --list
    @echo
    @echo Restate:
    @-restate services status
    @echo
    @echo Fly:
    @-fly status

# Run the trading system
run:
    go run main.go

# Register services with Restate
register:
    restate -y deployments register host.docker.internal:9080

# Submit a market order
[group('test')]
test-market-order:
    curl http://localhost:8080/OrderService/ProcessOrder \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "account_id": "acc123", \
      "symbol": "AAPL", \
      "quantity": 100, \
      "type": "MARKET", \
      "side": "BUY" \
    }'

# Submit a limit order
[group('test')]
test-limit-order:
    curl http://localhost:8080/OrderService/ProcessOrder \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "account_id": "acc123", \
      "symbol": "GOOGL", \
      "quantity": 50, \
      "price": 150.0, \
      "type": "LIMIT", \
      "side": "SELL" \
    }'

# Submit an invalid order (will be rejected)
[group('test')]
test-invalid-order:
    curl http://localhost:8080/OrderService/ProcessOrder \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "account_id": "acc123", \
      "symbol": "INVALID", \
      "quantity": 100, \
      "type": "MARKET", \
      "side": "BUY" \
    }'

# Setup env
[group('dependencies')]
setup-env:
    go env -w GOPRIVATE="github.com/miltkall/*"


# Start restate server
[group('restate')]
up:
    docker compose up  --detach --remove-orphans

# Stop restate server
[group('restate')]
down:
    docker compose down
