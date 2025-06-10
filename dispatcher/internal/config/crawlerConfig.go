package config 

import (
	"os"
	"gopkg.in/yaml.v3"
	"encoding/json"
)

type CrawlerConfig struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
}
type DataPattern struct {
	Regex   string `json:"regex"`
	Layout  string `json:"layout"`
}
type DateRule struct {
	Pattern  []DataPattern `json:"pattern"`
	Selector string `json:"selector"`
}

type CrawlerRule struct {
	Title        string   `json:"title"`
	Author       string   `json:"author"`
	Content      string   `json:"content"`
	Date         DateRule `json:"date"`
	RelatedLinks string   `json:"related_links"`
	RegexLink    string   `json:"regex_link"`
}

func LoadCrawlerConfig(path string) (*CrawlerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg CrawlerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func LoadCrawlerRules(path string) (map[string]CrawlerRule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rules map[string]CrawlerRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}