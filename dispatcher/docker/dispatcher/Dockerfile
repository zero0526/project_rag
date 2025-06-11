FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /dispatcher ./cmd/dispatcher

FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /dispatcher .

COPY configs/dispatcher.yaml ./configs/
COPY internal/config/selectors.json ./internal/config/

CMD ["./dispatcher"]