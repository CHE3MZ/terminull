// Package toml implements a koanf.Parser for the TOML subset termconfig
// accepts: comments, dotted keys, tables, arrays of tables, basic and
// literal strings, integers, floats, booleans, arrays, and inline tables.
// Multi-line strings are supported; dates and times become time.Time.
package toml

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// TOML implements a TOML parser.
type TOML struct{}

// Parser returns a TOML parser.
func Parser() *TOML { return &TOML{} }

// Unmarshal parses TOML bytes into a config map.
func (p *TOML) Unmarshal(b []byte) (map[string]any, error) {
	return Parse(b)
}

// Marshal writes a config map as TOML tables.
func (p *TOML) Marshal(o map[string]any) ([]byte, error) {
	var b bytes.Buffer
	if err := writeTable(&b, nil, o); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

type parser struct {
	src  []byte
	pos  int
	line int
}

// Parse parses the supported TOML subset.
func Parse(src []byte) (map[string]any, error) {
	p := &parser{src: src, line: 1}
	root := map[string]any{}
	cur := root
	for {
		p.skipWSP()
		if p.eof() {
			return root, nil
		}
		if p.peek() == '[' {
			path, array, err := p.parseHeader()
			if err != nil {
				return nil, err
			}
			if array {
				next, err := appendTable(root, path)
				if err != nil {
					return nil, p.err(err)
				}
				cur = next
			} else {
				next, err := table(root, path)
				if err != nil {
					return nil, p.err(err)
				}
				cur = next
			}
		} else {
			if err := p.parseKeyValue(cur); err != nil {
				return nil, err
			}
		}
		if err := p.endLine(); err != nil {
			return nil, err
		}
	}
}

func (p *parser) parseHeader() ([]string, bool, error) {
	p.pos++ // [
	array := false
	if p.peek() == '[' {
		array = true
		p.pos++
	}
	p.skipSpace()
	path, err := p.parseKey()
	if err != nil {
		return nil, false, err
	}
	p.skipSpace()
	if p.peek() != ']' {
		return nil, false, p.errf("expected ]")
	}
	p.pos++
	if array {
		if p.peek() != ']' {
			return nil, false, p.errf("expected ]]")
		}
		p.pos++
	}
	return path, array, nil
}

func (p *parser) parseKeyValue(dst map[string]any) error {
	path, err := p.parseKey()
	if err != nil {
		return err
	}
	p.skipSpace()
	if p.peek() != '=' {
		return p.errf("expected =")
	}
	p.pos++
	p.skipSpace()
	value, err := p.parseValue()
	if err != nil {
		return err
	}
	return p.err(setPath(dst, path, value))
}

func (p *parser) parseKey() ([]string, error) {
	var parts []string
	for {
		p.skipSpace()
		var part string
		switch p.peek() {
		case '"', '\'':
			s, err := p.parseString()
			if err != nil {
				return nil, err
			}
			part = s
		default:
			start := p.pos
			for !p.eof() {
				c := p.peek()
				if isBareKey(c) {
					p.pos++
					continue
				}
				break
			}
			part = string(p.src[start:p.pos])
		}
		if part == "" {
			return nil, p.errf("empty key")
		}
		parts = append(parts, part)
		p.skipSpace()
		if p.peek() != '.' {
			return parts, nil
		}
		p.pos++
	}
}

func (p *parser) parseValue() (any, error) {
	if p.eof() {
		return nil, p.errf("missing value")
	}
	switch p.peek() {
	case '"', '\'':
		return p.parseString()
	case '[':
		return p.parseArray()
	case '{':
		return p.parseInlineTable()
	}

	start := p.pos
	for !p.eof() {
		c := p.peek()
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '#' || c == ',' || c == ']' || c == '}' {
			break
		}
		p.pos++
	}
	token := string(p.src[start:p.pos])
	if token == "" {
		return nil, p.errf("missing value")
	}
	switch token {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "inf", "+inf":
		return math.Inf(1), nil
	case "-inf":
		return math.Inf(-1), nil
	case "nan", "+nan", "-nan":
		return math.NaN(), nil
	}
	if v, ok := parseNumber(token); ok {
		return v, nil
	}
	if v, ok := parseDate(token); ok {
		return v, nil
	}
	return nil, p.errf("invalid value %q", token)
}

func (p *parser) parseArray() ([]any, error) {
	p.pos++ // [
	var out []any
	for {
		p.skipWSP()
		if p.eof() {
			return nil, p.errf("unterminated array")
		}
		if p.peek() == ']' {
			p.pos++
			return out, nil
		}
		v, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
		p.skipWSP()
		switch p.peek() {
		case ',':
			p.pos++
		case ']':
			p.pos++
			return out, nil
		default:
			return nil, p.errf("expected , or ] in array")
		}
	}
}

func (p *parser) parseInlineTable() (map[string]any, error) {
	p.pos++ // {
	out := map[string]any{}
	for {
		p.skipWSP()
		if p.eof() {
			return nil, p.errf("unterminated inline table")
		}
		if p.peek() == '}' {
			p.pos++
			return out, nil
		}
		path, err := p.parseKey()
		if err != nil {
			return nil, err
		}
		p.skipSpace()
		if p.peek() != '=' {
			return nil, p.errf("expected = in inline table")
		}
		p.pos++
		p.skipSpace()
		v, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		if err := setPath(out, path, v); err != nil {
			return nil, p.err(err)
		}
		p.skipWSP()
		switch p.peek() {
		case ',':
			p.pos++
		case '}':
			p.pos++
			return out, nil
		default:
			return nil, p.errf("expected , or } in inline table")
		}
	}
}

func (p *parser) parseString() (string, error) {
	quote := p.peek()
	if p.pos+2 < len(p.src) && p.src[p.pos+1] == quote && p.src[p.pos+2] == quote {
		return p.parseMultiline(quote)
	}
	p.pos++
	var b strings.Builder
	for {
		if p.eof() {
			return "", p.errf("unterminated string")
		}
		c := p.peek()
		if c == '\n' || c == '\r' {
			return "", p.errf("newline in string")
		}
		p.pos++
		if c == quote {
			return b.String(), nil
		}
		if quote == '\'' || c != '\\' {
			b.WriteByte(c)
			continue
		}
		if p.eof() {
			return "", p.errf("unterminated escape")
		}
		esc := p.peek()
		p.pos++
		switch esc {
		case 'b':
			b.WriteByte('\b')
		case 't':
			b.WriteByte('\t')
		case 'n':
			b.WriteByte('\n')
		case 'f':
			b.WriteByte('\f')
		case 'r':
			b.WriteByte('\r')
		case '"':
			b.WriteByte('"')
		case '\\':
			b.WriteByte('\\')
		case 'u', 'U':
			n := 4
			if esc == 'U' {
				n = 8
			}
			if p.pos+n > len(p.src) {
				return "", p.errf("short unicode escape")
			}
			v, err := strconv.ParseUint(string(p.src[p.pos:p.pos+n]), 16, 32)
			if err != nil || !utf8.ValidRune(rune(v)) {
				return "", p.errf("invalid unicode escape")
			}
			p.pos += n
			b.WriteRune(rune(v))
		default:
			return "", p.errf("invalid escape \\%c", esc)
		}
	}
}

func (p *parser) parseMultiline(quote byte) (string, error) {
	p.pos += 3
	if p.peek() == '\r' {
		p.pos++
	}
	if p.peek() == '\n' {
		p.pos++
		p.line++
	}
	var b strings.Builder
	for {
		if p.eof() {
			return "", p.errf("unterminated multiline string")
		}
		if p.pos+2 < len(p.src) && p.src[p.pos] == quote && p.src[p.pos+1] == quote && p.src[p.pos+2] == quote {
			p.pos += 3
			return b.String(), nil
		}
		c := p.peek()
		p.pos++
		if c == '\n' {
			p.line++
		}
		if quote == '\'' || c != '\\' {
			b.WriteByte(c)
			continue
		}
		if p.eof() {
			return "", p.errf("unterminated escape")
		}
		esc := p.peek()
		p.pos++
		switch esc {
		case 'b':
			b.WriteByte('\b')
		case 't':
			b.WriteByte('\t')
		case 'n':
			b.WriteByte('\n')
		case 'f':
			b.WriteByte('\f')
		case 'r':
			b.WriteByte('\r')
		case '"':
			b.WriteByte('"')
		case '\\':
			b.WriteByte('\\')
		case '\n':
			p.line++
			for !p.eof() && (p.peek() == ' ' || p.peek() == '\t' || p.peek() == '\r' || p.peek() == '\n') {
				if p.peek() == '\n' {
					p.line++
				}
				p.pos++
			}
		default:
			return "", p.errf("invalid escape \\%c", esc)
		}
	}
}

func (p *parser) endLine() error {
	p.skipSpace()
	if p.peek() == '#' {
		for !p.eof() && p.peek() != '\n' {
			p.pos++
		}
	}
	if p.eof() {
		return nil
	}
	if p.peek() == '\r' {
		p.pos++
	}
	if p.peek() != '\n' {
		return p.errf("expected newline")
	}
	p.pos++
	p.line++
	return nil
}

func (p *parser) skipSpace() {
	for !p.eof() && (p.peek() == ' ' || p.peek() == '\t') {
		p.pos++
	}
}

func (p *parser) skipWSP() {
	for !p.eof() {
		switch p.peek() {
		case ' ', '\t', '\r':
			p.pos++
		case '\n':
			p.pos++
			p.line++
		case '#':
			for !p.eof() && p.peek() != '\n' {
				p.pos++
			}
		default:
			return
		}
	}
}

func (p *parser) eof() bool { return p.pos >= len(p.src) }

func (p *parser) peek() byte {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *parser) err(err error) error {
	if err == nil {
		return nil
	}
	return p.errf("%v", err)
}

func (p *parser) errf(format string, args ...any) error {
	return fmt.Errorf("toml line %d: %s", p.line, fmt.Sprintf(format, args...))
}

func isBareKey(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-'
}

func parseNumber(token string) (any, bool) {
	clean := strings.ReplaceAll(token, "_", "")
	sign := 1
	body := clean
	if strings.HasPrefix(body, "+") {
		body = body[1:]
	} else if strings.HasPrefix(body, "-") {
		sign = -1
		body = body[1:]
	}
	lower := strings.ToLower(body)
	base := 10
	switch {
	case strings.HasPrefix(lower, "0x"):
		base, body = 16, body[2:]
	case strings.HasPrefix(lower, "0o"):
		base, body = 8, body[2:]
	case strings.HasPrefix(lower, "0b"):
		base, body = 2, body[2:]
	}
	if base != 10 {
		v, err := strconv.ParseInt(body, base, 64)
		if err != nil {
			return nil, false
		}
		return int64(sign) * v, true
	}
	if strings.ContainsAny(clean, ".eE") {
		v, err := strconv.ParseFloat(clean, 64)
		if err != nil {
			return nil, false
		}
		return v, true
	}
	v, err := strconv.ParseInt(clean, 10, 64)
	if err != nil {
		return nil, false
	}
	return v, true
}

func parseDate(token string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"15:04:05.999999999",
		"15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, token); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func setPath(dst map[string]any, path []string, value any) error {
	m := dst
	for _, key := range path[:len(path)-1] {
		next, ok := m[key]
		if !ok {
			child := map[string]any{}
			m[key] = child
			m = child
			continue
		}
		switch v := next.(type) {
		case map[string]any:
			m = v
		case []any:
			last, ok := lastTable(v)
			if !ok {
				return fmt.Errorf("key %q is not a table", key)
			}
			m = last
		default:
			return fmt.Errorf("key %q is not a table", key)
		}
	}
	last := path[len(path)-1]
	if _, exists := m[last]; exists {
		return fmt.Errorf("duplicate key %q", last)
	}
	m[last] = value
	return nil
}

func table(root map[string]any, path []string) (map[string]any, error) {
	m := root
	for _, key := range path {
		next, ok := m[key]
		if !ok {
			child := map[string]any{}
			m[key] = child
			m = child
			continue
		}
		switch v := next.(type) {
		case map[string]any:
			m = v
		case []any:
			last, ok := lastTable(v)
			if !ok {
				return nil, fmt.Errorf("key %q is not a table", key)
			}
			m = last
		default:
			return nil, fmt.Errorf("key %q is not a table", key)
		}
	}
	return m, nil
}

func appendTable(root map[string]any, path []string) (map[string]any, error) {
	parent, err := table(root, path[:len(path)-1])
	if err != nil {
		return nil, err
	}
	key := path[len(path)-1]
	var arr []any
	if existing, ok := parent[key]; ok {
		arr, ok = existing.([]any)
		if !ok {
			return nil, fmt.Errorf("key %q is not an array of tables", key)
		}
	}
	next := map[string]any{}
	parent[key] = append(arr, next)
	return next, nil
}

func lastTable(v []any) (map[string]any, bool) {
	if len(v) == 0 {
		return nil, false
	}
	m, ok := v[len(v)-1].(map[string]any)
	return m, ok
}

func writeTable(b *bytes.Buffer, prefix []string, m map[string]any) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var tables []string
	for _, key := range keys {
		switch v := m[key].(type) {
		case map[string]any:
			tables = append(tables, key)
		case []any:
			if len(v) > 0 {
				if _, ok := v[0].(map[string]any); ok {
					tables = append(tables, key)
					continue
				}
			}
			if err := writePair(b, key, v); err != nil {
				return err
			}
		default:
			if err := writePair(b, key, v); err != nil {
				return err
			}
		}
	}
	for _, key := range tables {
		path := append(append([]string{}, prefix...), key)
		switch v := m[key].(type) {
		case map[string]any:
			fmt.Fprintf(b, "\n[%s]\n", strings.Join(quoteKeys(path), "."))
			if err := writeTable(b, path, v); err != nil {
				return err
			}
		case []any:
			for _, item := range v {
				fmt.Fprintf(b, "\n[[%s]]\n", strings.Join(quoteKeys(path), "."))
				if err := writeTable(b, path, item.(map[string]any)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func writePair(b *bytes.Buffer, key string, v any) error {
	fmt.Fprintf(b, "%s = ", quoteKey(key))
	switch value := v.(type) {
	case string:
		fmt.Fprintf(b, "%q\n", value)
	case bool:
		fmt.Fprintf(b, "%t\n", value)
	case int:
		fmt.Fprintf(b, "%d\n", value)
	case int64:
		fmt.Fprintf(b, "%d\n", value)
	case uint64:
		fmt.Fprintf(b, "%d\n", value)
	case float64:
		fmt.Fprintf(b, "%g\n", value)
	case time.Time:
		fmt.Fprintf(b, "%q\n", value.Format(time.RFC3339Nano))
	case []any:
		fmt.Fprint(b, "[")
		for i, item := range value {
			if i > 0 {
				fmt.Fprint(b, ", ")
			}
			s, err := scalarString(item)
			if err != nil {
				return err
			}
			fmt.Fprint(b, s)
		}
		fmt.Fprint(b, "]\n")
	case nil:
		fmt.Fprint(b, "\"\"\n")
	default:
		return fmt.Errorf("toml: unsupported value %T", v)
	}
	return nil
}

func scalarString(v any) (string, error) {
	switch value := v.(type) {
	case string:
		return strconv.Quote(value), nil
	case bool:
		return strconv.FormatBool(value), nil
	case int:
		return strconv.Itoa(value), nil
	case int64:
		return strconv.FormatInt(value, 10), nil
	case float64:
		return strconv.FormatFloat(value, 'g', -1, 64), nil
	case time.Time:
		return strconv.Quote(value.Format(time.RFC3339Nano)), nil
	case nil:
		return `""`, nil
	default:
		return "", fmt.Errorf("toml: unsupported array value %T", v)
	}
}

func quoteKeys(parts []string) []string {
	out := make([]string, len(parts))
	for i, part := range parts {
		out[i] = quoteKey(part)
	}
	return out
}

func quoteKey(key string) string {
	if key == "" {
		return strconv.Quote(key)
	}
	for i := 0; i < len(key); i++ {
		if !isBareKey(key[i]) {
			return strconv.Quote(key)
		}
	}
	return key
}
