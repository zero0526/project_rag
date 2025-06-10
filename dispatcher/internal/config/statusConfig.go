package config

import (
	"time"
)

type URLStatus string
const (
    Pending    URLStatus = "pending"
    InProgress URLStatus = "in-progress"
    Success    URLStatus = "success"
    Failed     URLStatus = "failed"
    Retry      URLStatus = "retry"
    Blocked    URLStatus = "blocked"
)

type URLTask struct {
    RequestID string
    URL       string
    Domain    string
    Status    URLStatus
    RetryCount int
    LastError  string
    AssignedCrawler string
    Timestamp time.Time
}

type DomainState struct {
    Domain        string
    FailCount     int
    TotalCount    int
    FailRate      float64
    IsBlocked     bool
    BlockExpireAt time.Time
    BackoffDelay  time.Duration
}

type CrawlerState struct {
    CrawlerID   string
    InProgress  int
    PendingQueue []*URLTask
}
