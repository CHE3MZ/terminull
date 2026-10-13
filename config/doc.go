// Package config loads layered configuration for terminull apps.
//
// Precedence, low to high: Defaults, config files, environment, Overrides
// (parsed CLI flags, handed over by package app).
//
// Files are searched in order and every existing file is layered, so
// project-local settings win over user-global ones:
//
//  1. explicit --config path (missing file is an error)
//  2. home dotfile: ~/.APPNAME.yaml
//  3. user config dir: %APPDATA%/APPNAME/APPNAME.yaml,
//     ~/.config/APPNAME/APPNAME.yaml, ~/Library/Application Support/...
//  4. working directory: ./APPNAME.yaml then ./.APPNAME.yaml
//
// Environment variables use APPNAME_ prefixes: APPNAME_DB_HOST becomes
// db.host. Numeric-looking and true/false values become typed.
//
// Unknown keys fail loudly, but only at UnmarshalStrict time: typos in
// files or env surface as errors naming the key instead of silent zero
// values.
package config
