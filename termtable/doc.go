// Package termtable renders tabular data in five built-in styles:
// bordered (unicode box, default), rounded, ascii, markdown, and
// borderless padded columns for scripts. Rows normalize to the header
// width, empty tables render nothing, and unknown styles error.
//
// The rendering engines are vendored privately under termtable/internal
// and do not surface in this API. Their MIT license texts and copyright
// notices live with each engine.
package termtable
