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

# Setup env
[group('dependencies')]
setup-env:
    go env -w GOPRIVATE="github.com/miltkall/*"


# Run the trading system
run:
    go run main.go

# Register services with Restate
[group('restate')]
register:
    restate -y deployments register --force host.docker.internal:9080

[group('restate')]
kill inv_id="":
    restate invocations cancel --kill {{inv_id}}



# 1. Simple run (Explain steps + text between ===>)
# 2. Run it and kill it during execution (cmd just output + UI + show that steps are not rerun)

# Submit a market order
[group('test')]
test-passing-order:
    curlie http://localhost:8080/OrderService/ProcessOrder \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "account_id": "acc123", \
      "symbol": "AAPL", \
      "quantity": 100, \
      "type": "MARKET", \
      "side": "BUY" \
    }'

# 3. Run it (show it is failing in the UI, kill it in UI, show CMD output)
# 4. the same kill it from the CDM

# Submit an order that will fail during settlement
[group('test')]
test-failing-order:
    curlie http://localhost:8080/OrderService/ProcessOrder \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "account_id": "acc123", \
      "symbol": "ERROR", \
      "quantity": 100, \
      "type": "MARKET", \
      "side": "BUY" \
    }'

# Test order saga processing
[group('test')]
test-saga:
    curl http://localhost:8080/OrderSagaService/ProcessOrderWithSaga \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "account_id": "acc123", \
      "symbol": "AAPL", \
      "quantity": 100, \
      "type": "MARKET", \
      "side": "BUY" \
    }'

# Virtual Objs
# OpenAPI spec
# Fly.io (suspention no cost when markets are closed, round rodin load-balancer on the instances)
# Query DB


# Start restate server
[group('restate-server')]
up:
    docker compose up  --detach --remove-orphans

# Stop restate server
[group('restate-server')]
down:
    docker compose down


# Value proposition distributed service without having PHD (true!)
# We concentrate on writing functions (also easily testable) rather then maintaining infra + monoliths!

# Additional use of intenal state and objects
# Many more examples https://github.com/restatedev/examples
# 
# We can use different languages (python, go, rust)
