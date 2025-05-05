# Default recipe that displays available commands, restate and fly status
[private]
default:
    @just --list
    @echo
    @echo Restate:
    @-restate services status
    @echo

# @echo Fly:
# @-fly status

# Setup env
[group('dependencies')]
setup-env:
    go env -w GOPRIVATE="github.com/miltkall/*"

# -1. Justfile like Makefiles but better and with good shell support & variables
# 0. Restate consists of the server and function (lamdas)
#       under the hood (grpc (protobuf) & journals on restate server)
# 
# Restate website kurz zeigen
# 

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

# 
# Service
#

# Run the trading system service
[group('service')]
run:
    go run main.go

# Run the trading system service with pretty output
[group('service')]
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

# Register golang service to restart server
[group('restate')]
register-go HOST="host.docker.internal:9080":
    restate -y deployments register --force {{HOST}}

# Register python service to restart server
[group('restate')]
register-python HOST="host.docker.internal:9081":
    restate -y deployments register --force {{HOST}}

# Kill invocation
[group('restate')]
kill INV_ID="":
    restate invocations cancel --kill {{INV_ID}}

# 
# DEMO
#

# 1. Simple run (Explain steps + text between ===>)
# 2. Run it and kill it during execution (cmd just output + show that steps are not rerun)

# Submit a market order
[group('demo-order')]
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

# 3. Run it (show it is failing in the UI (:9070), kill it in UI, show CMD output)
# (4. the same kill it from the CDM)

# Submit an order that will fail during settlement
[group('demo-order')]
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
[group('demo-saga')]
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
[group('demo-saga')]
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
[group('demo-saga')]
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
# 8. above => strategy and orders gets executed => THE STRATEGY can very easily be a StVE

# Initialize a trading strategy
[group('demo-strategy')]
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
[group('demo-strategy')]
get-strategy STRATEGY_ID="demo-strategy123":
    curlie http://localhost:8080/TradingStrategyService/{{STRATEGY_ID}}/GetStrategy

# Send a price signal below the target (arms the strategy)
[group('demo-strategy')]
price-signal-below STRATEGY_ID="demo-strategy123" PRICE="180.25":
    curlie http://localhost:8080/TradingStrategyService/{{STRATEGY_ID}}/ProcessPriceSignal \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "symbol": "AAPL", \
      "price": {{PRICE}} \
    }'

# Send a price signal above the target (should trigger if not already triggered)
[group('demo-strategy')]
price-signal-above STRATEGY_ID="demo-strategy123" PRICE="190.75":
    curlie http://localhost:8080/TradingStrategyService/{{STRATEGY_ID}}/ProcessPriceSignal \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "symbol": "AAPL", \
      "price": {{PRICE}} \
    }'

# 9. all together (show UI!)

# Demo workflow - run a complete strategy lifecycle
[group('demo-workflow')]
demo-workflow:
    #!/usr/bin/env bash
    STRATEGY_ID="strategy-$(uuidgen)"
    echo "Using strategy ID: $STRATEGY_ID"
    echo
    echo "===> Step 1: Initialize strategy"
    just init-strategy $STRATEGY_ID
    echo
    echo "===> Step 2: Check initial state"
    just get-strategy $STRATEGY_ID
    echo
    echo "===> Step 3: Send price signal below target (arms strategy)"
    just price-signal-below $STRATEGY_ID 180.25
    echo
    echo "===> Step 4: Check armed state"
    just get-strategy $STRATEGY_ID
    echo
    echo "===> Step 5: Send price signal above target (triggers order execution)"
    just price-signal-above $STRATEGY_ID 186.25
    echo
    echo "===> Step 6: Check final state"
    just get-strategy $STRATEGY_ID

# 10.5. OpenAPI + Playground
# UI, pick one service and show it

# Get openAPI
[group('demo-openapi')]
openapi:
    curl -s localhost:9070/services/OrderService/openapi | yq

# 11. Solar ETL
# Test the solar ETL service with valid data
[group('solar-etl')]
test-solar-etl:
    curlie http://localhost:8080/SolarETLService/processSolarData \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "timestamp": "2025-05-04T12:30:00Z", \
      "plant_id": "solar-plant-1", \
      "value": 3500.5 \
    }'

# Test with invalid data (above max)
[group('solar-etl')]
test-solar-etl-high:
    curlie http://localhost:8080/SolarETLService/processSolarData \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "timestamp": "2025-05-04T13:30:00Z", \
      "plant_id": "solar-plant-1", \
      "value": 6000.0 \
    }'

# Test with invalid data (below min)
[group('solar-etl')]
test-solar-etl-low:
    curlie http://localhost:8080/SolarETLService/processSolarData \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "timestamp": "2025-05-04T13:30:00Z", \
      "plant_id": "solar-plant-1", \
      "value": -50.0 \
    }'

# Add a valid value to time series
[group('solar-etl')]
add-time-series-value METRIC="power-output" VALUE="150.5":
    curlie http://localhost:8080/TimeSeriesValidator/{{METRIC}}/addValue \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{ \
      "timestamp": "2025-05-04T12:30:00Z", \
      "metric_name": "{{METRIC}}", \
      "value": {{VALUE}} \
    }'

# Get history for a metric
[group('solar-etl')]
get-time-series-history METRIC="power-output":
    curlie http://localhost:8080/TimeSeriesValidator/{{METRIC}}/getHistory

# Run a complete demo workflow
[group('solar-etl')]
demo-time-series:
    #!/usr/bin/env bash
    METRIC="solar-power"
    echo "Using metric: $METRIC"
    echo
    echo "===> Step 1: Add initial value (50.0)"
    just add-time-series-value $METRIC 50.0
    echo
    echo "===> Step 2: Add second value with small increase (75.5)"
    just add-time-series-value $METRIC 75.5
    echo
    echo "===> Step 3: Add third value with moderate increase (125.0)"
    just add-time-series-value $METRIC 125.0
    echo
    echo "===> Step 4: Try to add implausible value (jump of +250.0)"
    just add-time-series-value $METRIC 375.0
    echo
    echo "===> Step 5: Add another valid value (150.0)"
    just add-time-series-value $METRIC 150.0


#
# CONCLUSION
#

# 
# Value proposition distributed service without having PHD (true!)
#
# We concentrate on writing functions (also easily testable) rather then maintaining infra + monoliths!
# We can also use AWS lamdas => No techstack change
# We can use different languages (python, go, rust) 
#

#
# Additional Trainings:
# AWS Training NO PROBLEM! Kann mich bis Beginn besser für die Stelle mich vorbereiten. 
#
