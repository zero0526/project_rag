package dispatch

import (
	"sync"
	"time"
	"log"
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
	HalfOpenAttempts  int  
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

	if now.Before(cs.BlockedUntil) {
		return false
	}

	if cs.BlockedUntil != (time.Time{}) && !cs.InHalfOpen {
		cs.BlockedUntil = time.Time{}
		cs.InHalfOpen = true
		cs.HalfOpenAttempts = 0
		log.Printf("[Breaker] Domain %s chuyển sang trạng thái Half-Open", domain)
	}

	// Nếu đang Half-Open, giới hạn 3 request
	if cs.InHalfOpen {
		if cs.HalfOpenAttempts >= 3 {
			return false
		}
		cs.HalfOpenAttempts++
		return true
	}

	return true
}


// chan theo so luong loi
func (b *Breaker) RecordFailure(domain string) {
	state, _ := b.States.LoadOrStore(domain, &CircuitState{})
	cs := state.(*CircuitState)
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	if cs.InHalfOpen {
		if cs.HalfOpenAttempts >= 3 {
			cs.InHalfOpen = false
			cs.HalfOpenAttempts = 0
			cs.ConsecutiveOpen++
			cs.BlockedUntil = time.Now().Add(b.calcBackoff(cs.ConsecutiveOpen))
			cs.Failures = 0
			log.Printf("[Breaker] Domain %s thất bại 3 lần trong Half-Open => Open lại", domain)
		}
		return
	}

	cs.Failures++
	if cs.Failures >= b.FailureThreshold {
		cs.ConsecutiveOpen++
		cs.BlockedUntil = time.Now().Add(b.calcBackoff(cs.ConsecutiveOpen))
		cs.Failures = 0
	}
}

//  ghi nhân success để rest fail
func (b *Breaker) RecordSuccess(domain string) {
	state, _ := b.States.LoadOrStore(domain, &CircuitState{})
	cs := state.(*CircuitState)
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	cs.InHalfOpen = false
	cs.ConsecutiveOpen = 0
	cs.HalfOpenAttempts = 0
	cs.Failures = 0
	cs.BlockedUntil = time.Time{}
}
// block thôi 
func (b *Breaker) BlockDomain(domain string) {
	state, _ := b.States.LoadOrStore(domain, &CircuitState{})
	cs := state.(*CircuitState)
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	duration := b.calcBackoff(cs.ConsecutiveOpen)
	cs.BlockedUntil = time.Now().Add(duration)
	cs.Failures = 0
	cs.InHalfOpen = false
	cs.ConsecutiveOpen++
}

func (b *Breaker) calcBackoff(attempts int) time.Duration {
	const maxShift = 16
	if attempts > maxShift {
		attempts = maxShift
	}
	return b.BackoffDuration * time.Duration(1<<attempts)
}
func (b *Breaker) GetState(domain string) string {
	state, ok := b.States.Load(domain)
	if !ok {
		return "Closed"
	}
	cs := state.(*CircuitState)
	cs.Mu.Lock()
	defer cs.Mu.Unlock()

	now := time.Now()
	switch {
	case now.Before(cs.BlockedUntil):
		return "Open"
	case cs.InHalfOpen:
		return "Half-Open"
	default:
		return "Closed"
	}
}

