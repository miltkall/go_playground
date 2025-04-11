ARG GO_VERSION=1.24.1
FROM golang:${GO_VERSION}-bookworm as builder

# ENV GOPRIVATE=github.com/miltkall/*

WORKDIR /usr/src/app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./

# Configure git for private repositories, remove replace directives for fresh fetch
# RUN --mount=type=secret,id=GH_TOKEN \
#     git config --global url."https://$(cat /run/secrets/GH_TOKEN)@github.com/".insteadOf "https://github.com/" && \
#     sed -i '/^replace/d' go.mod && \ 
#     go mod tidy
RUN --mount=type=cache,target=/go/pkg/mod go mod tidy

# Copy the rest of the source code
COPY . .

# Build with cached modules, replace directives for fresh fetch
# Note: Using -buildvcs=true to include VCS information
RUN --mount=type=cache,target=/go/pkg/mod \
    sed -i '/^replace/d' go.mod && \ 
    CGO_ENABLED=0 GOOS=linux go build -buildvcs=true -v -o /run-app .

# Minimal deployment image
FROM gcr.io/distroless/static-debian12
COPY --from=builder /run-app /run-app
ENTRYPOINT ["/run-app"]

