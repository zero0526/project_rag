package dispatch

import (
	"sync"
	"time"
)

// MessageModel represents a message from Kafka.
type MessageModel struct {
	Category string `json:"category"`
	AbsURL   string `json:"absUrl"`
}

// CrawlerResponse represents the response from the crawler.
type CrawlerResponse struct {
	Status     string           `json:"status"` 
	URLResults []URLResult      `json:"urlResults"`
}

// URLResult represents the status of a crawled URL.
type URLResult struct {
	URL    string `json:"url"`
	Status string `json:"status"` // "success", "failed", "blocked"
	RelativeURL []string `json:"reltiveURL"`
}

// BatchState tracks the state of a batch for failure rate analysis.
type BatchState struct {
	DomainResults map[string]map[string]int // domain -> status -> count
	Timestamp     time.Time
}

// DispatcherState holds the global state.
type DispatcherState struct {
	DomainMessages sync.Map // key: rootURL (dots replaced with _), value: []MessageModel
	RetryMessages  sync.Map // key: rootURL, value: *sync.Map (url -> RetryMeta)
	BatchWindow    []BatchState
	WindowMu       sync.Mutex
}

type RetryMeta struct {
	Count      int
	FirstRetry time.Time
}