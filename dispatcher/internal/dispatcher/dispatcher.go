package dispatch

import (
	"context"
	"dispatcher/internal/utils"
	"time"
	"sync"
	"github.com/IBM/sarama"
)

type Dispatcher struct {
	Consumer *Consumer
	Batcher  *Batcher
	Sender   *Sender
	Retry    *Retry
}

func NewDispatcher(
	ctx context.Context,
	brokers []string,
	groupID string,
	numInBatch int,
	numWindow int,
	rateError float64,
	topics []string,
) (*Dispatcher, sarama.ConsumerGroup, error) {
	state := &DispatcherState{}
	breaker := NewBreaker(5, 30*time.Second)
	retry := NewRetry(state, 3, 1*time.Hour)

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, nil, err
	}

	consumer := NewConsumer(state)
	batcher := NewBatcher(state, numInBatch)
	sender := NewSender(state, breaker, retry, numWindow, rateError)

	dispatcher := &Dispatcher{
		Consumer: consumer,
		Batcher:  batcher,
		Sender:   sender,
		Retry:    retry,
	}

	return dispatcher, consumerGroup, nil
}

// Run starts the main dispatch loop.
func Run(ctx context.Context, dispatcher *Dispatcher, consumerGroup sarama.ConsumerGroup, topics []string) {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := consumerGroup.Consume(ctx, topics, dispatcher.Consumer); err != nil {
					util.LogError("consumer group", err)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				dispatcher.Retry.ProcessRetries()

				batch := dispatcher.Batcher.CreateBatch()
				if len(batch) > 0 {
					if err := dispatcher.Sender.SendBatch(batch); err != nil {
						util.LogError("send batch", err)
					}
				}
			}
		}
	}()

	wg.Wait()
}

