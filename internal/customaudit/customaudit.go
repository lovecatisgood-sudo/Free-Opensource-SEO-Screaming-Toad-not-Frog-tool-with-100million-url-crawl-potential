// Package customaudit implements bounded, declarative document extraction.
// It deliberately supports no scripts, callbacks, network access, or arbitrary
// file paths; definitions are safe to import and execute inside a crawl.
package customaudit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

const SchemaVersion = 1

type Definition struct {
	SchemaVersion int        `json:"schema_version"`
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Enabled       bool       `json:"enabled"`
	Mode          string     `json:"mode"`          // raw or rendered
	SelectorKind  string     `json:"selector_kind"` // css or xpath
	Selector      string     `json:"selector"`
	Extraction    Extraction `json:"extraction"`
	Condition     Condition  `json:"condition"`
	Finding       *Finding   `json:"finding,omitempty"`
	Limits        Limits     `json:"limits"`
}

type Extraction struct {
	Kind      string `json:"kind"` // text, html, attribute, count
	Attribute string `json:"attribute,omitempty"`
}

type Condition struct {
	Kind    string `json:"kind"` // always, exists, absent, equals, contains, regex
	Pattern string `json:"pattern,omitempty"`
}

type Finding struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type Limits struct {
	MaximumMatches    int `json:"maximum_matches"`
	MaximumValueBytes int `json:"maximum_value_bytes"`
	MaximumTotalBytes int `json:"maximum_total_bytes"`
}

func DefaultLimits() Limits {
	return Limits{MaximumMatches: 100, MaximumValueBytes: 4096, MaximumTotalBytes: 64 << 10}
}

type Result struct {
	DefinitionID string   `json:"definition_id"`
	Mode         string   `json:"mode"`
	Values       []string `json:"values"`
	MatchCount   int      `json:"match_count"`
	ConditionMet bool     `json:"condition_met"`
	Finding      bool     `json:"finding"`
	Truncated    bool     `json:"truncated"`
}

func (d *Definition) NormalizeAndValidate() error {
	if d.SchemaVersion == 0 {
		d.SchemaVersion = SchemaVersion
	}
	if d.Limits == (Limits{}) {
		d.Limits = DefaultLimits()
	}
	if d.SchemaVersion != SchemaVersion {
		return errors.New("unsupported custom-audit schema version")
	}
	if !validID(d.ID) || strings.TrimSpace(d.Name) == "" || len(d.Name) > 200 {
		return errors.New("custom-audit ID or name is invalid")
	}
	if d.Mode != "raw" && d.Mode != "rendered" {
		return errors.New("custom-audit mode must be raw or rendered")
	}
	if d.SelectorKind != "css" && d.SelectorKind != "xpath" {
		return errors.New("selector kind must be css or xpath")
	}
	if d.Selector == "" || len(d.Selector) > 512 {
		return errors.New("selector must contain 1 to 512 bytes")
	}
	if d.Extraction.Kind == "" {
		d.Extraction.Kind = "text"
	}
	if d.Extraction.Kind != "text" && d.Extraction.Kind != "html" && d.Extraction.Kind != "attribute" && d.Extraction.Kind != "count" {
		return errors.New("unsupported extraction kind")
	}
	if d.Extraction.Kind == "attribute" && (!validAttribute(d.Extraction.Attribute) || len(d.Extraction.Attribute) > 100) {
		return errors.New("attribute extraction requires a safe attribute name")
	}
	if d.Condition.Kind == "" {
		d.Condition.Kind = "always"
	}
	if !slices.Contains([]string{"always", "exists", "absent", "equals", "contains", "regex"}, d.Condition.Kind) {
		return errors.New("unsupported condition kind")
	}
	if len(d.Condition.Pattern) > 1024 {
		return errors.New("condition pattern exceeds 1024 bytes")
	}
	if d.Condition.Kind == "regex" {
		if _, err := regexp.Compile(d.Condition.Pattern); err != nil {
			return fmt.Errorf("compile condition regex: %w", err)
		}
	}
	if d.Finding != nil && (!slices.Contains([]string{"info", "warning", "error"}, d.Finding.Severity) || strings.TrimSpace(d.Finding.Message) == "" || len(d.Finding.Message) > 500) {
		return errors.New("custom finding is invalid")
	}
	if d.Limits.MaximumMatches < 1 || d.Limits.MaximumMatches > 1000 || d.Limits.MaximumValueBytes < 1 || d.Limits.MaximumValueBytes > 64<<10 || d.Limits.MaximumTotalBytes < d.Limits.MaximumValueBytes || d.Limits.MaximumTotalBytes > 4<<20 {
		return errors.New("custom-audit limits are outside safe bounds")
	}
	if d.SelectorKind == "css" {
		_, err := parseCSS(d.Selector)
		return err
	}
	_, err := parseXPath(d.Selector)
	return err
}

func Execute(definition Definition, document []byte) (Result, error) {
	if err := definition.NormalizeAndValidate(); err != nil {
		return Result{}, err
	}
	if len(document) > 8<<20 {
		return Result{}, errors.New("custom-audit document exceeds 8 MiB")
	}
	root, err := html.Parse(bytes.NewReader(document))
	if err != nil {
		return Result{}, err
	}
	var steps []selectorStep
	if definition.SelectorKind == "css" {
		steps, err = parseCSS(definition.Selector)
	} else {
		steps, err = parseXPath(definition.Selector)
	}
	if err != nil {
		return Result{}, err
	}
	nodes := selectNodes(root, steps, definition.Limits.MaximumMatches+1)
	result := Result{DefinitionID: definition.ID, Mode: definition.Mode, MatchCount: len(nodes)}
	if len(nodes) > definition.Limits.MaximumMatches {
		nodes = nodes[:definition.Limits.MaximumMatches]
		result.Truncated = true
	}
	if definition.Extraction.Kind == "count" {
		result.Values = []string{fmt.Sprint(result.MatchCount)}
	} else {
		total := 0
		for _, node := range nodes {
			value := extract(node, definition.Extraction)
			if len(value) > definition.Limits.MaximumValueBytes {
				value = value[:definition.Limits.MaximumValueBytes]
				result.Truncated = true
			}
			if total+len(value) > definition.Limits.MaximumTotalBytes {
				result.Truncated = true
				break
			}
			result.Values = append(result.Values, value)
			total += len(value)
		}
	}
	result.ConditionMet = conditionMet(definition.Condition, result.Values, result.MatchCount)
	result.Finding = result.ConditionMet && definition.Finding != nil
	return result, nil
}

func Export(definitions []Definition) ([]byte, error) {
	for index := range definitions {
		if err := definitions[index].NormalizeAndValidate(); err != nil {
			return nil, fmt.Errorf("definition %d: %w", index, err)
		}
	}
	return json.MarshalIndent(struct {
		SchemaVersion int          `json:"schema_version"`
		Definitions   []Definition `json:"definitions"`
	}{SchemaVersion, definitions}, "", "  ")
}

func Import(body []byte) ([]Definition, error) {
	if len(body) > 1<<20 {
		return nil, errors.New("custom-audit import exceeds 1 MiB")
	}
	var envelope struct {
		SchemaVersion int          `json:"schema_version"`
		Definitions   []Definition `json:"definitions"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		return nil, err
	}
	if envelope.SchemaVersion != SchemaVersion || len(envelope.Definitions) > 100 {
		return nil, errors.New("custom-audit import version or count is invalid")
	}
	seen := map[string]struct{}{}
	for index := range envelope.Definitions {
		d := &envelope.Definitions[index]
		if err := d.NormalizeAndValidate(); err != nil {
			return nil, fmt.Errorf("definition %d: %w", index, err)
		}
		if _, ok := seen[d.ID]; ok {
			return nil, errors.New("duplicate custom-audit ID")
		}
		seen[d.ID] = struct{}{}
	}
	return envelope.Definitions, nil
}

func validID(value string) bool {
	ok, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,99}$`, value)
	return ok
}
func validAttribute(value string) bool {
	ok, _ := regexp.MatchString(`^[a-zA-Z_:][a-zA-Z0-9_:.-]*$`, value)
	return ok && !strings.HasPrefix(strings.ToLower(value), "on")
}

type selectorStep struct {
	combinator byte
	tag, id    string
	classes    []string
	attributes []attributeMatch
}
type attributeMatch struct {
	name, value string
	hasValue    bool
}

func parseCSS(value string) ([]selectorStep, error)   { return parseSelector(value, false) }
func parseXPath(value string) ([]selectorStep, error) { return parseSelector(value, true) }

func parseSelector(value string, xpath bool) ([]selectorStep, error) {
	value = strings.TrimSpace(value)
	if strings.Contains(value, ",") || strings.ContainsAny(value, "\x00\r\n") {
		return nil, errors.New("selector lists and control characters are not supported")
	}
	if xpath {
		if strings.HasPrefix(value, ".//") {
			value = value[1:]
		}
		if !strings.HasPrefix(value, "/") {
			return nil, errors.New("XPath must start with / or //")
		}
		var steps []selectorStep
		for len(value) > 0 {
			comb := byte('>')
			if strings.HasPrefix(value, "//") {
				comb = ' '
				value = value[2:]
			} else if strings.HasPrefix(value, "/") {
				value = value[1:]
			} else {
				return nil, errors.New("invalid XPath separator")
			}
			end := strings.Index(value, "/")
			token := value
			if end >= 0 {
				token = value[:end]
				value = value[end:]
			} else {
				value = ""
			}
			step, err := parseSimple(token)
			if err != nil {
				return nil, err
			}
			step.combinator = comb
			steps = append(steps, step)
		}
		if len(steps) == 0 {
			return nil, errors.New("empty XPath")
		}
		return steps, nil
	}
	var tokens []string
	var combinators []byte
	start := 0
	bracket := 0
	quote := rune(0)
	pending := byte(' ')
	runes := []rune(value)
	flush := func(end int) {
		if text := strings.TrimSpace(string(runes[start:end])); text != "" {
			tokens = append(tokens, text)
			combinators = append(combinators, pending)
			pending = ' '
		}
	}
	for i, r := range runes {
		if quote != 0 {
			if r == quote {
				quote = 0
			}
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			continue
		}
		if r == '[' {
			bracket++
		}
		if r == ']' {
			bracket--
		}
		if bracket == 0 && (r == '>' || r == ' ' || r == '\t') {
			flush(i)
			if r == '>' {
				pending = '>'
			}
			start = i + 1
		}
	}
	flush(len(runes))
	if bracket != 0 || quote != 0 || len(tokens) == 0 {
		return nil, errors.New("invalid CSS selector")
	}
	steps := make([]selectorStep, 0, len(tokens))
	for i, token := range tokens {
		step, err := parseSimple(token)
		if err != nil {
			return nil, err
		}
		step.combinator = combinators[i]
		steps = append(steps, step)
	}
	return steps, nil
}

func parseSimple(token string) (selectorStep, error) {
	var step selectorStep
	token = strings.TrimSpace(token)
	if token == "" {
		return step, errors.New("empty selector step")
	}
	for len(token) > 0 {
		switch token[0] {
		case '#', '.':
			kind := token[0]
			token = token[1:]
			n := strings.IndexAny(token, "#.[")
			if n < 0 {
				n = len(token)
			}
			value := token[:n]
			if value == "" {
				return step, errors.New("empty selector identifier")
			}
			if kind == '#' {
				step.id = value
			} else {
				step.classes = append(step.classes, value)
			}
			token = token[n:]
		case '[':
			end := strings.IndexByte(token, ']')
			if end < 0 {
				return step, errors.New("unterminated attribute selector")
			}
			expr := strings.TrimSpace(token[1:end])
			expr = strings.TrimPrefix(expr, "@")
			name, value, found := strings.Cut(expr, "=")
			name = strings.TrimSpace(name)
			value = strings.Trim(strings.TrimSpace(value), "\"'")
			if !validAttribute(name) {
				return step, errors.New("invalid selector attribute")
			}
			step.attributes = append(step.attributes, attributeMatch{name, value, found})
			token = token[end+1:]
		default:
			n := strings.IndexAny(token, "#.[")
			if n < 0 {
				n = len(token)
			}
			tag := token[:n]
			if tag != "*" && !validAttribute(tag) {
				return step, errors.New("invalid selector tag")
			}
			step.tag = strings.ToLower(tag)
			token = token[n:]
		}
	}
	return step, nil
}

func selectNodes(root *html.Node, steps []selectorStep, limit int) []*html.Node {
	current := []*html.Node{root}
	for index, step := range steps {
		var next []*html.Node
		seen := map[*html.Node]struct{}{}
		for _, parent := range current {
			candidates := children(parent)
			if index == 0 || step.combinator == ' ' {
				candidates = descendants(parent, limit)
			}
			for _, node := range candidates {
				if matches(node, step) {
					if _, ok := seen[node]; !ok {
						seen[node] = struct{}{}
						next = append(next, node)
						if len(next) >= limit {
							return next
						}
					}
				}
			}
		}
		current = next
	}
	return current
}
func children(node *html.Node) []*html.Node {
	var out []*html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			out = append(out, c)
		}
	}
	return out
}
func descendants(node *html.Node, limit int) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		for c := n.FirstChild; c != nil && len(out) < limit*20; c = c.NextSibling {
			if c.Type == html.ElementNode {
				out = append(out, c)
			}
			walk(c)
		}
	}
	walk(node)
	return out
}
func matches(node *html.Node, step selectorStep) bool {
	if node.Type != html.ElementNode {
		return false
	}
	if step.tag != "" && step.tag != "*" && !strings.EqualFold(node.Data, step.tag) {
		return false
	}
	if step.id != "" && attr(node, "id") != step.id {
		return false
	}
	classes := strings.Fields(attr(node, "class"))
	for _, wanted := range step.classes {
		if !slices.Contains(classes, wanted) {
			return false
		}
	}
	for _, wanted := range step.attributes {
		value, ok := attrOK(node, wanted.name)
		if !ok || wanted.hasValue && value != wanted.value {
			return false
		}
	}
	return true
}
func attrOK(node *html.Node, name string) (string, bool) {
	for _, a := range node.Attr {
		if strings.EqualFold(a.Key, name) {
			return a.Val, true
		}
	}
	return "", false
}
func attr(node *html.Node, name string) string { v, _ := attrOK(node, name); return v }
func extract(node *html.Node, extraction Extraction) string {
	switch extraction.Kind {
	case "attribute":
		return attr(node, extraction.Attribute)
	case "html":
		var b bytes.Buffer
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			_ = html.Render(&b, c)
		}
		return b.String()
	default:
		return nodeText(node)
	}
}
func nodeText(node *html.Node) string {
	var values []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			if v := strings.TrimSpace(n.Data); v != "" {
				values = append(values, v)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(node)
	return strings.Join(values, " ")
}
func conditionMet(c Condition, values []string, count int) bool {
	switch c.Kind {
	case "always", "exists":
		return count > 0
	case "absent":
		return count == 0
	case "equals":
		for _, v := range values {
			if v == c.Pattern {
				return true
			}
		}
	case "contains":
		for _, v := range values {
			if strings.Contains(v, c.Pattern) {
				return true
			}
		}
	case "regex":
		re, _ := regexp.Compile(c.Pattern)
		for _, v := range values {
			if re.MatchString(v) {
				return true
			}
		}
	}
	return false
}
