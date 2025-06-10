package crawler

import (
	"log"
	"regexp"
	"strings"
    "time"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"

    "dispatcher/internal/config"
    "dispatcher/internal/utils"
)

type RawDocument struct {
	URL      string `json:"url"`
	Category string `json:"category"`
	RawHTML  string `json:"raw_html"`
}

type ExtractedDocument struct {
	URL         string   `json:"url"`
	Category    string   `json:"category"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Content     string   `json:"content"`
	Date        string   `json:"date"`
	RelatedURLs []string `json:"related_links"`
}
type CrawlResult struct {
	Status       string   `json:"status"`
	RelatedLinks []string `json:"relatedLinks,omitempty"`
}

var removeTags = []string{"script", "style", "noscript", "meta", "link", "iframe", "svg", "img"}

func CrawlPage(url string, category string, rules map[string]config.CrawlerRule, redis *RedisClient) CrawlResult {
	domainKey, err := util.ExtractDomain(url)
	if err != nil {
		log.Println("Fail to extract domain")
		return CrawlResult{Status: "failed"}
	}

	keyDomain := util.FormatKey(domainKey)
	fmt.Println(keyDomain,"keydomainavassdjkvbaksjvbsk")
	rule, hasRule := rules[keyDomain]
	c := colly.NewCollector()

	var crawlStatus = CrawlResult{Status: "failed"}

	c.OnResponse(func(r *colly.Response) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(r.Body)))
		if err != nil {
			log.Printf("Failed to parse HTML for %s: %v", url, err)
			crawlStatus = CrawlResult{Status: "failed"}
			return
		}

		// Loại bỏ các thẻ không cần thiết cho dù có rule hay không
		for _, tag := range removeTags {
			doc.Find(tag).Each(func(i int, s *goquery.Selection) {
				s.Remove()
			})
		}

		cleanHTML, _ := doc.Html()
		// fmt.Println(cleanHTML)
		// Kiểm tra nếu nội dung quá ngắn
		if len(strings.TrimSpace(cleanHTML)) < 100 {
			log.Printf("Blocked due to too short content: %s", url)
			crawlStatus = CrawlResult{Status: "blocked"}
			return
		}

		// Nếu có rule, trích xuất structured
		if hasRule {
			extracted := extractWithRule(url, category, doc, rule)
			err := redis.PushToList("extracted_list", extracted)
			if err != nil {
				log.Printf("Redis push error: %v", err)
				crawlStatus = CrawlResult{Status: "failed"}
			} else {
				crawlStatus = CrawlResult{Status: "success"}
				crawlStatus.RelatedLinks = extracted.RelatedURLs
			}
		} else {
			raw := RawDocument{
				URL:      url,
				Category: category,
				RawHTML:  cleanHTML,
			}
			err := redis.PushToList("raw_html_list", raw)
			if err != nil {
				log.Printf("Redis push error: %v", err)
				crawlStatus = CrawlResult{Status: "failed"}
			} else {
				crawlStatus = CrawlResult{Status: "success"}
			}
		}
	})

	err = c.Visit(url)
	if err != nil {
		log.Printf("Failed to visit %s: %v", url, err)
		return CrawlResult{Status: "failed"}
	}

	return crawlStatus
}


func extractWithRule(url, category string, doc *goquery.Document, rule config.CrawlerRule) ExtractedDocument {
	title := strings.TrimSpace(doc.Find(rule.Title).Text())
	var parts []string
	doc.Find(rule.Content).Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			parts = append(parts, text)
		}
	})
	content := strings.Join(parts, "\n\n")
	parts = parts[:0]
	doc.Find(rule.Author).Each(func(i int, s *goquery.Selection){
		text := strings.TrimSpace(s.Text())
		if text != ""{
			parts = append(parts, text)
		}
	})
	author := strings.Join(parts, ",::,")
	date := strings.TrimSpace(extractDate(doc,rule.Date, url))

	relatedLinks := []string{}
	linkRegex := regexp.MustCompile(rule.RegexLink)

	doc.Find(rule.RelatedLinks).Each(func(_ int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			if linkRegex.MatchString(href) {
				relatedLinks = append(relatedLinks, href)
			}
		}
	})

	return ExtractedDocument{
		URL:         url,
		Category:    category,
		Title:       title,
		Author:      author,
		Content:     content,
		Date:        date,
		RelatedURLs: relatedLinks,
	}
}

func extractDate(doc *goquery.Document, dateRule config.DateRule, url string) string {
    rawText := strings.TrimSpace(doc.Find(dateRule.Selector).Text())
    if rawText == "" {
        log.Println("Date selector returned empty text")
		log.Println(url)
        return ""
    }

    for _, pattern := range dateRule.Pattern {
        re, err := regexp.Compile(pattern.Regex)
        if err != nil {
            log.Printf("Invalid regex in date rule: %v", err)
            continue
        }

        match := re.FindString(rawText)
        if match == "" {
            continue
        }

        parsedTime, err := time.Parse(pattern.Layout, match)
        if err != nil {
            log.Printf("Failed to parse date '%s' with layout '%s': %v", match, pattern.Layout, err)
            return match // Trả về chuỗi match nếu không parse được
        }

        return parsedTime.Format(time.RFC3339)
    }

    log.Printf("No date matched any pattern in: %s", rawText)
    return ""
}
