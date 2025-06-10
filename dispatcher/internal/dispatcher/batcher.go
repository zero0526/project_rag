package dispatch

type Batcher struct {
	State *DispatcherState
	numInBatch  int 
}

func NewBatcher(state *DispatcherState, numInBatch  int) *Batcher {
	return &Batcher{State: state, numInBatch: numInBatch}
}

// CreateBatch builds a batch with interleaved domains (max 50 URLs).
func (b *Batcher) CreateBatch() []MessageModel {
	var batch []MessageModel
	domainMessages := make(map[string][]MessageModel)
	var keys []string

	// Collect available messages from sync.Map.
	b.State.DomainMessages.Range(func(key, value interface{}) bool {
		k := key.(string)
		msgs := value.(*[]MessageModel)
		if len(*msgs) > 0 {
			domainMessages[k] = *msgs
			keys = append(keys, k)
		}
		return true
	})

	// Interleave messages from different domains.
	for len(batch) < b.numInBatch && len(keys) > 0 {
		for i := 0; i < len(keys); i++ {
			key := keys[i]
			if len(domainMessages[key]) == 0 {
				keys = append(keys[:i], keys[i+1:]...)
				i--
				continue
			}
			batch = append(batch, domainMessages[key][0])
			domainMessages[key] = domainMessages[key][1:]
			if len(domainMessages[key]) == 0 {
				keys = append(keys[:i], keys[i+1:]...)
				i--
			}
		}
	}

	// Update sync.Map with remaining messages.
	for key, msgs := range domainMessages {
		b.State.DomainMessages.Store(key, &msgs)
	}

	return batch
}