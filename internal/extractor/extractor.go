package extractor

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

type Heading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
}
type Link struct {
	RawURL  string `json:"raw_url"`
	URL     string `json:"url"`
	Text    string `json:"text"`
	Rel     string `json:"rel"`
	Element string `json:"element"`
}
type Image struct {
	RawURL     string `json:"raw_url"`
	URL        string `json:"url"`
	Alt        string `json:"alt"`
	AltPresent bool   `json:"alt_present"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
}
type Hreflang struct {
	Language string `json:"language"`
	RawURL   string `json:"raw_url"`
	URL      string `json:"url"`
}
type StructuredData struct {
	Format string   `json:"format"`
	Types  []string `json:"types"`
	Valid  bool     `json:"valid"`
	Error  string   `json:"error,omitempty"`
}

type Page struct {
	URL, Title, MetaDescription, MetaRobots, XRobotsTag    string
	Canonicals                                             []string
	Headings                                               []Heading
	Links                                                  []Link
	Images                                                 []Image
	Hreflangs                                              []Hreflang
	StructuredData                                         []StructuredData
	Social                                                 map[string]string
	Language, Viewport, PaginationNext, PaginationPrevious string
	VisibleText                                            string
	WordCount                                              int
	HTMLHash, ContentHash                                  string
}

func Extract(documentURL string, headers http.Header, body []byte) (Page, error) {
	base, err := url.Parse(documentURL)
	if err != nil {
		return Page{}, fmt.Errorf("parse document URL: %w", err)
	}
	document, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return Page{}, fmt.Errorf("parse HTML: %w", err)
	}
	page := Page{URL: documentURL, XRobotsTag: strings.Join(headers.Values("X-Robots-Tag"), ", "), Social: make(map[string]string)}
	if htmlNode := findElement(document, "html"); htmlNode != nil {
		page.Language = attribute(htmlNode, "lang")
	}
	if baseNode := findElement(document, "base"); baseNode != nil {
		if parsed, parseErr := url.Parse(attribute(baseNode, "href")); parseErr == nil {
			candidate := base.ResolveReference(parsed)
			if candidate.Scheme == "http" || candidate.Scheme == "https" {
				base = candidate
			}
		}
	}
	var visible []string
	walk(document, false, func(node *html.Node, hidden bool) {
		if node.Type == html.TextNode && !hidden {
			if value := normalizeText(node.Data); value != "" {
				visible = append(visible, value)
			}
			return
		}
		if node.Type != html.ElementNode {
			return
		}
		switch strings.ToLower(node.Data) {
		case "title":
			if page.Title == "" {
				page.Title = nodeText(node)
			}
		case "meta":
			name, property, content := strings.ToLower(attribute(node, "name")), strings.ToLower(attribute(node, "property")), attribute(node, "content")
			switch name {
			case "description":
				if page.MetaDescription == "" {
					page.MetaDescription = content
				}
			case "robots":
				if page.MetaRobots == "" {
					page.MetaRobots = content
				}
			case "viewport":
				if page.Viewport == "" {
					page.Viewport = content
				}
			}
			if strings.HasPrefix(property, "og:") || strings.HasPrefix(name, "twitter:") {
				key := property
				if key == "" {
					key = name
				}
				if _, exists := page.Social[key]; !exists {
					page.Social[key] = content
				}
			}
		case "link":
			rel, raw := strings.ToLower(attribute(node, "rel")), attribute(node, "href")
			resolved := resolve(base, raw)
			for _, token := range strings.Fields(rel) {
				switch token {
				case "canonical":
					if resolved != "" {
						page.Canonicals = append(page.Canonicals, resolved)
					}
				case "next":
					page.PaginationNext = resolved
				case "prev", "previous":
					page.PaginationPrevious = resolved
				case "alternate":
					if language := attribute(node, "hreflang"); language != "" && resolved != "" {
						page.Hreflangs = append(page.Hreflangs, Hreflang{language, raw, resolved})
					}
				}
			}
		case "a", "area":
			raw := attribute(node, "href")
			if resolved := resolve(base, raw); resolved != "" {
				page.Links = append(page.Links, Link{raw, resolved, nodeText(node), attribute(node, "rel"), node.Data})
			}
		case "iframe", "frame":
			raw := attribute(node, "src")
			if resolved := resolve(base, raw); resolved != "" {
				page.Links = append(page.Links, Link{RawURL: raw, URL: resolved, Element: node.Data})
			}
		case "img":
			raw := attribute(node, "src")
			if resolved := resolve(base, raw); resolved != "" {
				_, present := attributeWithPresence(node, "alt")
				page.Images = append(page.Images, Image{raw, resolved, attribute(node, "alt"), present, positiveInt(attribute(node, "width")), positiveInt(attribute(node, "height"))})
			}
		case "h1", "h2", "h3", "h4", "h5", "h6":
			level, _ := strconv.Atoi(node.Data[1:])
			page.Headings = append(page.Headings, Heading{level, nodeText(node)})
		case "script":
			if strings.EqualFold(attribute(node, "type"), "application/ld+json") {
				page.StructuredData = append(page.StructuredData, parseJSONLD(nodeText(node)))
			}
		}
		if itemType := attribute(node, "itemtype"); itemType != "" {
			page.StructuredData = append(page.StructuredData, StructuredData{Format: "microdata", Types: strings.Fields(itemType), Valid: true})
		}
		if rdfType := attribute(node, "typeof"); rdfType != "" {
			page.StructuredData = append(page.StructuredData, StructuredData{Format: "rdfa", Types: strings.Fields(rdfType), Valid: true})
		}
	})
	page.VisibleText = normalizeText(strings.Join(visible, " "))
	page.WordCount = len(strings.Fields(page.VisibleText))
	htmlHash, contentHash := sha256.Sum256(body), sha256.Sum256([]byte(page.VisibleText))
	page.HTMLHash = hex.EncodeToString(htmlHash[:])
	page.ContentHash = hex.EncodeToString(contentHash[:])
	return page, nil
}

func walk(node *html.Node, hidden bool, visit func(*html.Node, bool)) {
	nowHidden := hidden
	if node.Type == html.ElementNode {
		switch strings.ToLower(node.Data) {
		case "head", "script", "style", "template", "noscript":
			nowHidden = true
		}
		if strings.EqualFold(attribute(node, "aria-hidden"), "true") || hasHiddenStyle(attribute(node, "style")) || hasAttribute(node, "hidden") {
			nowHidden = true
		}
	}
	visit(node, nowHidden)
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walk(child, nowHidden, visit)
	}
}
func nodeText(node *html.Node) string {
	var values []string
	var collect func(*html.Node)
	collect = func(n *html.Node) {
		if n.Type == html.TextNode {
			values = append(values, n.Data)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			collect(child)
		}
	}
	collect(node)
	return normalizeText(strings.Join(values, " "))
}
func normalizeText(value string) string {
	return strings.Join(strings.FieldsFunc(value, unicode.IsSpace), " ")
}
func resolve(base *url.URL, raw string) string {
	if raw == "" || len(raw) > 8192 {
		return ""
	}
	reference, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	result := base.ResolveReference(reference)
	result.Fragment = ""
	if result.Scheme != "http" && result.Scheme != "https" {
		return ""
	}
	return result.String()
}
func attribute(node *html.Node, name string) string {
	value, _ := attributeWithPresence(node, name)
	return value
}
func attributeWithPresence(node *html.Node, name string) (string, bool) {
	for _, item := range node.Attr {
		if strings.EqualFold(item.Key, name) {
			return strings.TrimSpace(item.Val), true
		}
	}
	return "", false
}
func hasAttribute(node *html.Node, name string) bool {
	_, found := attributeWithPresence(node, name)
	return found
}
func hasHiddenStyle(style string) bool {
	style = strings.ToLower(strings.ReplaceAll(style, " ", ""))
	return strings.Contains(style, "display:none") || strings.Contains(style, "visibility:hidden")
}
func positiveInt(value string) int {
	result, _ := strconv.Atoi(value)
	if result < 0 {
		return 0
	}
	return result
}
func findElement(node *html.Node, name string) *html.Node {
	if node.Type == html.ElementNode && strings.EqualFold(node.Data, name) {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findElement(child, name); found != nil {
			return found
		}
	}
	return nil
}
func parseJSONLD(value string) StructuredData {
	result := StructuredData{Format: "json-ld", Valid: true}
	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		result.Valid, result.Error = false, err.Error()
		return result
	}
	types := make(map[string]struct{})
	collectJSONLDTypes(decoded, types)
	for value := range types {
		result.Types = append(result.Types, value)
	}
	slices.Sort(result.Types)
	return result
}
func collectJSONLDTypes(value any, types map[string]struct{}) {
	switch typed := value.(type) {
	case map[string]any:
		if raw, exists := typed["@type"]; exists {
			switch item := raw.(type) {
			case string:
				types[item] = struct{}{}
			case []any:
				for _, child := range item {
					if text, ok := child.(string); ok {
						types[text] = struct{}{}
					}
				}
			}
		}
		for _, child := range typed {
			collectJSONLDTypes(child, types)
		}
	case []any:
		for _, child := range typed {
			collectJSONLDTypes(child, types)
		}
	}
}
