package crawler

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

const maximumLinkAttributeBytes = 8192

// DiscoverLinks performs bounded raw HTML link discovery without executing code.
func DiscoverLinks(documentURL string, body []byte, maximum int) ([]string, error) {
	if maximum < 1 {
		return nil, nil
	}
	base, err := url.Parse(documentURL)
	if err != nil {
		return nil, fmt.Errorf("parse document URL: %w", err)
	}
	activeBase := base
	baseSeen := false
	tokenizer := html.NewTokenizer(bytes.NewReader(body))
	result := make([]string, 0, min(maximum, 256))
	seen := make(map[string]struct{})
	for len(result) < maximum {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			if errors := tokenizer.Err(); errors != nil && errors != io.EOF {
				return result, fmt.Errorf("tokenize HTML: %w", errors)
			}
			break
		}
		if tokenType != html.StartTagToken && tokenType != html.SelfClosingTagToken {
			continue
		}
		token := tokenizer.Token()
		tag := strings.ToLower(token.Data)
		attribute := ""
		switch tag {
		case "a", "area", "link":
			attribute = "href"
		case "iframe", "frame":
			attribute = "src"
		case "base":
			if baseSeen {
				continue
			}
			baseSeen = true
			if value := attributeValue(token, "href"); value != "" {
				if reference, parseErr := url.Parse(value); parseErr == nil {
					candidate := base.ResolveReference(reference)
					if candidate.Scheme == "http" || candidate.Scheme == "https" {
						activeBase = candidate
					}
				}
			}
			continue
		default:
			continue
		}
		raw := strings.TrimSpace(attributeValue(token, attribute))
		if raw == "" || len(raw) > maximumLinkAttributeBytes {
			continue
		}
		reference, err := url.Parse(raw)
		if err != nil {
			continue
		}
		candidate := activeBase.ResolveReference(reference)
		candidate.Fragment = ""
		if candidate.Scheme != "http" && candidate.Scheme != "https" {
			continue
		}
		value := candidate.String()
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func attributeValue(token html.Token, name string) string {
	for _, attribute := range token.Attr {
		if strings.EqualFold(attribute.Key, name) {
			return attribute.Val
		}
	}
	return ""
}
