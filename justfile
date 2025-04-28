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

# -1. Justfile like Makefiles but better and with good shell support
# 0. Restate consists of the server and function (lamdas)
#       under the hood (grpc (protobuf) & journals on restate server)
# Restate website kurz zeigen

# 
# Restate Server
#

# Start restate server
[group('restate-server')]
up:
    docker compose up  --detach --remove-orphans

# Stop restate server
[group('restate-server')]
down:
    docker compose down --volumes



# Run the trading system
run:
    go run main.go

# Run the trading system with pretty output
run-pretty:
    go run main.go 2>&1 | jq -C '.'

# 
# FLY 
#


# Launc application
[group('fly')]
launch:
    -fly app destroy --yes miltkall-go-playground 
    fly launch --config fly.toml --local-only --regions "fra,cdg,lhr" --now

# Stream application logs from all instances
[group('fly')]
logs:
    fly logs

# Show current status of the deployed application and its instances
[group('fly')]
status:
    fly status

# 
# Restate
#

# Register services with Restate
[group('restate')]
register HOST="host.docker.internal:9080":
    restate -y deployments register --force {{HOST}}

[group('restate')]
kill INV_ID="":
    restate invocations cancel --kill {{INV_ID}}

# 
# DEMO
#

# 1. Simple run (Explain steps + text between ===>)
# 2. Run it and kill it during execution (cmd just output + UI + show that steps are not rerun)

# Submit a market order
[group('demo')]
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
[group('demo')]
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

# 5. Run it and the next 3 commands (show that commands are getting reverted in case of failure)

# Test order saga processing
[group('demo')]
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

# Submit a saga order that will fail during execution
[group('demo')]
test-saga-fail-execution:
    curlie http://localhost:8080/OrderSagaService/ProcessOrderWithSaga \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "account_id": "acc123", \
      "symbol": "FAIL_DURING_EXECUTION", \
      "quantity": 100, \
      "type": "MARKET", \
      "side": "BUY" \
    }'

# Submit a saga order that will fail during settlement
[group('demo')]
test-saga-fail-settlement:
    curlie http://localhost:8080/OrderSagaService/ProcessOrderWithSaga \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "account_id": "acc123", \
      "symbol": "FAIL_DURING_SETTLEMENT", \
      "quantity": 100, \
      "type": "MARKET", \
      "side": "BUY" \
    }'

# 6. init-strategy + get it + UI
# 7. below + UI
# 8. above => strategy and orders gets executed

# Initialize a trading strategy
[group('demo')]
init-strategy STRATEGY_ID="demo-strategy123":
    curlie http://localhost:8080/TradingStrategyService/{{STRATEGY_ID}}/InitializeStrategy \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "order_request": { \
        "account_id": "acc123", \
        "symbol": "AAPL", \
        "quantity": 100, \
        "type": "MARKET", \
        "side": "BUY" \
      }, \
      "target_price": 185.50, \
      "price_condition": "ABOVE" \
    }'

# Get the current state of a strategy
[group('demo')]
get-strategy STRATEGY_ID="demo-strategy123":
    curlie http://localhost:8080/TradingStrategyService/{{STRATEGY_ID}}/GetStrategy

# Send a price signal below the target (arms the strategy)
[group('demo')]
price-signal-below STRATEGY_ID="demo-strategy123" PRICE="180.25":
    curlie http://localhost:8080/TradingStrategyService/{{STRATEGY_ID}}/ProcessPriceSignal \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "symbol": "AAPL", \
      "price": {{PRICE}} \
    }'

# Send a price signal above the target (should trigger if not already triggered)
[group('demo')]
price-signal-above STRATEGY_ID="demo-strategy123" PRICE="190.75":
    curlie http://localhost:8080/TradingStrategyService/{{STRATEGY_ID}}/ProcessPriceSignal \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "symbol": "AAPL", \
      "price": {{PRICE}} \
    }'

# 9. all together

# Demo workflow - run a complete strategy lifecycle
[group('demo')]
demo-workflow:
    #!/usr/bin/env bash
    STRATEGY_ID="strategy-$(uuidgen)"
    echo "Using strategy ID: $STRATEGY_ID"
    
    echo "\nStep 1: Initialize strategy"
    just init-strategy $STRATEGY_ID
    
    echo "\nStep 2: Check initial state"
    just get-strategy $STRATEGY_ID
    
    echo "\nStep 3: Send price signal below target (arms strategy)"
    just price-signal-below $STRATEGY_ID 180.25
    
    echo "\nStep 4: Check armed state"
    just get-strategy $STRATEGY_ID
    
    echo "\nStep 5: Send price signal above target (triggers order execution)"
    just price-signal-above $STRATEGY_ID 186.25
    
    echo "\nStep 6: Check final state"
    just get-strategy $STRATEGY_ID

# 10.5. OpenAPI + Playground
# UI, pick one service and show it

# Get openAPI
[group('demo')]
openapi:
    curl -s localhost:9070/services/OrderService/openapi | yq

# 10. fly
# Fly.io (suspention no cost when markets are closed, round rodin load-balancer on the instances, high availability for multiple regions)

# 11. Query 
# Query DB

#
# CONCLUSION
#

# 
# Value proposition distributed service without having PHD (true!)
# We concentrate on writing functions (also easily testable) rather then maintaining infra + monoliths!
#
# Additional use of intenal state and objects
# Many more examples https://github.com/restatedev/examples
# Provide clients/openAPI spec to query current state of objects + invocations => useful frontend dev
# 
# We can use different languages (python, go, rust) !!! => Lots of python trading libs
#
# neonDB + wireguard (tailscale) and you have a complete stack!
