//go:build integration

// Package integration runs in-process integration tests against the real test
// PostgreSQL, exercising full business flows across domains plus the NFR
// concurrency, idempotency, performance and integrity suites. Run with:
// go test -tags integration ./test/integration/...
package integration
