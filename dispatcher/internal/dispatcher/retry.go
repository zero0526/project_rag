package dispatch

import (
	"sync"
	"time"
    "dispatcher/internal/utils"
    "fmt"
)

type Retry struct {
	State      *DispatcherState
	MaxRetries int
	TTL        time.Duration
}

func NewRetry(state *DispatcherState, maxRetries int, ttl time.Duration) *Retry {
	return &Retry{
		State:      state,
		MaxRetries: maxRetries,
		TTL:        ttl,
	}
}

// QueueForRetry adds messages to the retry queue.
// pha count retry
func (r *Retry) QueueForRetry(messages []MessageModel) {
	for _, msg := range messages {
		domain, err := util.ExtractDomain(msg.AbsURL)
		if err != nil {
			continue
		}
		key := util.FormatKey(domain)

		// Load or initialize retry map for the domain.
		retries, _ := r.State.RetryMessages.LoadOrStore(key, &sync.Map{})
		retryMap := retries.(*sync.Map)

		// Increment retry count.
		count, _ := retryMap.LoadOrStore(msg.AbsURL, 0)
		retryCount := count.(int) + 1
		//  delete if beyond retryCount
		if retryCount > r.MaxRetries {
			fmt.Printf("[INFO] Max retries reached for URL: %s\n", msg.AbsURL)
			retryMap.Delete(msg.AbsURL)
			continue
		}

		retryMap.Store(msg.AbsURL, retryCount)
		r.State.RetryMessages.Store(key, retryMap)
	}
}

// ProcessRetries moves retry messages back to the main queue.

func (r *Retry) ProcessRetries() {
	r.State.RetryMessages.Range(func(key, value interface{}) bool {
		retryMap := value.(*sync.Map)
		domainMessages, _ := r.State.DomainMessages.LoadOrStore(key, &[]MessageModel{})
		messageList := domainMessages.(*[]MessageModel)

		retryMap.Range(func(url, count interface{}) bool {
			// Check TTL (simplified as retry count check).
			if time.Now().After(time.Now().Add(-r.TTL)) {
				*messageList = append(*messageList, MessageModel{AbsURL: url.(string)})
				retryMap.Delete(url)
			}
			return true
		})

		r.State.DomainMessages.Store(key, messageList)
		return true
	})
}