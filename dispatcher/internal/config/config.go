package config

import (
	"encoding/json"
	"os"

	"gopkg.in/yaml.v3"
)

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	GroupID string   `yaml:"group_id"`
	TopicPath  string `yaml:"topic_file"`
}

type DispatcherConfig struct {
	NumInBatch int     `yaml:"num_in_batch"`
	NumWindow  int     `yaml:"num_window"`
	RateError  float64 `yaml:"rate_error"`
	URL2Cralwers []string  `yaml:"url2crawlers"`
}

type Config struct {
	Kafka      KafkaConfig      `yaml:"kafka"`
	Dispatcher DispatcherConfig `yaml:"dispatcher"`
}
// LoadConfig đọc cấu hình từ file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// kafka_brokers:
//   - kafka:9092
// crawler_api: http://crawler:8080/crawl
// topics_file_path: /app/configs/topics.json
// retry_attempts: 3
// retry_delay_ms: 1000

// LoadTopics đọc danh sách topic từ file topics.json
func LoadTopics(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var topics []string
	if err := json.Unmarshal(data, &topics); err != nil {
		return nil, err
	}
	return topics, nil
}
