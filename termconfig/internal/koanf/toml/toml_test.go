package toml

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()

	src := []byte(`
# app config
name = "demo"
debug = true
retries = 3
ratio = 1.5
tags = ["a", "b"]
owner.name = "acme"
owner.email = "dev@example.test"
inline = { a = 1, b = "two" }
message = """
hello
world
"""
when = 2026-10-24T12:00:00Z
local = 2026-10-24 12:00:00
day = 2026-10-24
"" = "blank"

[database]
host = "localhost"
ports = [5432, 5433]

[servers.alpha]
ip = "10.0.0.1"
role = 'primary'


[[products]]
name = "hammer"

[[products]]
name = "nail"
price = 0.01
`)
	got, err := Parser().Unmarshal(src)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if got["name"] != "demo" || got["debug"] != true {
		t.Fatalf("root scalars = %#v", got)
	}
	if got["retries"] != int64(3) || got["ratio"] != 1.5 {
		t.Fatalf("numbers = %#v, %#v", got["retries"], got["ratio"])
	}
	if !reflect.DeepEqual(got["tags"], []any{"a", "b"}) {
		t.Fatalf("tags = %#v", got["tags"])
	}
	owner := got["owner"].(map[string]any)
	if owner["name"] != "acme" || owner["email"] != "dev@example.test" {
		t.Fatalf("owner = %#v", owner)
	}
	db := got["database"].(map[string]any)
	if db["host"] != "localhost" || !reflect.DeepEqual(db["ports"], []any{int64(5432), int64(5433)}) {
		t.Fatalf("database = %#v", db)
	}
	servers := got["servers"].(map[string]any)["alpha"].(map[string]any)
	if servers["role"] != "primary" {
		t.Fatalf("servers.alpha = %#v", servers)
	}
	products := got["products"].([]any)
	if len(products) != 2 || products[0].(map[string]any)["name"] != "hammer" || products[1].(map[string]any)["price"] != 0.01 {
		t.Fatalf("products = %#v", products)
	}
	inline := got["inline"].(map[string]any)
	if inline["a"] != int64(1) || inline["b"] != "two" {
		t.Fatalf("inline = %#v", inline)
	}
	if got["message"] != "hello\nworld\n" {
		t.Fatalf("message = %q", got["message"])
	}
	if got[""] != "blank" {
		t.Fatalf("empty key = %#v", got[""])
	}
	when, ok := got["when"].(time.Time)
	if !ok || !when.Equal(time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("when = %#v", when)
	}
	if local, ok := got["local"].(time.Time); !ok || local.Format("2006-01-02 15:04:05") != "2026-10-24 12:00:00" {
		t.Fatalf("local = %#v", got["local"])
	}
	if day, ok := got["day"].(time.Time); !ok || day.Format("2006-01-02") != "2026-10-24" {
		t.Fatalf("day = %#v", got["day"])
	}
}

func TestParseErrors(t *testing.T) {
	t.Parallel()

	for _, src := range []string{
		"a = 1\na = 2\n",
		"a = \n",
		"a = \"unterminated\n",
		"[a\nb = 1\n",
		"[a]\nx = 1\n[a]\ny = 2\n",
		"a = 0123\n",
	} {
		if _, err := Parser().Unmarshal([]byte(src)); err == nil {
			t.Errorf("Unmarshal(%q) returned nil error", src)
		}
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	t.Parallel()

	in := map[string]any{
		"name":  "demo",
		"debug": true,
		"n":     int64(4),
		"ratio": 2.5,
		"tags":  []any{"a", "b"},
		"db":    map[string]any{"host": "localhost", "port": int64(5432)},
		"items": []any{map[string]any{"name": "one"}},
	}
	data, err := Parser().Marshal(in)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if !strings.Contains(string(data), "[db]") || !strings.Contains(string(data), "[[items]]") {
		t.Fatalf("Marshal() =\n%s", data)
	}
	got, err := Parser().Unmarshal(data)
	if err != nil {
		t.Fatalf("round-trip Unmarshal() error = %v\n%s", err, data)
	}
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("round trip = %#v, want %#v", got, in)
	}
}
