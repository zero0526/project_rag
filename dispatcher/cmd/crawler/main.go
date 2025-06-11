package main

import (
    "dispatcher/internal/crawler"
	"dispatcher/internal/config"
	"dispatcher/internal/dispatcher" 
	"encoding/json"
	"log"
	"net/http"
)
func main() {
	defer func() {
        if r := recover(); r != nil {
            log.Fatalf("Program panicked: %v", r)
        }
    }()
    rules, err := config.LoadCrawlerRules("../../internal/config/selectors.json")
    if err != nil {
        log.Fatalf("Failed to load crawler rules: %v", err)
    }

    redis := crawler.NewRedisClient("localhost:6380")
	http.HandleFunc("/crawl", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var batch []dispatch.MessageModel
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		log.Printf("Crawler received batch with %d URLs:", len(batch))
		for _, msg := range batch {
			log.Printf("- Category: %s, URL: %s", msg.Category, msg.AbsURL)
		}

		resp := dispatch.CrawlerResponse{
			Status:     "ok",
			URLResults: make([]dispatch.URLResult, len(batch)),
		}
		for i, msg := range batch {
            st := crawler.CrawlPage(msg.AbsURL, msg.Category, rules, redis)
			resp.URLResults[i] = dispatch.URLResult{
				URL:    msg.AbsURL,
				Status: st.Status, 
				RelativeURL: st.RelatedLinks,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Crawler failed to encode response: %v", err)
		}
	})

	log.Println("Starting simulated crawler on :8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Crawler server failed: %v", err)
	}
}