// Package main is the background worker entrypoint for the Swantara ERP
// service. It runs an Asynq task server and registers handlers for queued
// email, subscription billing, deferral recognition and integration event
// dispatch tasks.
package main
