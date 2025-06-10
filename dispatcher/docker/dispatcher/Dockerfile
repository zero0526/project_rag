# Stage 1: Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/dispatcher/ ./cmd/dispatcher/
COPY internal/ ./internal/
COPY configs/ ./configs/
RUN CGO_ENABLED=0 GOOS=linux go build -o /dispatcher ./cmd/dispatcher/main.go

# Stage 2: Runtime stage
FROM alpine:3.18
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /dispatcher /app/dispatcher
COPY --from=builder /app/configs/ /app/configs/
CMD ["/app/dispatcher"]