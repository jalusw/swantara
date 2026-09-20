package handler

import (
	"github.com/jalusw/swantara/apps/service/internal/accounting"
)

type JournalEntryHandler struct {
	poster accounting.PostingService
}

func NewJournalEntryHandler(poster accounting.PostingService) JournalEntryHandler {
	return JournalEntryHandler{poster: poster}
}

var accountMoveQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"journal_id":      {},
	"state":           {},
	"origin_type":     {},
	"origin_id":       {},
	"date":            {},
	"name":            {},
	"ref":             {},
	"created_at":      {},
	"updated_at":      {},
}
