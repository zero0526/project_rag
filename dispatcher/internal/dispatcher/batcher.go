package dispatch

type Batcher struct {
	State *DispatcherState
	NumInBatch  int 
	breaker *Breaker
}

func NewBatcher(state *DispatcherState, numInBatch  int, breaker *Breaker) *Batcher {
	return &Batcher{State: state, NumInBatch: numInBatch, breaker: breaker}
}

func (b *Batcher) CreateBatch() []MessageModel {
	var batch []MessageModel
	var keys []string
	domainMessages := make(map[string]*[]MessageModel)

	b.State.DomainMessages.Range(func(key, value interface{}) bool {
		k := key.(string)
		msgs := value.(*[]MessageModel)
		if len(*msgs) > 0 && b.breaker.Allow(k) {
			domainMessages[k] = msgs
			keys = append(keys, k)
		}
		return true
	})

	for len(batch) < b.NumInBatch && len(keys) > 0 {
		for i := 0; i < len(keys); i++ {
			key := keys[i]
			msgs := domainMessages[key]

			if len(*msgs) == 0 {
				keys = append(keys[:i], keys[i+1:]...)
				i--
				continue
			}

			batch = append(batch, (*msgs)[0])
			*msgs = (*msgs)[1:]

			if len(*msgs) == 0 {
				keys = append(keys[:i], keys[i+1:]...)
				i--
			}
		}
	}

	return batch
}
