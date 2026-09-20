// Package jobs tracks the lifecycle of scheduled job runs (running, completed,
// failed) with processed counts so that long-running operations such as billing
// and deferral recognition can be checkpointed and resumed after interruption.
package jobs
