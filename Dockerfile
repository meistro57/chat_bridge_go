FROM golang:1.23.4-bullseye AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "1.0.0") && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags "-X github.com/markjamesm/chat-bridge-go/internal/version.Version=${VERSION}" \
        -o /chat-bridge ./...

FROM scratch
WORKDIR /app
COPY --from=builder /chat-bridge /chat-bridge
ENTRYPOINT ["/chat-bridge"]
