package crawler

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

type SitemapKind string

const (
	SitemapURLSet SitemapKind = "urlset"
	SitemapIndex  SitemapKind = "sitemapindex"
)

type SitemapDocument struct {
	Kind      SitemapKind
	Locations []string
}

type SitemapLimits struct {
	MaximumBytes   int64
	MaximumEntries int
}

func DefaultSitemapLimits() SitemapLimits {
	return SitemapLimits{MaximumBytes: 50 << 20, MaximumEntries: 50_000}
}

func ParseSitemap(reader io.Reader, limits SitemapLimits) (SitemapDocument, error) {
	if limits.MaximumBytes < 1 || limits.MaximumEntries < 1 || limits.MaximumEntries > 1_000_000 {
		return SitemapDocument{}, errors.New("invalid sitemap limits")
	}
	limited := &io.LimitedReader{R: reader, N: limits.MaximumBytes + 1}
	decoder := xml.NewDecoder(limited)
	decoder.Strict = true
	document := SitemapDocument{Locations: make([]string, 0, min(limits.MaximumEntries, 1024))}
	insideLoc := false
	var locationText strings.Builder
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return SitemapDocument{}, fmt.Errorf("decode sitemap: %w", err)
		}
		switch value := token.(type) {
		case xml.StartElement:
			switch strings.ToLower(value.Name.Local) {
			case "urlset":
				if document.Kind != "" && document.Kind != SitemapURLSet {
					return SitemapDocument{}, errors.New("sitemap has conflicting root types")
				}
				document.Kind = SitemapURLSet
			case "sitemapindex":
				if document.Kind != "" && document.Kind != SitemapIndex {
					return SitemapDocument{}, errors.New("sitemap has conflicting root types")
				}
				document.Kind = SitemapIndex
			case "loc":
				insideLoc = true
				locationText.Reset()
			}
		case xml.CharData:
			if insideLoc {
				locationText.Write(value)
			}
		case xml.EndElement:
			if strings.EqualFold(value.Name.Local, "loc") {
				location := strings.TrimSpace(locationText.String())
				if location != "" {
					if len(document.Locations) >= limits.MaximumEntries {
						return SitemapDocument{}, errors.New("sitemap entry limit reached")
					}
					document.Locations = append(document.Locations, location)
				}
				insideLoc = false
			}
		}
	}
	if limited.N <= 0 {
		return SitemapDocument{}, errors.New("sitemap byte limit reached")
	}
	if document.Kind == "" {
		return SitemapDocument{}, errors.New("document is not a sitemap URL set or index")
	}
	return document, nil
}
