package dispatch

import (
	"bytes"
	"dispatcher/internal/breaker"
	"dispatcher/internal/config"
	"dispatcher/internal/kafka"
	"dispatcher/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sony/gobreaker"
)


type Sender struct {
	State   *DispatcherState
	Breaker *Breaker
	Retry   *Retry
	numWindow int
	rateError float64
	CrawlerBreaker map[string]*gobreaker.CircuitBreaker
	Batcher *Batcher
}
var (
    RedisClient  *producer.RedisClient
    KafkaProducer *producer.Producer
	cfg *config.Config
	MaxQueueSize int = 100
)
type QueueStatus struct {
	QueueLength int `json:"queue_length"`
}
func NewSender(state *DispatcherState, breaker *Breaker, retry *Retry,Batcher *Batcher, numWindow int, rateError float64) *Sender {
	var err error
	cfg, err = config.LoadConfig("../../configs/dispatcher.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	KafkaProducer, err = producer.NewProducer(cfg.Kafka.Brokers) 
	if err != nil{
		log.Fatalf("fail to create new producer")
	} 
	RedisClient = producer.NewRedisClient("localhost", 6379, "")
	crawlerBreakers := make(map[string]*gobreaker.CircuitBreaker)
	for _,cb := range(cfg.Dispatcher.URL2Cralwers){
		crawlerId,_ := getCrawlerId(cb)
		crawlerBreakers[cb] = breakerCrawler.NewCrawlerBreaker(crawlerId)
	} 
	return &Sender{State: state, Breaker: breaker,CrawlerBreaker: crawlerBreakers,Batcher:Batcher, Retry: retry, numWindow: numWindow, rateError: rateError}
}

// SendBatch sends a batch to the crawler and processes the response.
func (s *Sender) DispatchToCrawlers() {
	for _, crawlerURL := range cfg.Dispatcher.URL2Cralwers {
		go func(crawlerURL string) {
			queueURL := fmt.Sprintf("%s/queue", crawlerURL)
			resp, err := http.Get(queueURL)
			if err != nil {
				log.Printf("[Dispatch] Failed to check queue: %v", err)
				return
			}
			defer resp.Body.Close()

			var qStat QueueStatus
			if err := json.NewDecoder(resp.Body).Decode(&qStat); err != nil {
				log.Printf("[Dispatch] Failed to decode queue response: %v", err)
				return
			}

			if qStat.QueueLength+ s.Batcher.NumInBatch >= MaxQueueSize {
				return // Crawler đang bận
			}

			batch := s.Batcher.CreateBatch()
			if len(batch) == 0 {
				return // Hết data trong Kafka
			}

			body, err := json.Marshal(batch)
			if err != nil {
				log.Printf("[Dispatch] Failed to marshal batch: %v", err)
				return
			}

			cb := s.CrawlerBreaker[crawlerURL]
			if cb == nil {
				log.Printf("[Dispatch] No circuit breaker for crawler: %s", crawlerURL)
				return
			}

			result, err := cb.Execute(func() (interface{}, error) {
				postURL := fmt.Sprintf("%s/enqueue", crawlerURL)
				resp, err := http.Post(postURL, "application/json", bytes.NewBuffer(body))
				if err != nil {
					s.Retry.QueueForRetry(batch)
					return nil, err
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusTooManyRequests {
					s.Retry.QueueForRetry(batch)
					return nil, fmt.Errorf("too many requests (429) from crawler: %s", crawlerURL)
				}

				var crawlerResp CrawlerResponse
				if err := json.NewDecoder(resp.Body).Decode(&crawlerResp); err != nil {
					return nil, err
				}

				if crawlerResp.Status != "ok" {
					var crawledURL []MessageModel
					for _, urlRs := range crawlerResp.URLResults {
						crawledURL = append(crawledURL, MessageModel{Category: urlRs.Category, AbsURL: urlRs.URL})
					}
					s.Retry.QueueForRetry(crawledURL)
					return crawlerResp, fmt.Errorf("crawler responded with status: %s", crawlerResp.Status)
				}

				return crawlerResp, nil
			})

			if err != nil {
				log.Printf("[Dispatch] Error from circuit breaker execution: %v", err)
				return
			}

			crawlerResp, ok := result.(CrawlerResponse)
			if !ok {
				log.Printf("[Dispatch] Failed to cast result to CrawlerResponse")
				return
			}

			s.updateBatchWindow(crawlerResp)
			s.processURLResults(crawlerResp.URLResults)

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
							Category: urlrs.Category,
							AbsURL:   url,
						})
					}
				}
			}
		}(crawlerURL)
	}
}

func (s *Sender) updateBatchWindow(resp CrawlerResponse) {
	s.State.WindowMu.Lock()
	defer s.State.WindowMu.Unlock()

	domainResults := make(map[string]map[string]int)
	for _, result := range resp.URLResults {
		domain, err := util.ExtractDomain(result.URL)
		if err != nil {
			continue
		}
		domain = util.FormatKey(domain)
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

func (s *Sender) processURLResults(results []URLResult) {
	for _, result := range results {
		domain, err := util.ExtractDomain(result.URL)
		if err != nil {
			continue
		}
		domain = util.FormatKey(domain)
		if result.Status == "failed" {
			s.Breaker.RecordFailure(domain)
			s.Retry.QueueForRetry([]MessageModel{{Category: result.Category, AbsURL: result.URL}})
		} else if result.Status == "blocked" {
			s.Breaker.BlockDomain(domain)
		} else if result.Status == "success" {
			s.Breaker.RecordSuccess(domain)
		}
	}
}

func getCrawlerId(urlString string) (string, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return "", err
	}

	port := u.Port()
	return port, nil
}
