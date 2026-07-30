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
}

// insertNearDuplicateIssues uses five SimHash bands and indexes each pair of
// bands in a separate pass. A Hamming distance of at most three can affect at
// most three bands, so every qualifying pair shares at least two complete
// bands. Processing one pair at a time keeps memory linear while retaining
// exact candidate generation. It records one deterministic representative per
// page, keeping finding volume linear even when many pages share a template.
func insertNearDuplicateIssues(ctx context.Context, tx *sql.Tx, crawlID string, threshold int, now string) error {
	rows, err := tx.QueryContext(ctx, `SELECT p.id,p.similarity_hash
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
		if err := rows.Scan(&item.id, &encoded); err != nil {
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
	matches := make([]int, len(pages))
	distances := make([]uint8, len(pages))
	for index := range matches {
		matches[index] = -1
	}
	for first := 0; first < 5; first++ {
		for second := first + 1; second < 5; second++ {
			buckets := make(map[uint64][]int, len(pages)/2)
			for index, page := range pages {
				key := similarityBandPairKey(page.value, first, second)
				if matches[index] < 0 {
					for _, candidate := range buckets[key] {
						distance := bits.OnesCount64(page.value ^ pages[candidate].value)
						if distance <= threshold {
							matches[index] = candidate
							distances[index] = uint8(distance)
							break
						}
					}
				}
				buckets[key] = append(buckets[key], index)
			}
		}
	}
	statement, err := tx.PrepareContext(ctx, `INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at) VALUES (?,'AUD-08',1,'page',?,'warning',?,?)`)
	if err != nil {
		return err
	}
	defer statement.Close()
	for index, match := range matches {
		if match < 0 {
			continue
		}
		page := pages[index]
		evidence, _ := json.Marshal(map[string]any{"similarity_hash": fmt.Sprintf("%016x", page.value), "representative_page_id": pages[match].id, "hamming_distance": distances[index], "threshold": threshold, "extraction_mode": "raw"})
		if _, err := statement.ExecContext(ctx, crawlID, strconv.FormatInt(page.id, 10), string(evidence), now); err != nil {
			return err
		}
	}
	return nil
}

func similarityBandPairKey(value uint64, first, second int) uint64 {
	return uint64(first*5+second)<<32 | similarityBand(value, first)<<16 | similarityBand(value, second)
}

func similarityBand(value uint64, band int) uint64 {
	if band == 4 {
		return value >> 52
	}
	return value >> (band * 13) & 0x1fff
}
