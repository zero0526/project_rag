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
func (r *Retry) QueueForRetry(messages []MessageModel) {
	for _, msg := range messages {
		domain, err := util.ExtractDomain(msg.AbsURL)
		if err != nil {
			continue
		}
		key := util.FormatKey(domain)

		retries, _ := r.State.RetryMessages.LoadOrStore(key, &sync.Map{})
		retryMap := retries.(*sync.Map)

		val, exists := retryMap.Load(msg.AbsURL)
		var meta RetryMeta
		if !exists {
			meta = RetryMeta{
				Count:      1,
				FirstRetry: time.Now(),
			}
		} else {
			meta = val.(RetryMeta)
			meta.Count++
		}

		if meta.Count > r.MaxRetries {
			fmt.Printf("[INFO] Max retries reached for URL: %s\n", msg.AbsURL)
			retryMap.Delete(msg.AbsURL)
			continue
		}

		retryMap.Store(msg.AbsURL, meta)
		r.State.RetryMessages.Store(key, retryMap)
	}
}

// ProcessRetries moves retry messages back to the main queue.

func (r *Retry) ProcessRetries() {
	r.State.RetryMessages.Range(func(key, value interface{}) bool {
		retryMap := value.(*sync.Map)
		domainMessages, _ := r.State.DomainMessages.LoadOrStore(key, &[]MessageModel{})
		messageList := domainMessages.(*[]MessageModel)

		retryMap.Range(func(url, meta  interface{}) bool {
			retryMeta := meta.(RetryMeta)
			if time.Since(retryMeta.FirstRetry) >= r.TTL {
				*messageList = append(*messageList, MessageModel{AbsURL: url.(string)})
				retryMap.Delete(url)
			}
			return true
		})

		r.State.DomainMessages.Store(key, messageList)
		return true
	})
}