// Package output prints command results: tables, JSON, YAML, ASCII graphs,
// human sizes, and shell-split command lines.
//
// Tables use ASCII borders only, so output is byte-identical on Linux,
// macOS, and legacy Windows consoles. JSON and YAML marshal with stable
// key order for golden tests and pipes.
package output
