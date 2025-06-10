package dispatch

import (
	"bytes"
	"dispatcher/internal/config"
	"dispatcher/internal/kafka"
	"dispatcher/internal/utils"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
	"github.com/sony/gobreaker"
	"dispatcher/internal/breaker"
	"fmt"
)


type Sender struct {
	State   *DispatcherState
	Breaker *Breaker
	Retry   *Retry
	numWindow int
	rateError float64
	CrawlerBreaker map[string]*gobreaker.CircuitBreaker
}
var (
    RedisClient  *producer.RedisClient
    KafkaProducer *producer.Producer
)
func NewSender(state *DispatcherState, breaker *Breaker, retry *Retry, numWindow int, rateError float64) *Sender {
	cfg, err := config.LoadConfig("../../configs/dispatcher.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	KafkaProducer, err = producer.NewProducer(cfg.Kafka.Brokers) 
	if err != nil{
		log.Fatalf("fail to create new producer")
	} 
	RedisClient = producer.NewRedisClient("localhost", 6379, "")
	crawlerBreakers := make(map[string]*gobreaker.CircuitBreaker)
	crawlerBreakers["http://localhost:8081/crawl"] = breakerCrawler.NewCrawlerBreaker("8081")
	return &Sender{State: state, Breaker: breaker,CrawlerBreaker: crawlerBreakers, Retry: retry, numWindow: numWindow, rateError: rateError}
}

// SendBatch sends a batch to the crawler and processes the response.
func (s *Sender) SendBatch(batch []MessageModel) error {
	if len(batch) == 0 {
		return nil
	}

	domain, err := util.ExtractDomain(batch[0].AbsURL)
	if err != nil {
		s.Retry.QueueForRetry(batch)
		return err
	}

	// Check circuit breaker cho domain
	if !s.Breaker.Allow(domain) {
		s.Retry.QueueForRetry(batch)
		return nil
	}

	// Prepare HTTP request body
	body, err := json.Marshal(batch)
	if err != nil {
		s.Retry.QueueForRetry(batch)
		return err
	}

	// Crawler URL và breaker riêng
	crawlerURL := "http://localhost:8081/crawl"
	cb, ok := s.CrawlerBreaker[crawlerURL]
	if !ok {
		cb = breakerCrawler.NewCrawlerBreaker("8081")
		s.CrawlerBreaker[crawlerURL] = cb
	}

	// Gọi HTTP POST thông qua breaker
	result, err := cb.Execute(func() (interface{}, error) {
		resp, err := http.Post(crawlerURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var crawlerResp CrawlerResponse
		if err := json.NewDecoder(resp.Body).Decode(&crawlerResp); err != nil {
			return nil, err
		}

		if crawlerResp.Status != "ok" {
			return crawlerResp, fmt.Errorf("crawler responded with status: %s", crawlerResp.Status)
		}

		return crawlerResp, nil
	})

	// Nếu có lỗi từ breaker (network hoặc logic từ crawler)
	if err != nil {
		s.Breaker.RecordFailure(domain)
		s.Retry.QueueForRetry(batch)
		return err
	}

	// Đảm bảo ép kiểu thành công
	crawlerResp := result.(CrawlerResponse)

	// Thành công → record success
	s.Breaker.RecordSuccess(domain)

	// Đẩy các RelativeURL mới vào Kafka nếu là batch crawl
	for _, urlrs := range crawlerResp.URLResults {
		for _, url := range urlrs.RelativeURL {
			if chk, err := RedisClient.AddConfig(url); chk && err == nil {
				app, err := util.ExtractDomain(url)
				if err != nil {
					log.Printf("error extracting domain: %v", err)
					continue
				}
				topic := strings.ReplaceAll(app, ".", "_")
				KafkaProducer.Send(topic, producer.MessageModel{
					Category: "",
					AbsURL:   url,
				})
			}
		}
	}

	// Cập nhật thống kê và xử lý kết quả URL
	s.updateBatchWindow(crawlerResp, batch)
	s.processURLResults(crawlerResp.URLResults)

	return nil
}

// updateBatchWindow tracks the last 3 batches for failure rate analysis.
func (s *Sender) updateBatchWindow(resp CrawlerResponse, batch []MessageModel) {
	s.State.WindowMu.Lock()
	defer s.State.WindowMu.Unlock()

	domainResults := make(map[string]map[string]int)
	for _, result := range resp.URLResults {
		domain, err := util.ExtractDomain(result.URL)
		if err != nil {
			continue
		}
		if _, exists := domainResults[domain]; !exists {
			domainResults[domain] = make(map[string]int)
		}
		domainResults[domain][result.Status]++
	}

	s.State.BatchWindow = append(s.State.BatchWindow, BatchState{
		DomainResults: domainResults,
		Timestamp:     time.Now(),
	})

	if len(s.State.BatchWindow) > s.numWindow {
		s.State.BatchWindow = s.State.BatchWindow[1:]
	}

	s.checkFailureRate()
}

// checkFailureRate blocks domains with high failure rates.
func (s *Sender) checkFailureRate() {
	domainStats := make(map[string]map[string]int)
	for _, batch := range s.State.BatchWindow {
		for domain, results := range batch.DomainResults {
			if _, exists := domainStats[domain]; !exists {
				domainStats[domain] = make(map[string]int)
			}
			for status, count := range results {
				domainStats[domain][status] += count
			}
		}
	}

	for domain, stats := range domainStats {
		total := stats["success"] + stats["failed"] + stats["blocked"]
		if total == 0 {
			continue
		}
		failureRate := float64(stats["failed"]+stats["blocked"]) / float64(total)
		if failureRate > s.rateError {
			s.Breaker.BlockDomain(domain)
		}
	}
}

// processURLResults handles individual URL results.
func (s *Sender) processURLResults(results []URLResult) {
	for _, result := range results {
		if result.Status == "failed" {
			s.Retry.QueueForRetry([]MessageModel{{AbsURL: result.URL}})
		} else if result.Status == "blocked" {
			domain, err := util.ExtractDomain(result.URL)
			if err == nil {
				s.Breaker.BlockDomain(domain)
			}
		}
	}
}
