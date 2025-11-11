# Multi-stage build for gcal-cli
FROM golang:1.23-alpine AS builder

# Install git for version info
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X github.com/btafoya/gcal-cli/internal/commands.Version=${VERSION} \
              -X github.com/btafoya/gcal-cli/internal/commands.Commit=${COMMIT} \
              -X github.com/btafoya/gcal-cli/internal/commands.BuildDate=${BUILD_DATE} \
              -w -s" \
    -o gcal-cli ./cmd/gcal-cli

# Final stage - minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/gcal-cli .

# Create config directory
RUN mkdir -p /root/.config/gcal-cli

ENTRYPOINT ["./gcal-cli"]
CMD ["--help"]
