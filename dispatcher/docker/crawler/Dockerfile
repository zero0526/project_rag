# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/crawler/ ./cmd/crawler/
COPY internal/ ./internal/

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /crawler ./cmd/crawler/main.go

# Stage 2: Minimal runtime image
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /crawler /app/crawler

EXPOSE 8080
COPY docker/crawler/entrypoint.sh ./entrypoint.sh
RUN chmod +x entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

