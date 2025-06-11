FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /crawler ./cmd/crawler

# Stage 2: Runtime
FROM alpine:3.18

WORKDIR /app

# Cài đặt các phụ thuộc runtime (nếu cần, ví dụ: ca-certificates cho HTTPS)
RUN apk add --no-cache ca-certificates

# Sao chép binary từ stage build
COPY --from=builder /crawler .

COPY configs/crawler.yaml ./configs/
COPY docker/crawler/entrypoint.sh .
COPY internal/config/selectors.json ./internal/config/

RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]