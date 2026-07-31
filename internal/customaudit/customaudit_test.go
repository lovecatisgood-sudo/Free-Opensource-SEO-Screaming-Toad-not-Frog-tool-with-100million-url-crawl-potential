package customaudit

import (
	"strings"
	"testing"
)

func TestCSSExtractionAndSafeRegexCondition(t *testing.T) {
	d := Definition{ID: "prices", Name: "Prices", Enabled: true, Mode: "raw", SelectorKind: "css", Selector: "main > .product[data-live] .price", Extraction: Extraction{Kind: "text"}, Condition: Condition{Kind: "regex", Pattern: `^\$[0-9]+$`}, Finding: &Finding{Severity: "warning", Message: "Price observed"}, Limits: DefaultLimits()}
	result, err := Execute(d, []byte(`<main><div class="product" data-live><span class="price">$42</span></div></main>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Values) != 1 || result.Values[0] != "$42" || !result.Finding {
		t.Fatalf("result=%+v", result)
	}
}

func TestXPathSubsetAndAttributeExtraction(t *testing.T) {
	d := Definition{ID: "links", Name: "Links", Enabled: true, Mode: "raw", SelectorKind: "xpath", Selector: `//main//a[@rel='next']`, Extraction: Extraction{Kind: "attribute", Attribute: "href"}, Condition: Condition{Kind: "exists"}, Limits: DefaultLimits()}
	result, err := Execute(d, []byte(`<main><div><a rel="next" href="/two">Next</a></div></main>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Values) != 1 || result.Values[0] != "/two" {
		t.Fatalf("result=%+v", result)
	}
}

func TestBudgetsTruncateEvidence(t *testing.T) {
	d := Definition{ID: "bounded", Name: "Bounded", Mode: "raw", SelectorKind: "css", Selector: "p", Extraction: Extraction{Kind: "text"}, Condition: Condition{Kind: "always"}, Limits: Limits{MaximumMatches: 2, MaximumValueBytes: 4, MaximumTotalBytes: 8}}
	result, err := Execute(d, []byte(`<p>abcdef</p><p>ghijkl</p><p>mnop</p>`))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Truncated || result.MatchCount != 3 || strings.Join(result.Values, ",") != "abcd,ghij" {
		t.Fatalf("result=%+v", result)
	}
}

func TestImportRejectsUnknownFieldsAndUnsafeAttributes(t *testing.T) {
	if _, err := Import([]byte(`{"schema_version":1,"definitions":[],"script":"x"}`)); err == nil {
		t.Fatal("unknown field accepted")
	}
	d := Definition{ID: "bad", Name: "Bad", Mode: "raw", SelectorKind: "css", Selector: "a", Extraction: Extraction{Kind: "attribute", Attribute: "onclick"}, Condition: Condition{Kind: "always"}, Limits: DefaultLimits()}
	if err := d.NormalizeAndValidate(); err == nil {
		t.Fatal("event attribute accepted")
	}
}
