// Package table merges two table engines behind one facade: aquasecurity
// (StyleAqua, the default: bordered unicode tables) and rodaine
// (StyleRodaine: borderless padded columns, most script-friendly).
// Engines live in subpackages with their upstream APIs intact; new
// styles arrive as new Style values without breaking callers.
//
// Fork attributions:
//
//	engine aqua: github.com/aquasecurity/table, MIT, Copyright (c) 2022 Aqua Security
//	engine rodaine: github.com/rodaine/table, MIT, Copyright (c) 2015 Chris Roche
package table
