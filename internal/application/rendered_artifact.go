package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/renderer"
)

type managedRenderedArtifactStore struct {
	frontier  *database.Frontier
	directory string
	mu        sync.Mutex
}

func (s *managedRenderedArtifactStore) StoreRenderedArtifacts(ctx context.Context, crawlID contracts.ID, crawlURLID int64, result renderer.Result, configuration contracts.RenderedEvidenceConfiguration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var items []struct {
		kind, mime, extension string
		body                  []byte
	}
	if configuration.RetainDOM {
		redacted, err := redactRenderedDOM([]byte(result.HTML))
		if err != nil {
			return err
		}
		items = append(items, struct {
			kind, mime, extension string
			body                  []byte
		}{"rendered_dom", "text/html; charset=utf-8", "html", redacted})
	}
	if configuration.CaptureScreenshot && len(result.Screenshot) > 0 {
		items = append(items, struct {
			kind, mime, extension string
			body                  []byte
		}{"viewport_screenshot", "image/jpeg", "jpg", result.Screenshot})
	}
	var pageBytes int64
	for _, item := range items {
		pageBytes += int64(len(item.body))
	}
	if pageBytes > configuration.EffectiveMaximumPageBytes() {
		return errors.New("rendered artifact page byte limit reached")
	}
	used, err := s.frontier.PageArtifactBytes(ctx, crawlID)
	if err != nil {
		return err
	}
	if used+pageBytes > configuration.EffectiveMaximumCrawlBytes() {
		return errors.New("rendered artifact crawl byte limit reached")
	}
	for _, item := range items {
		if err := s.storeOne(ctx, crawlID, crawlURLID, item.kind, item.mime, item.extension, result.Viewport, result.EngineVersion, item.body, configuration.EffectiveRetentionDays()); err != nil {
			return err
		}
	}
	return nil
}

func (s *managedRenderedArtifactStore) storeOne(ctx context.Context, crawlID contracts.ID, crawlURLID int64, kind, mime, extension, viewport, engine string, body []byte, retentionDays int) error {
	id, err := contracts.NewID("artifact")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(s.directory, ".render-*")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	committed := false
	defer func() {
		_ = temporary.Close()
		if !committed {
			_ = os.Remove(temporaryName)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return err
	}
	hash := sha256.Sum256(body)
	if _, err := temporary.Write(body); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	name := string(id) + "." + extension
	destination := filepath.Join(s.directory, name)
	if err := os.Rename(temporaryName, destination); err != nil {
		return err
	}
	now := time.Now().UTC()
	record := database.PageArtifactRecord{ArtifactRecord: database.ArtifactRecord{ID: id, CrawlID: crawlID, Format: kind, RelativePath: name, Checksum: hex.EncodeToString(hash[:]), SizeBytes: int64(len(body)), CreatedAt: now.Format(time.RFC3339Nano), ExpiresAt: now.Add(time.Duration(retentionDays) * 24 * time.Hour).Format(time.RFC3339Nano)}, CrawlURLID: crawlURLID, Kind: kind, MIMEType: mime, Viewport: viewport, EngineVersion: engine}
	if err := s.frontier.RecordPageArtifact(ctx, record); err != nil {
		_ = os.Remove(destination)
		return err
	}
	committed = true
	return nil
}

func redactRenderedDOM(body []byte) ([]byte, error) {
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			tag := strings.ToLower(node.Data)
			if tag == "script" {
				node.FirstChild = &html.Node{Type: html.TextNode, Data: "[script omitted]", Parent: node}
				node.LastChild = node.FirstChild
			}
			if tag == "textarea" {
				node.FirstChild = &html.Node{Type: html.TextNode, Data: "[redacted]", Parent: node}
				node.LastChild = node.FirstChild
			}
			for index := range node.Attr {
				name := strings.ToLower(node.Attr[index].Key)
				switch {
				case tag == "input" && name == "value", strings.Contains(name, "password"), strings.Contains(name, "token"), strings.Contains(name, "secret"), strings.Contains(name, "authorization"), strings.Contains(name, "api-key"):
					node.Attr[index].Val = "[redacted]"
				case name == "href" || name == "src" || name == "action" || name == "formaction":
					node.Attr[index].Val = redactArtifactURL(node.Attr[index].Val)
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	var output bytes.Buffer
	if err := html.Render(&output, root); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
func redactArtifactURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil {
		return "[invalid-url]"
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}
