// Package plain renders borderless padded-column tables. It adapts
// github.com/rodaine/table (MIT, Copyright (c) 2015 Chris Roche); see
// LICENSE in this directory. It is a private engine for termtable and is
// not part of the public API. Divergences: CJK-safe width measurement by
// default and error returns on the print path.
package plain
