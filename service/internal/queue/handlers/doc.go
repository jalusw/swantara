// Package handlers implements the Asynq worker logic: each handler unmarshals
// a queued task payload, restores the request context and drives the underlying
// domain service for email, integration events, deferral recognition and
// subscription billing.
package handlers
