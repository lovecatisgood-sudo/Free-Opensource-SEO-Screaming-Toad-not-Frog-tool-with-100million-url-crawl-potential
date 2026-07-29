package crawler

import (
	"net/url"
	"strings"

	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

const (
	maximumQueryParameters = 50
	maximumRepeatedSegment = 8
	maximumPathSegments    = 200
)

func DetectTrap(target fetchpolicy.NormalizedURL) string {
	segments := strings.FieldsFunc(target.URL.EscapedPath(), func(r rune) bool { return r == '/' })
	if len(segments) > maximumPathSegments {
		return "excessive_path_depth"
	}
	runs := 1
	for index := 1; index < len(segments); index++ {
		if segments[index] == segments[index-1] && segments[index] != "" {
			runs++
			if runs > maximumRepeatedSegment {
				return "repeated_path_segment"
			}
		} else {
			runs = 1
		}
	}
	query, err := url.ParseQuery(target.URL.RawQuery)
	if err != nil {
		return "malformed_query"
	}
	count := 0
	for _, values := range query {
		count += len(values)
		if count > maximumQueryParameters {
			return "excessive_query_parameters"
		}
	}
	return ""
}
