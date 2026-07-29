package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/bits"
	"strconv"
)

type similarityPage struct {
	id    int64
	value uint64
	hash  string
}

// insertNearDuplicateIssues uses four 16-bit SimHash bands. A Hamming distance
// of at most three guarantees at least one equal band, avoiding a full pairwise
// scan. It records one representative match per page, keeping finding volume
// linear even when many pages share a template.
func insertNearDuplicateIssues(ctx context.Context, tx *sql.Tx, crawlID string, threshold int, now string) error {
	rows, err := tx.QueryContext(ctx, `SELECT p.id,p.similarity_hash,p.content_hash
FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id
WHERE cu.crawl_id=? AND length(p.similarity_hash)=16 ORDER BY p.id`, crawlID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var pages []similarityPage
	for rows.Next() {
		var item similarityPage
		var encoded string
		if err := rows.Scan(&item.id, &encoded, &item.hash); err != nil {
			return err
		}
		item.value, err = strconv.ParseUint(encoded, 16, 64)
		if err != nil {
			continue
		}
		pages = append(pages, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	buckets := make(map[uint64]map[string]int, len(pages)*2)
	for index, page := range pages {
		seen := make(map[int]struct{})
		match := -1
		distance := 65
		for band := 0; band < 4; band++ {
			key := uint64(band)<<16 | (page.value >> (band * 16) & 0xffff)
			for _, candidate := range buckets[key] {
				if _, ok := seen[candidate]; ok {
					continue
				}
				seen[candidate] = struct{}{}
				other := pages[candidate]
				candidateDistance := bits.OnesCount64(page.value ^ other.value)
				if candidateDistance <= threshold && candidateDistance < distance {
					match, distance = candidate, candidateDistance
					if distance == 0 {
						break
					}
				}
			}
			if distance == 0 {
				break
			}
		}
		if match >= 0 {
			evidence, _ := json.Marshal(map[string]any{"similarity_hash": fmt.Sprintf("%016x", page.value), "representative_page_id": pages[match].id, "hamming_distance": distance, "threshold": threshold, "extraction_mode": "raw"})
			if _, err := tx.ExecContext(ctx, `INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at) VALUES (?,'AUD-08',1,'page',?,'warning',?,?)`, crawlID, strconv.FormatInt(page.id, 10), string(evidence), now); err != nil {
				return err
			}
		}
		for band := 0; band < 4; band++ {
			key := uint64(band)<<16 | (page.value >> (band * 16) & 0xffff)
			bucket := buckets[key]
			if bucket == nil {
				bucket = make(map[string]int)
				buckets[key] = bucket
			}
			if _, exists := bucket[page.hash]; !exists {
				bucket[page.hash] = index
			}
		}
	}
	return nil
}
