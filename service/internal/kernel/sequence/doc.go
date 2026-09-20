// Package sequence manages document numbering from doc_sequences. It reserves
// the next number transactionally with row locks and period-based resets,
// returning stable human-readable document numbers.
package sequence
