package accounting

import (
	"context"
	"encoding/xml"
	"strconv"
	"strings"
	"time"
)

type ImportCAMTRequest struct {
	OrganizationID uint64
	JournalID      uint64
	Date           time.Time
	RawXML         []byte
}

type camtDocument struct {
	Statements []camtStatement `xml:"BkToCstmrStmt>Stmt"`
}

type camtStatement struct {
	Entries []camtEntry `xml:"Ntry"`
}

type camtEntry struct {
	Amount      camtAmount   `xml:"Amt"`
	CreditDebit string       `xml:"CdtDbtInd"`
	BookingDate camtBookDate `xml:"BookgDt"`
	AccountRef  string       `xml:"AcctSvcrRef"`
	Details     camtDetails  `xml:"NtryDtls"`
}

type camtAmount struct {
	Currency string `xml:"Ccy,attr"`
	Value    string `xml:",chardata"`
}

type camtBookDate struct {
	Date     string `xml:"Dt"`
	DateTime string `xml:"DtTm"`
}

type camtDetails struct {
	Transactions []camtTransaction `xml:"TxDtls"`
}

type camtTransaction struct {
	EndToEndID string `xml:"Refs>EndToEndId"`
	TxID       string `xml:"Refs>TxId"`
	Narration  string `xml:"AddtlNtryInf"`
}

func (s BankStatementService) ImportCAMT(ctx context.Context, request ImportCAMTRequest) (*BankStatement, []*BankStatementLine, error) {
	if len(request.RawXML) == 0 {
		return nil, nil, ErrInvalidCAMT
	}
	var document camtDocument
	if err := xml.Unmarshal(request.RawXML, &document); err != nil {
		return nil, nil, ErrInvalidCAMT
	}
	lines := make([]BankStatementLineRequest, 0)
	for _, statement := range document.Statements {
		for _, entry := range statement.Entries {
			line, ok := camtLine(entry)
			if !ok {
				continue
			}
			if line.Ref != "" {
				existing, err := s.lines.Search(ctx, "ref", line.Ref)
				if err != nil {
					return nil, nil, err
				}
				if existing != nil {
					continue
				}
			}
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return nil, nil, ErrStatementNoLines
	}
	created, err := s.Create(ctx, CreateBankStatementRequest{
		OrganizationID: request.OrganizationID,
		JournalID:      request.JournalID,
		Date:           request.Date,
		Lines:          lines,
	})
	if err != nil {
		return nil, nil, err
	}
	unreconciled, err := s.Match(ctx, created.ID)
	if err != nil {
		return nil, nil, err
	}
	return created, unreconciled, nil
}

func camtLine(entry camtEntry) (BankStatementLineRequest, bool) {
	value, err := strconv.ParseFloat(strings.TrimSpace(entry.Amount.Value), 64)
	if err != nil || value == 0 {
		return BankStatementLineRequest{}, false
	}
	if entry.CreditDebit == "DBIT" {
		value = -value
	}
	line := BankStatementLineRequest{Amount: value}
	if ccy := strings.TrimSpace(entry.Amount.Currency); ccy != "" {
		line.CurrencyCode = &ccy
	}
	if date, ok := camtDate(entry.BookingDate); ok {
		line.Date = &date
	}
	ref := firstNonEmpty(entry.Details.firstRef(), entry.AccountRef)
	if ref != "" {
		line.Ref = ref
	}
	line.Narration = entry.Details.firstNarration()
	return line, true
}

func camtDate(value camtBookDate) (time.Time, bool) {
	for _, raw := range []string{value.Date, value.DateTime} {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		for _, layout := range []string{"2006-01-02", time.RFC3339, "2006-01-02T15:04:05"} {
			if parsed, err := time.Parse(layout, raw); err == nil {
				return parsed, true
			}
		}
	}
	return time.Time{}, false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func (d camtDetails) firstRef() string {
	for _, tx := range d.Transactions {
		if ref := firstNonEmpty(tx.EndToEndID, tx.TxID); ref != "" {
			return ref
		}
	}
	return ""
}

func (d camtDetails) firstNarration() string {
	for _, tx := range d.Transactions {
		if strings.TrimSpace(tx.Narration) != "" {
			return tx.Narration
		}
	}
	return ""
}
