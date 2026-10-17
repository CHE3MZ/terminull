// Package termtable renders tabular data in five built-in styles:
// bordered (unicode box, default), rounded, ascii, markdown, and
// borderless padded columns for scripts. Rows normalize to the header
// width, empty tables render nothing, and unknown styles error.
//
// The renderer is private under termtable/internal/table, written for
// this module: no third-party table code and no attribution duties
// surface in this API.
package termtable
