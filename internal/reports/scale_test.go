package reports

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
)

type synthetic100KSource struct{}

func (synthetic100KSource) ListPages(_ context.Context, _ contracts.ID, request contracts.PageRequest) (contracts.Page[database.PageRecord], error) {
	offset := 0
	if request.Cursor != "" {
		offset, _ = strconv.Atoi(request.Cursor)
	}
	limit := request.BoundedLimit()
	end := min(offset+limit, 100_000)
	items := make([]database.PageRecord, 0, end-offset)
	for index := offset; index < end; index++ {
		items = append(items, database.PageRecord{ID: int64(index + 1), URL: fmt.Sprintf("https://example.com/page/%06d", index), Title: "Synthetic page", StatusCode: 200, Depth: 2})
	}
	next := ""
	if end < 100_000 {
		next = strconv.Itoa(end)
	}
	return contracts.Page[database.PageRecord]{Items: items, NextCursor: next}, nil
}
func (synthetic100KSource) ListIssues(_ context.Context, _ contracts.ID, request contracts.PageRequest) (contracts.Page[database.IssueRecord], error) {
	return contracts.Page[database.IssueRecord]{Items: []database.IssueRecord{}}, nil
}

func TestStreaming100KPageExportsStayMemoryBounded(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 100k export gate in short mode")
	}
	ctx := context.Background()
	source := synthetic100KSource{}
	started := time.Now()
	if err := PagesCSV(ctx, source, "crawl_scale", io.Discard); err != nil {
		t.Fatal(err)
	}
	if err := PagesNDJSON(ctx, source, "crawl_scale", io.Discard); err != nil {
		t.Fatal(err)
	}
	if err := WorkbookXLSX(ctx, source, "crawl_scale", io.Discard); err != nil {
		t.Fatal(err)
	}
	runtime.GC()
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	if memory.Alloc > 128<<20 {
		t.Fatalf("streaming exports retained %dMiB", memory.Alloc/(1024*1024))
	}
	t.Logf("exported=100000 formats=3 elapsed=%s heap_alloc=%dMiB", time.Since(started).Round(time.Millisecond), memory.Alloc/(1024*1024))
}
