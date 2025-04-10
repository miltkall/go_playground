# Show help and list available commands
[private]
@default:
    just --list
    @echo
    @echo Restate:
    @-restate services status

# Setup env
[group('dependencies')]
setup-env:
    go env -w GOPRIVATE="github.com/miltkall/*"


# Setup env
[group('restate')]
up:
    docker compose up  --detach --remove-orphans

[group('restate')]
down:
    docker compose down


