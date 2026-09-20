package accounting

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

const camtSample = `<?xml version="1.0" encoding="UTF-8"?>
<Document xmlns="urn:iso:std:iso:20022:tech:xsd:camt.053.001.02">
  <BkToCstmrStmt>
    <Stmt>
      <Id>STMT-1</Id>
      <Ntry>
        <Amt Ccy="USD">100.00</Amt>
        <CdtDbtInd>CRDT</CdtDbtInd>
        <BookgDt><Dt>2026-08-01</Dt></BookgDt>
        <NtryDtls><TxDtls>
          <Refs><EndToEndId>E2E-1</EndToEndId></Refs>
          <AddtlNtryInf>Customer payment</AddtlNtryInf>
        </TxDtls></NtryDtls>
      </Ntry>
      <Ntry>
        <Amt Ccy="USD">50.00</Amt>
        <CdtDbtInd>DBIT</CdtDbtInd>
        <BookgDt><Dt>2026-08-02</Dt></BookgDt>
        <NtryDtls><TxDtls>
          <Refs><EndToEndId>DUPLICATE</EndToEndId></Refs>
        </TxDtls></NtryDtls>
      </Ntry>
      <Ntry>
        <Amt Ccy="USD">25.00</Amt>
        <CdtDbtInd>CRDT</CdtDbtInd>
        <BookgDt><Dt>2026-08-03</Dt></BookgDt>
      </Ntry>
    </Stmt>
  </BkToCstmrStmt>
</Document>`

func TestBankStatementService_ImportCAMT_SkipsDuplicates(t *testing.T) {
	ctx := context.Background()
	var createdLines []*BankStatementLine
	statements := BankStatementDAOMock{
		CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, statement *BankStatement, lines []*BankStatementLine) (*BankStatement, error) {
			statement.ID = 7
			createdLines = lines
			return statement, nil
		},
	}
	statements.FindFunc = func(_ context.Context, _ uint64) (*BankStatement, error) {
		return &BankStatement{State: BankStatementStateOpen}, nil
	}
	lines := BankStatementLineDAOMock{}
	lines.SearchFunc = func(_ context.Context, _ string, value any) (*BankStatementLine, error) {
		if value == "DUPLICATE" {
			return &BankStatementLine{}, nil
		}
		return nil, nil
	}
	lines.ListUnreconciledByStatementFunc = func(_ context.Context, _ uint64) ([]*BankStatementLine, error) {
		return nil, nil
	}
	journals := dao.CRUDMock[reference.Journal]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{}, nil
		},
	}
	svc := NewBankStatementService(statements, lines, PaymentDAOMock{}, journals, TransactionerMock{})

	statement, _, err := svc.ImportCAMT(ctx, ImportCAMTRequest{OrganizationID: 10, JournalID: 20, RawXML: []byte(camtSample)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statement.ID != 7 {
		t.Errorf("statement id = %d, want 7", statement.ID)
	}
	if len(createdLines) != 2 {
		t.Fatalf("lines = %d, want 2", len(createdLines))
	}
	if createdLines[0].Amount.Float64() != 100 || createdLines[0].Ref == nil || *createdLines[0].Ref != "E2E-1" {
		t.Errorf("line 0 = %v/%v, want 100/E2E-1", createdLines[0].Amount, createdLines[0].Ref)
	}
	if createdLines[1].Amount.Float64() != 25 {
		t.Errorf("line 1 = %v, want 25", createdLines[1].Amount)
	}
}

func TestBankStatementService_ImportCAMT_RejectsInvalid(t *testing.T) {
	ctx := context.Background()
	svc := NewBankStatementService(BankStatementDAOMock{}, BankStatementLineDAOMock{}, PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})

	_, _, err := svc.ImportCAMT(ctx, ImportCAMTRequest{RawXML: []byte("not xml")})
	if helper.AssertError(t, err, true, ErrInvalidCAMT) {
		return
	}
}
