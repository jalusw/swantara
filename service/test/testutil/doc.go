// Package testutil bootstraps the integration test database: it loads the test
// config, connects Postgres, registers the audit plugin, runs goose migrations,
// seeds system roles and truncates the domain tables between test scenarios.
package testutil
