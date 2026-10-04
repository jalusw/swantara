//go:build e2e

// Package e2e runs black-box end-to-end HTTP tests against a live Swantara API
// (default http://localhost:8080), covering auth, users, tenant-scoped routes
// and the core business flows: CRM lead-to-opportunity, procure-to-pay and
// order-to-cash, plus reference catalogs, payroll catalogs, service
// maintenance, POS payment accounts, procurement catalogs, audit logs,
// integration events, journal entries, PDC instruments, period-close
// reporting, interorganization rules, consolidation runs, inventory,
// manufacturing and procurement routes.
// Run with: go test -tags e2e ./test/e2e/...
package e2e
