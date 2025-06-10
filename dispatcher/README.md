[Kafka topic: input-url-topic]     (URLs per domain)
              │
              ▼
   ┌────────────────────────┐
   │      Dispatcher        │
   │ (Go, Kafka consumer)   │
   └────────────────────────┘
              │
              ▼ HTTP POST /crawl
 ┌────────────────────────────────┐
 │           Crawler             │
 │  (Go, Colly, nhận URL,        │
 │   load config selector,       │
 │   crawl HTML & extract data)  │
 └────────────────────────────────┘
        │               │
        ▼               ▼
[Kafka: article_raw_data]    [Kafka: retry-url-topic]
(Payload structured)         (URL với TTL-- nếu lỗi)


distributed-crawler/
├── cmd/
│   ├── dispatcher/              # App: Kafka consumer, gửi HTTP request tới crawler
│   │   └── main.go
│   └── crawler/                 # App: HTTP server nhận URL, crawl, push Kafka
│       └── main.go
│
├── internal/
│   ├── config/                  # Load crawler.yaml và selectors.json
│   │   ├── loader.go
│   │   └── config.go
│   ├── kafka/                   # Kafka producer wrapper
│   │   ├── producer.go
│   │   └── consumer.go
│   └── crawler/                 # Logic crawling HTML
│       └── crawler.go
│
├── configs/
│   ├── crawler.yaml             # Cấu hình Kafka, HTTP port, retry TTL
│   └── selectors.json           # Các domain và selector để extract title, content,...
│
├── docker/
│   ├── dispatcher/
│   │   └── Dockerfile
│   └── crawler/
│       └── Dockerfile
│
├── deploy/
│   └── docker-compose.yml       # Dàn Kafka, Dispatcher, Crawler
│
├── go.mod
└── README.md


distributed-crawler/
├── cmd/
│   ├── dispatcher/
│   │   └── main.go              # Thêm Prometheus metrics + circuit breaker logic
│   └── crawler/
│       └── main.go              # Thêm Prometheus metrics
│
├── internal/
│   ├── breaker/                 # Circuit breaker logic
│   │   └── breaker.go
│   ├── config/
│   ├── kafka/
│   ├── crawler/
│   └── metrics/                 # Prometheus metrics định nghĩa và ghi nhận
│       └── metrics.go
│
├── configs/
│   ├── crawler.yaml
│   ├── selectors.json
│   └── prometheus.yml           # Prometheus scrape config
│
├── docker/
│   ├── dispatcher/
│   ├── crawler/
│   └── monitoring/              # Dockerfile cho Prometheus và Grafana (nếu cần)
│       ├── prometheus/
│       │   └── prometheus.yml
│       └── grafana/
│           └── dashboards/
│               └── crawler-dashboard.json
│
├── deploy/
│   └── docker-compose.yml       # Thêm service: prometheus, grafana
│
├── go.mod
└── README.md


<!-- crawler -->
FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o crawler ./cmd/crawler

COPY configs/ ./configs/

CMD ["./crawler"]

<!-- dispathcer  -->
FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o dispatcher ./cmd/dispatcher

COPY configs/ ./configs/

CMD ["./dispatcher"]



version: '3.8'

services:
  zookeeper:
    image: wurstmeister/zookeeper
    ports:
      - "2181:2181"

  kafka:
    image: wurstmeister/kafka
    ports:
      - "9092:9092"
    environment:
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_HOST_NAME: kafka
    depends_on:
      - zookeeper

  dispatcher:
    build:
      context: ..
      dockerfile: docker/dispatcher/Dockerfile
    container_name: dispatcher
    depends_on:
      - kafka
    environment:
      - KAFKA_BROKER=kafka:9092

  crawler:
    build:
      context: ..
      dockerfile: docker/crawler/Dockerfile
    container_name: crawler
    depends_on:
      - kafka
    environment:
      - KAFKA_BROKER=kafka:9092
    ports:
      - "8080:8080"  # để dispatcher gọi được





internal config,go load config từ crawler.yaml và topics.json

internal/kafka/consumer.go  Chứa logic Kafka Consumer, subscribe vào topics, nhận message, và gửi request tới Crawler.


c. cmd/dispatcher/main.go Entry point cho Dispatcher, khởi tạo Consumer và xử lý graceful shutdown.

configs/crawler.yaml  Cấu hình cho Dispatcher, bao gồm thông số retry.


configs/topics.json  Danh sách topic (đã cung cấp).

docker/dispatcher/Dockerfile    Dockerfile để build Dispatcher.

deploy/docker-compose.yml  Cấu hình để chạy toàn bộ hệ thống.



Kafka topics (domain_example_com, ...) 
   ↓
consumer.go (Sarama consumer group)
   ↓
enqueueDomainMsg(domain, msg)
   ↓
domains[domain].queue (buffer channel per domain)
   ↓
batcher.go dispatchLoop() gom batch xen kẽ các domain
   ↓
sendBatch() gửi batch qua HTTP đến crawler replica
   ↓
crawler trả về CrawlerResponse
   ↓
handleCrawlerResponse() phân tích
   ├─ update retry per URL
   ├─ update domain breaker state
   └─ enqueue lại URL nếu cần


