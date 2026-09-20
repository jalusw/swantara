// Package main is the HTTP API entrypoint for the Swantara ERP service.
// It loads configuration, connects the database, wires the DAOs, services and
// handlers of every domain onto a Fiber v3 router under /api/v1, and serves the
// API with authentication and organization membership guards until shutdown.
package main
