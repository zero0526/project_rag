package dispatch

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"dispatcher/internal/utils"
)

type Consumer struct {
	State *DispatcherState
}

func NewConsumer(state *DispatcherState) *Consumer {
	return &Consumer{State: state}
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error { return nil }

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var msg MessageModel
		if err := json.Unmarshal(message.Value, &msg); err != nil {
			util.LogError("unmarshal message", err)
			continue
		}

		domain, err := util.ExtractDomain(msg.AbsURL)
		if err != nil {
			util.LogError("extract domain", err)
			continue
		}

		key := util.FormatKey(domain)
		msgs, _ := c.State.DomainMessages.LoadOrStore(key, &[]MessageModel{})
		messageList := msgs.(*[]MessageModel)
		*messageList = append(*messageList, msg)
		c.State.DomainMessages.Store(key, messageList)

		session.MarkMessage(message, "")
	}
	return nil
}