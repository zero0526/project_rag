package breakerCrawler

import (
    "github.com/sony/gobreaker"
    "time"
    "fmt"
)

func NewCrawlerBreaker(crawlerID string) *gobreaker.CircuitBreaker {
    return gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        fmt.Sprintf("crawler-breaker-%s", crawlerID),
        MaxRequests: 3, // Số request tối đa ở trạng thái Half-Open
        Interval:    60 * time.Second, // Thời gian reset số liệu lỗi
        Timeout:     30 * time.Second, // Thời gian chờ trước khi chuyển sang Open
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            return counts.ConsecutiveFailures >= 3 // Chuyển sang Open sau 3 lỗi liên tiếp
        },
    })
}