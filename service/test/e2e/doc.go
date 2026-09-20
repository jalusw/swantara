//go:build e2e

// Package e2e runs black-box end-to-end HTTP tests against a live Swantara API
// (default http://localhost:8080), covering auth, users, tenant-scoped routes
// and the core business flows: CRM lead-to-opportunity, procure-to-pay and
// order-to-cash. Run with: go test -tags e2e ./test/e2e/...
package e2e
