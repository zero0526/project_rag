package main

import (
    "dispatcher/internal/crawler"
	"dispatcher/internal/config"
	"dispatcher/internal/dispatcher" 
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
	"fmt"
	"strconv"
	"os"
)

type QueueStatus struct {
	QueueLength int `json:"queue_length"`
}

var (
	rules         map[string]config.CrawlerRule
	redisClient   *crawler.RedisClient
	MaxQueueSize = 100

	queue         []dispatch.MessageModel
	queueMu       sync.Mutex

	processedQueue []dispatch.URLResult
	processedMu    sync.Mutex
)
func main() {
	var err error
	rules, err = config.LoadCrawlerRules("../../internal/config/selectors.json")
	if err != nil {
		log.Fatalf("Failed to load crawler rules: %v", err)
	}

	redisClient = crawler.NewRedisClient("news_crawler_redis:6379")

	go processQueue()

	http.HandleFunc("/enqueue", enqueueHandler)
	http.HandleFunc("/queue", queueStatusHandler)
	portStr := os.Getenv("CRAWLER_PORT")
	if portStr == "" {
		portStr = "8081" 
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid port in CRAWLER_PORT: %v", err)
	}
	log.Printf("Crawler running at :%d",port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Crawler server failed: %v", err)
	}
}

// POST /enqueue
func enqueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var batch []dispatch.MessageModel
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Kiểm tra kích thước queue
	queueMu.Lock()
	if len(queue)+len(batch) > MaxQueueSize {
		queueMu.Unlock()
		http.Error(w, "Queue is full, please try again later", http.StatusTooManyRequests)
		return
	}
	queue = append(queue, batch...)
	queueMu.Unlock()

	// Trả kết quả đã xử lý trước đó
	processedMu.Lock()
	results := make([]dispatch.URLResult, len(processedQueue))
	copy(results, processedQueue)
	processedQueue = processedQueue[:0]
	processedMu.Unlock()

	resp := dispatch.CrawlerResponse{
		Status:     "ok",
		URLResults: results,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Crawler failed to encode response: %v", err)
	}
}

func queueStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	queueMu.Lock()
	count := len(queue)
	queueMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"queue_length": count,
	})
}

func processQueue() {
	for {
		queueMu.Lock()
		if len(queue) == 0 {
			queueMu.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}

		msg := queue[0]
		queue = queue[1:]
		queueMu.Unlock()

		// log.Printf("Processing: %s", msg.AbsURL)
		res := crawler.CrawlPage(msg.AbsURL, msg.Category, rules, redisClient)

		result := dispatch.URLResult{
			URL:         msg.AbsURL,
			Status:      res.Status,
			RelativeURL: res.RelatedLinks,
			Category:  msg.Category,
		}

		// Đưa kết quả xử lý vào processedQueue
		processedMu.Lock()
		processedQueue = append(processedQueue, result)
		processedMu.Unlock()

		// log.Printf("Done: %s -> %s", msg.AbsURL, res.Status)
	}
}


