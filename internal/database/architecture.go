package database

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

type ArchitectureNode struct {
	URLID      int64   `json:"url_id"`
	URL        string  `json:"url"`
	Depth      int     `json:"depth"`
	StatusCode int     `json:"status_code"`
	Inlinks    int     `json:"inlinks"`
	Outlinks   int     `json:"outlinks"`
	Segment    string  `json:"segment"`
	Score      float64 `json:"score"`
	Orphan     bool    `json:"orphan"`
}
type ArchitectureEdge struct {
	SourceURLID int64  `json:"source_url_id"`
	TargetURLID int64  `json:"target_url_id"`
	Rel         string `json:"rel,omitempty"`
}
type ArchitectureGraph struct {
	CrawlID   contracts.ID       `json:"crawl_id"`
	Nodes     []ArchitectureNode `json:"nodes"`
	Edges     []ArchitectureEdge `json:"edges"`
	Truncated bool               `json:"truncated"`
	NodeLimit int                `json:"node_limit"`
	EdgeLimit int                `json:"edge_limit"`
}

func (f *Frontier) Architecture(ctx context.Context, crawlID contracts.ID, nodeLimit, edgeLimit int) (ArchitectureGraph, error) {
	if nodeLimit < 1 || nodeLimit > 5000 || edgeLimit < 1 || edgeLimit > 20000 {
		return ArchitectureGraph{}, errors.New("architecture limits are outside supported bounds")
	}
	rows, err := f.db.QueryContext(ctx, `SELECT u.id,u.request_key,cu.depth,COALESCE((SELECT status_code FROM fetch_attempt WHERE crawl_url_id=cu.id ORDER BY attempt DESC LIMIT 1),0),
		(SELECT count(*) FROM link WHERE crawl_id=cu.crawl_id AND target_url_id=u.id AND link_kind='internal'),
		(SELECT count(*) FROM link WHERE crawl_id=cu.crawl_id AND source_url_id=u.id AND link_kind='internal')
		FROM crawl_url cu JOIN url u ON u.id=cu.url_id WHERE cu.crawl_id=? ORDER BY 5 DESC,6 DESC,cu.depth,u.id LIMIT ?`, crawlID, nodeLimit+1)
	if err != nil {
		return ArchitectureGraph{}, err
	}
	defer rows.Close()
	graph := ArchitectureGraph{CrawlID: crawlID, Nodes: []ArchitectureNode{}, Edges: []ArchitectureEdge{}, NodeLimit: nodeLimit, EdgeLimit: edgeLimit}
	ids := map[int64]bool{}
	for rows.Next() {
		var node ArchitectureNode
		if err := rows.Scan(&node.URLID, &node.URL, &node.Depth, &node.StatusCode, &node.Inlinks, &node.Outlinks); err != nil {
			return ArchitectureGraph{}, err
		}
		if len(graph.Nodes) == nodeLimit {
			graph.Truncated = true
			continue
		}
		node.Segment = architectureSegment(node.URL)
		node.Score = float64(node.Inlinks*2+node.Outlinks) / (float64(node.Depth) + 1)
		node.Orphan = node.Depth > 0 && node.Inlinks == 0
		graph.Nodes = append(graph.Nodes, node)
		ids[node.URLID] = true
	}
	if err := rows.Err(); err != nil {
		return ArchitectureGraph{}, err
	}
	edges, err := f.db.QueryContext(ctx, `SELECT source_url_id,target_url_id,rel FROM link WHERE crawl_id=? AND link_kind='internal' ORDER BY id LIMIT ?`, crawlID, edgeLimit+1)
	if err != nil {
		return ArchitectureGraph{}, err
	}
	defer edges.Close()
	for edges.Next() {
		var edge ArchitectureEdge
		if err := edges.Scan(&edge.SourceURLID, &edge.TargetURLID, &edge.Rel); err != nil {
			return ArchitectureGraph{}, err
		}
		if !ids[edge.SourceURLID] || !ids[edge.TargetURLID] {
			continue
		}
		if len(graph.Edges) == edgeLimit {
			graph.Truncated = true
			continue
		}
		graph.Edges = append(graph.Edges, edge)
	}
	if err := edges.Err(); err != nil {
		return ArchitectureGraph{}, err
	}
	return graph, nil
}

func architectureSegment(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "unknown"
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "/"
	}
	return "/" + parts[0] + "/"
}
