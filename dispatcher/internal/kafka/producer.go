package producer

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

type MessageModel struct {
	Category string `json:"category"`
	AbsURL   string `json:"absUrl"`
}
type Producer struct {
	syncProducer sarama.SyncProducer
	brokers      []string
}

func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true

	prod, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		syncProducer: prod,
		brokers:      brokers,
	}, nil
}
func (p *Producer) topicExists(topic string) (bool, error) {
	admin, err := sarama.NewClusterAdmin(p.brokers, sarama.NewConfig())
	if err != nil {
		return false, err
	}
	defer admin.Close()

	topics, err := admin.ListTopics()
	if err != nil {
		return false, err
	}

	_, exists := topics[topic]
	return exists, nil
}

func (p *Producer) Send(topic string, message MessageModel) error {
	// Kiểm tra topic tồn tại
	exists, err := p.topicExists(topic)
	if err != nil {
		log.Printf("Warning: Failed to check topic existence: %v", err)
	} else if !exists {
		log.Printf("Warning: Topic '%s' does not exist!", topic)
	}

	// Mã hóa message
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	kafkaMsg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msgBytes),
	}

	partition, offset, err := p.syncProducer.SendMessage(kafkaMsg)
	if err != nil {
		return err
	}

	log.Printf("Sent message to topic %s [partition: %d, offset: %d]\n", topic, partition, offset)
	return nil
}

func (p *Producer) Close() error {
	return p.syncProducer.Close()
}
