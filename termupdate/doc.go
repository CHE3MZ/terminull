// Package termupdate checks release sources for newer versions.
//
// Checks are explicit: callers invoke Check, and nothing runs on package
// initialization, app startup, or a background timer. termupdate only
// reports release metadata; it never downloads or installs updates.
package termupdate
