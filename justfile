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

# 
# FLY 
#

# Deploy the application to Fly.io using the configuration in fly.toml
[group('fly')]
deploy:
    flyctl deploy --build-secret GH_TOKEN=${GITHUB_TOKEN}

# Create new application
[group('fly')]
create app-name:
    flyctl apps create {{app-name}}


# Stream application logs from all instances
[group('fly')]
logs:
    flyctl logs

# List all secrets currently set for the application
[group('fly')]
secrets-list:
    flyctl secrets list

# Import DATABASE_URL secret
[group('fly')]
secrets-import-db-url:
    fly secrets set DATABASE_URL=${DATABASE_URL}

# Show current status of the deployed application and its instances
[group('fly')]
status:
    flyctl status

# 
# Restate
#

# Register services with Restate
[group('restate')]
register:
    restate -y deployments register --force host.docker.internal:9080

[group('restate')]
kill inv_id="":
    restate invocations cancel --kill {{inv_id}}

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
    docker compose down




# Virtual Objs

# OpenAPI spec
# Fly.io (suspention no cost when markets are closed, round rodin load-balancer on the instances)
# Query DB


# Value proposition distributed service without having PHD (true!)
# We concentrate on writing functions (also easily testable) rather then maintaining infra + monoliths!

# Additional use of intenal state and objects
# Many more examples https://github.com/restatedev/examples
# 
# We can use different languages (python, go, rust)
