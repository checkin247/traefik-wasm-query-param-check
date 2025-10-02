# Dockerfile
FROM golang:1.25.1-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY src/go.mod ./

# Download dependencies and create go.sum
RUN go mod tidy

# Copy source code
COPY src/ ./

# Build WASM plugin using standard Go compiler with wasip1 target
RUN CGO_ENABLED=0 GOOS=wasip1 GOARCH=wasm go build -o /plugin.wasm main.go

# Final stage - minimal image
FROM scratch AS final
COPY --from=builder /plugin.wasm /plugin.wasm
