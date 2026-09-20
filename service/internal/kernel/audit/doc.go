// Package audit appends an immutable audit trail on data changes: a GORM
// plugin snapshots before/after state and a service records diffs, while an
// Audited decorator enforces immutability of posted entities.
package audit
