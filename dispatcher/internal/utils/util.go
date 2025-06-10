package util

import (
	"log"
	"strings"
	"net/url"
)

// ExtractDomain extracts the root domain from a URL.
func ExtractDomain(absURL string) (string, error) {
	u, err := url.Parse(absURL)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
}
// FormatKey converts a domain to a sync.Map key (replace . with _).
func FormatKey(domain string) string {
	return strings.ReplaceAll(domain, ".", "_")
}

// LogError logs errors with context.
func LogError(context string, err error) {
	log.Printf("[ERROR] %s: %v\n", context, err)
}