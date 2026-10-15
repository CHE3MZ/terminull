// Package retry reruns fallible operations (network calls, toolchain
// runs) with capped attempts and backoff. Permanent failures should be
// returned wrapped so callers can exclude them via Options.RetryIf.
package retry
