package dispatch

import (
	"sync"
	"time"
)

type Breaker struct {
	States    sync.Map // key: domain, value: *CircuitState
	FailureThreshold int
	BackoffDuration  time.Duration
}

type CircuitState struct {
	Failures       int
	BlockedUntil   time.Time
	ConsecutiveOpen int // số lần liên tiếp bị mở lại sau Half-Open
	InHalfOpen     bool
	Mu             sync.Mutex
}

func NewBreaker(failureThreshold int, backoffDuration time.Duration) *Breaker {
	return &Breaker{
		FailureThreshold: failureThreshold,
		BackoffDuration:  backoffDuration,
	}
}

// Allow checks if requests to a domain are allowed.
func (b *Breaker) Allow(domain string) bool {
	state, _ := b.States.LoadOrStore(domain, &CircuitState{})
	cs := state.(*CircuitState)
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	now := time.Now()
	if now.After(cs.BlockedUntil) {
		if cs.InHalfOpen {
			// Đã hết thời gian Half-Open → vẫn trong Half-Open, cho phép request thử
			return true
		} else if cs.BlockedUntil != (time.Time{}) {
			// Mới hết thời gian Open → chuyển sang Half-Open
			cs.InHalfOpen = true
			return true
		}
	}
	return now.After(cs.BlockedUntil)
}

// RecordFailure increments failure count and blocks if threshold is reached.
func (b *Breaker) RecordFailure(domain string) {
	state, _ := b.States.LoadOrStore(domain, &CircuitState{})
	cs := state.(*CircuitState)
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	cs.Failures++
	if cs.Failures >= b.FailureThreshold {
		cs.BlockedUntil = time.Now().Add(b.BackoffDuration)
		cs.Failures = 0
	}
}
func (b *Breaker) RecordSuccess(domain string) {
	state, _ := b.States.LoadOrStore(domain, &CircuitState{})
	cs := state.(*CircuitState)
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	if cs.InHalfOpen {
		// Thành công trong Half-Open → reset về Closed
		cs.InHalfOpen = false
		cs.ConsecutiveOpen = 0
		cs.Failures = 0
		cs.BlockedUntil = time.Time{}
	}
}

// BlockDomain explicitly blocks a domain.
func (b *Breaker) BlockDomain(domain string) {
	state, _ := b.States.LoadOrStore(domain, &CircuitState{})
	cs := state.(*CircuitState)
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	cs.BlockedUntil = time.Now().Add(b.BackoffDuration)
}