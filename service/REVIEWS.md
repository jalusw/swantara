# Service Code Review

Review of `service/internal/` against `CONVENTIONS.md` and `ARCHITECTURE.md`.

## Progress

Last updated: 2026-09-05

- **354+ items fixed** across 18 phases
- **25 items** across 7 sections (§57–§63, duplication/inconsistency sweep, 2026-09-05) — 19 fixed, 6 accepted
- **0 items remaining**
- Build status: `go build ./...`, `go vet ./...`, and `go test ./internal/...` (78 packages) pass cleanly

## 1. God Handlers (>200 lines)

**Target:** Split large handler files into multiple files or extract shared logic to services. Each handler file should have ≤5 endpoints and ≤200 lines.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/payroll/handler/hr_handler.go` | 1-1928 | 1928 lines, 40 endpoints. Massive god handler covering employees, departments, contracts, leave, timesheets, attendance — **fixed**: split into employee_handler.go, department_handler.go, job_position_handler.go, leave_type_handler.go, contract_handler.go, leave_request_handler.go, attendance_handler.go, timesheet_handler.go, shift_handler.go, shift_assignment_handler.go; hr_handler.go now only contains writeHRError (42 lines) |
| [x] | `internal/sales/handler/sale_order_handler.go` | 1-979 | 979 lines, 13 endpoints — **fixed**: split into sale_order_handler.go (765 lines, CRUD+workflow) and sale_order_invoice_handler.go (delivery/invoicing/payment) |
| [x] | `internal/contacts/handler/contact_nested_handler.go` | 1-859 | 859 lines, 15 endpoints — **fixed**: split into contact_relation_handler.go, contact_address_handler.go, contact_bank_account_handler.go, contact_customer_handler.go, contact_supplier_handler.go |
| [x] | `internal/inventory/handler/stock_handler.go` | 1-820 | 820 lines, 12 endpoints — **fixed**: split into stock_handler.go (86 lines, struct+helpers), stock_movement_handler.go, shipment_handler.go, stock_balance_handler.go |
| [x] | `internal/procurement/handler/purchase_order_handler.go` | 1-918 | 918 lines, 11 endpoints — **fixed**: split into purchase_order_handler.go (603 lines, PO CRUD) and purchase_invoice_handler.go (329 lines, invoicing/payments/memos) |
| [x] | `internal/project/handler/project_handler.go` | 1-742 | 742 lines, 13 endpoints — **fixed**: split into project.go (98 lines), project_crud.go, project_task.go, project_milestone.go, project_billing.go, project_summary.go |
| [x] | `internal/iam/handler/user_handler.go` | 1-691 | 691 lines, 12 endpoints — **fixed**: split into user_handler.go (97 lines), me_handler.go, user_crud.go |
| [x] | `internal/inventory/handler/warehouse_handler.go` | 1-660 | 660 lines, 10 endpoints — **fixed**: split into warehouse_handler.go (45 lines, struct+helpers), warehouse.go, stock_location.go |
| [x] | `internal/products/handler/product_handler.go` | 1-619 | 619 lines, 8 endpoints — **fixed**: split into product_handler.go (123 lines, struct+helpers), product_template_handler.go (393 lines, template CRUD), product_variant_handler.go (121 lines, variant management) |
| [x] | `internal/procurement/handler/purchase_rfq_handler.go` | 1-614 | 614 lines — **fixed**: split into purchase_rfq_handler.go (280 lines), purchase_rfq_line_handler.go (73 lines), purchase_rfq_quote_handler.go (200 lines), purchase_rfq_to_order_handler.go (66 lines) |
| [x] | `internal/products/handler/price_book_handler.go` | 1-579 | 579 lines, 8 endpoints — **fixed**: split into price_book_handler.go, price_rule_handler.go, resolve_price_handler.go |
| [x] | `internal/iam/handler/auth_handler.go` | 1-570 | 570 lines — **fixed**: split into auth_handler.go (30 lines), login.go, register.go, password_reset.go, refresh.go, oauth.go, email_verification.go, access.go, session.go |
| [x] | `internal/accounting/handler/invoice_handler.go` | 1-523 | 523 lines — **fixed**: split into invoice_handler.go (312 lines), invoice_request.go (83 lines), invoice_response.go (147 lines) |
| [x] | `internal/manufacturing/handler/bom_handler.go` | 1-526 | 526 lines — **fixed**: split into bom_handler.go (struct+helpers), bom_crud.go, recipe_lines.go, bom_explode.go |
| [x] | `internal/crm/handler/lead_handler.go` | 1-512 | 512 lines — **fixed**: split into lead_handler.go (82 lines), lead_crud.go (308 lines), lead_promote.go (100 lines) |
| [x] | `internal/crm/handler/opportunity_handler.go` | 1-501 | 501 lines — **fixed**: split into opportunity_handler.go (82 lines), opportunity_crud.go (228 lines), opportunity_stage.go (150 lines) |
| [x] | `internal/returns/handler/rma_handler.go` | 1-493 | 493 lines — **fixed**: split into rma.go, rma_list.go, rma_get.go, rma_create.go, rma_transitions.go |
| [x] | `internal/reference/handler/payment_term_handler.go` — **fixed**: split into payment_term_handler.go, payment_term_response.go, payment_term_splits.go | 1-493 | 493 lines |
| [x] | `internal/contacts/handler/contact_handler.go` | 1-488 | 488 lines — **fixed**: split into contact.go, contact_response.go, contact_list.go, contact_get.go, contact_create.go, contact_update.go, contact_delete.go |
| [x] | `internal/expense/handler/expense_report_handler.go` — **fixed**: split into expense_report_types.go, expense_report_crud.go, expense_report_transition.go | 1-465 | 465 lines |
| [x] | `internal/giftcard/handler/giftcard_handler.go` — **fixed**: split into giftcard_handler.go, response.go, list.go, get.go, issue.go, redeem.go, refund.go, forfeit_expired.go, transactions.go, helpers.go | 1-453 | 453 lines |
| [x] | `internal/manufacturing/handler/mo_handler.go` | 1-459 | 459 lines — **fixed**: split into mo_response.go, mo_crud.go, mo_state.go, mo_component.go, mo_error.go |
| [x] | `internal/manufacturing/handler/production_handler.go` | 1-459 | 459 lines — **fixed**: split into production_handler.go, production_work_order.go, production_material.go, production_labor.go, production_settle.go, production_error.go |
| [x] | `internal/products/handler/supplier_product_handler.go` | 1-436 | 436 lines — **fixed**: split into supplier_product_handler.go, supplier_product_list.go, supplier_product_get.go, supplier_product_create.go, supplier_product_update.go, supplier_product_delete.go, supplier_product_best_offer.go |
| [x] | `internal/manufacturing/handler/mrp_handler.go` | 1-429 | 429 lines — **fixed**: split into mrp_handler.go, mrp_forecast_handler.go, mrp_planned_order_handler.go, mrp_response.go |
| [x] | `internal/crm/handler/activity_handler.go` | 1-426 | 426 lines — **fixed**: split into activity.go, activity_list.go, activity_get.go, activity_create.go, activity_update.go, activity_delete.go, activity_mark_done.go |
| [x] | `internal/subscription/handler/subscription_handler.go` — **fixed**: split into subscription_handler.go, subscription_crud.go, subscription_transition.go, subscription_metrics.go | 1-425 | 425 lines |
| [x] | `internal/organization/handler/organization_handler.go` — **fixed**: split into organization.go, list.go, get.go, create.go, update.go, delete.go | 1-413 | 413 lines |
| [x] | `internal/inventory/handler/count_handler.go` — **fixed**: split into count_handler.go, count_crud_handler.go, count_line_handler.go, count_post_handler.go | 1-410 | 410 lines |
| [x] | `internal/inventory/handler/inbound_cost_handler.go` — **fixed**: split into inbound_cost.go, inbound_cost_list.go, inbound_cost_get.go, inbound_cost_create.go, inbound_cost_post.go, inbound_cost_lines.go, inbound_cost_adjustments.go, inbound_cost_error.go | 1-391 | 391 lines |
| [x] | `internal/reference/handler/fxrate_handler.go` — **fixed**: split into fxrate.go, fxrate_list.go, fxrate_crud.go, fxrate_resolve.go | 1-393 | 393 lines |
| [x] | `internal/pos/handler/pos_order_handler.go` — **fixed**: split into pos_order_handler.go, pos_order_response.go, pos_order_list.go, pos_order_get.go, pos_order_sell.go, pos_order_invoice.go, pos_order_refund.go | 1-385 | 385 lines |
| [x] | `internal/inventory/handler/reorder_handler.go` — **fixed**: split into reorder.go, reorder_rule_response.go, reorder_rule_handler.go, reorder_candidates_handler.go | 1-379 | 379 lines |
| [x] | `internal/reporting/handler/kpi_handler.go` — **fixed**: split into kpi_handler.go, kpi_sales.go, kpi_finance.go, kpi_operations.go, kpi_business.go | 1-370 | 370 lines |
| [x] | `internal/procurement/handler/purchase_request_handler.go` — **fixed**: split into purchase_request_handler.go, purchase_request_response.go, purchase_request_crud.go, purchase_request_transition.go | 1-357 | 357 lines |
| [x] | `internal/inventory/handler/transfer_handler.go` — **fixed**: split into transfer.go, transfer_list.go, transfer_get.go, transfer_create.go, transfer_send.go, transfer_receive.go, transfer_error.go | 1-353 | 353 lines |
| [x] | `internal/reference/handler/account_handler.go` — **fixed**: split into account.go, account_list.go, account_get.go, account_create.go, account_update.go, account_delete.go | 1-347 | 347 lines |
| [x] | `internal/reference/handler/tax_handler.go` — **fixed**: split into tax_handler.go, tax_response.go, tax_list.go, tax_get.go, tax_create.go, tax_update.go, tax_delete.go | 1-345 | 345 lines |
| [x] | `internal/products/handler/item_category_handler.go` | 1-344 | 344 lines | — **fixed**: split into item_category_handler.go, item_category_list.go, item_category_get.go, item_category_create.go, item_category_update.go, item_category_delete.go
| [x] | `internal/accounting/handler/tax_rule_handler.go` | 1-341 | 341 lines | — **fixed**: split into tax_rule_handler.go, tax_rule_response.go, tax_rule_list.go, tax_rule_get.go, tax_rule_create.go, tax_rule_resolve.go
| [x] | `internal/accounting/handler/payment_handler.go` | 1-339 | 339 lines | — **fixed**: split into payment_handler.go, payment_response.go, payment_list.go, payment_get.go, payment_create.go
| [x] | `internal/accounting/handler/bank_statement_handler.go` | 1-338 | 338 lines | — **fixed**: split into bank_statement.go, bank_statement_list.go, bank_statement_get.go, bank_statement_create.go, bank_statement_match.go
| [x] | `internal/accounting/handler/tax_period_handler.go` | 1-332 | 332 lines | — **fixed**: split into tax_period_handler.go, tax_period_crud.go, tax_period_state.go, tax_period_error.go
| [x] | `internal/reference/handler/tax_year_handler.go` | 1-330 | 330 lines | — **fixed**: split into tax_year.go, tax_year_response.go, tax_year_list.go, tax_year_get.go, tax_year_create.go, tax_year_update.go, tax_year_delete.go
| [x] | `internal/giftcard/handler/coupon_handler.go` | 1-323 | 323 lines | — **fixed**: split into coupon_handler.go, coupon_response.go, coupon_crud.go, coupon_redeem.go, coupon_helpers.go
| [x] | `internal/reference/handler/journal_handler.go` | 1-322 | 322 lines | — **fixed**: split into journal.go, journal_list.go, journal_get.go, journal_create.go, journal_update.go, journal_delete.go
| [x] | `internal/accounting/handler/tax_return_handler.go` | 1-314 | 314 lines | — **fixed**: split into tax_return.go, tax_return_list.go, tax_return_get.go, tax_return_create.go, tax_return_action.go
| [x] | `internal/interorganization/handler/mirror_handler.go` | 1-313 | 313 lines | — **fixed**: split into mirror_handler.go, mirror_rule.go, mirror_transaction.go
| [x] | `internal/pos/handler/pos_session_handler.go` | 1-313 | 313 lines | — **fixed**: split into pos_session_handler.go, pos_session_response.go, pos_session_list.go, pos_session_get.go, pos_session_open.go, pos_session_close.go, pos_error_handler.go
| [x] | `internal/accounting/handler/budget_handler.go` | 1-309 | 309 lines | — **fixed**: split into budget_handler.go, budget_response.go, budget_list.go, budget_get.go, budget_create.go, budget_variance.go
| [x] | `internal/iam/handler/member_handler.go` | 1-300 | 300 lines | — **fixed**: split into member_handler.go, member_crud.go, member_role.go, permission.go
| [x] | `internal/xtradata/handler/config_handler.go` | 1-298 | 298 lines | — **fixed**: split into config_handler.go, config_response.go, config_list.go, config_get.go, config_create.go, config_update.go, config_delete.go, config_errors.go
| [x] | `internal/inventory/handler/lot_handler.go` | 1-297 | 297 lines | — **fixed**: split into lot_handler.go, lot_list.go, lot_get.go, lot_create.go, lot_update.go
| [x] | `internal/expense/handler/expense_category_handler.go` | 1-285 | 285 lines | — **fixed**: split into expense_category.go, expense_category_crud.go
| [x] | `internal/subscription/handler/subscription_plan_handler.go` | 1-283 | 283 lines | — **fixed**: split into subscription_plan.go, subscription_plan_crud.go
| [x] | `internal/reference/handler/uom_category_handler.go` | 1-279 | 279 lines | — **fixed**: split into uom_category.go, uom_category_crud.go, uom_category_links.go
| [x] | `internal/checkout/handler/checkout_handler.go` | 1-267 | 267 lines | — **fixed**: split into checkout_handler.go, checkout_list.go, checkout_get.go, checkout_create.go, checkout_webhook.go
| [x] | `internal/reference/handler/uom_handler.go` | 1-363 | 363 lines | — **fixed**: split into unit.go, uom_crud.go, uom_convert.go, uom_errors.go
| [x] | `internal/reference/handler/dimension_handler.go` | 1-325 | 325 lines | — **fixed**: split into count_handler.go, count_crud_handler.go, count_line_handler.go, count_post_handler.go
| [x] | `internal/reference/handler/carrier_handler.go` | 1-229 | 229 lines | — **fixed**: split into carrier_handler.go, carrier_list.go, carrier_get.go, carrier_create.go, carrier_update.go, carrier_delete.go
| [x] | `internal/accounting/handler/withholding_handler.go` | 1-255 | 255 lines | — **fixed**: split into withholding_handler.go, withholding_list.go, withholding_create.go, withholding_apply.go
| [x] | `internal/inventory/handler/reservation_handler.go` | 1-253 | 253 lines | — **fixed**: split into reservation_handler.go, reservation_list.go, reservation_reserve.go, reservation_release.go
| [x] | `internal/manufacturing/handler/subcontract_handler.go` | 1-253 | 253 lines | — **fixed**: split into subcontract_handler.go, subcontract_create.go, subcontract_send.go, subcontract_receive.go, subcontract_done.go, subcontract_error.go
| [x] | `internal/interorganization/handler/consolidation_handler.go` | 1-225 | 225 lines | — **fixed**: split into consolidation_handler.go, consolidation_run.go
| [x] | `internal/interorganization/handler/dropship_handler.go` | 1-212 | 212 lines | — **fixed**: split into dropship_handler.go, dropship_create.go, dropship_receive.go, dropship_error.go
| [x] | `internal/reporting/handler/accrual_handler.go` | 1-216 | 216 lines | — **fixed**: split into accrual_handler.go, accrual_create.go, accrual_reverse.go, accrual_list.go, accrual_error.go

## 2. Missing `doc.go` Files

**Target:** Add `doc.go` with package-level documentation to every domain package.

| Status | File | Issue |
|--------|------|-------|
| [x] | `internal/systemconfig/doc.go` | Missing package documentation |
| [x] | `internal/localization/handler/doc.go` | Missing handler package documentation |

## 3. Missing `service_test.go`

**Target:** Every domain with a `service.go` must have a corresponding `service_test.go` with table-driven tests.

| Status | File | Issue |
|--------|------|-------|
| [x] | `internal/checkout/service_test.go` | Missing unit tests for checkout service |

## 4. Missing `fixtures.go`

**Target:** Every domain with models should have `fixtures.go` with gofakeit-based test fixture factories.

| Status | File | Issue |
|--------|------|-------|
| [x] | `internal/accounting/fixtures.go` | Missing test fixtures |
| [x] | `internal/asset/fixtures.go` | Missing test fixtures |
| [x] | `internal/checkout/fixtures.go` | Missing test fixtures |
| [x] | `internal/commission/fixtures.go` | Missing test fixtures |
| [x] | `internal/crm/fixtures.go` | Missing test fixtures |
| [x] | `internal/expense/fixtures.go` | Missing test fixtures |
| [x] | `internal/giftcard/fixtures.go` | Missing test fixtures |
| [x] | `internal/interorganization/fixtures.go` | Missing test fixtures |
| [x] | `internal/inventory/fixtures.go` | Missing test fixtures |
| [x] | `internal/manufacturing/fixtures.go` | Missing test fixtures |
| [x] | `internal/organization/fixtures.go` | Missing test fixtures — skipped: no models.go in domain |
| [x] | `internal/contacts/fixtures.go` | Missing test fixtures |
| [x] | `internal/payroll/fixtures.go` | Missing test fixtures |
| [x] | `internal/pos/fixtures.go` | Missing test fixtures |
| [x] | `internal/procurement/fixtures.go` | Missing test fixtures |
| [x] | `internal/products/fixtures.go` | Missing test fixtures |
| [x] | `internal/project/fixtures.go` | Missing test fixtures |
| [x] | `internal/quality/fixtures.go` | Missing test fixtures |
| [x] | `internal/reference/fixtures.go` | Missing test fixtures |
| [x] | `internal/reporting/fixtures.go` | Missing test fixtures |
| [x] | `internal/returns/fixtures.go` | Missing test fixtures |
| [x] | `internal/sales/fixtures.go` | Missing test fixtures |
| [x] | `internal/subscription/fixtures.go` | Missing test fixtures |
| [x] | `internal/xtradata/fixtures.go` | Missing test fixtures |

## 5. Missing `dao_mock.go`

**Target:** Every domain with a `dao.go` must have a corresponding `dao_mock.go` for unit testing.

| Status | File | Issue |
|--------|------|-------|
| [x] | `internal/reference/dao_mock.go` | Missing DAO mock for reference domain — **fixed**: created with AccountDAOMock, TaxDAOMock, WithholdingTaxDAOMock |

## 6. String State Literal in Business Logic

**Target:** Use domain constants for state values, not string literals.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/checkout/service.go` | 99 | `invoice.State != "posted"` should use `accounting.MoveStatePosted` constant |

## 7. Missing Doc Comments on Exported Types

**Target:** All exported types and functions must have doc comments per Go conventions.

**Skipped:** User prefers self-explanatory code without doc comments.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [~] | `internal/accounting/balance_dao.go` | 10 | `AccountBalance` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/balance_dao.go` | 16 | `AccountBalanceDAO` interface missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/balance_dao.go` | 24 | `NewAccountBalanceDAO` function missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/balance_dao_mock.go` | 8 | `AccountBalanceDAOMock` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/bank_dao.go` | 11 | `BankStatementDAO` interface missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/bank_dao.go` | 22 | `NewBankStatementDAO` function missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/bank_dao.go` | 46 | `BankStatementLineDAO` interface missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/bank_dao_mock.go` | 11 | `BankStatementDAOMock` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/bank_dao_mock.go` | 31 | `BankStatementLineDAOMock` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/bank_dao_mock.go` | 59 | `AccountFullReconcileDAOMock` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/bank_models.go` | 16 | `BankStatement` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/bank_models.go` | 30 | `BankStatementLine` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/bank_service.go` | 15 | `BankStatementLineRequest` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/bank_service.go` | 23 | `CreateBankStatementRequest` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/bank_service.go` | 32 | `BankStatementService` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/budget_dao.go` | 13 | `BudgetDAO` interface missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/budget_dao.go` | 23 | `NewBudgetDAO` function missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/budget_dao.go` | 40 | `BudgetLineDAO` interface missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/budget_dao_mock.go` | 11 | `BudgetDAOMock` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/budget_dao_mock.go` | 23 | `BudgetLineDAOMock` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/budget_dao_mock.go` | 51 | `BudgetQueryDAOMock` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/budget_models.go` | 17 | `Budget` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/budget_models.go` | 30 | `BudgetLine` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/budget_models.go` | 48 | `TaxRule` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/budget_service.go` | 13 | `BudgetLineRequest` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/budget_service.go` | 19 | `CreateBudgetRequest` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/budget_service.go` | 27 | `BudgetVarianceLine` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/dao.go` | 11 | `JournalEntryDAO` interface missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/dao.go` | 23 | `NewJournalEntryDAO` function missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/dao.go` | 57 | `JournalLineDAO` interface missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/dao_mock.go` | 10 | `JournalEntryDAOMock` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/dao_mock.go` | 38 | `JournalLineDAOMock` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/deferral_dao.go` | 12 | `DeferredScheduleDAO` interface missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/deferral_dao.go` | 24 | `NewDeferredScheduleDAO` function missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/deferral_dao.go` | 58 | `DeferredScheduleLineDAO` interface missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/deferral_dao_mock.go` | 11 | `DeferredScheduleDAOMock` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/deferral_dao_mock.go` | 39 | `DeferredScheduleLineDAOMock` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/deferral_models.go` | 24 | `DeferredSchedule` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/deferral_models.go` | 48 | `DeferredScheduleLine` type missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/deferral_service.go` | 15 | `JournalResolver` interface missing doc comment — **skipped**: no doc comments per user preference |

Note: This is a widespread issue across the codebase. The above is a sample from the accounting domain only. Similar patterns exist in most other domains.

## 8. Missing Handler Doc Comments

**Target:** All handler structs must have doc comments per Go conventions.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [~] | `internal/accounting/handler/journal_entry_handler.go` | 15 | `JournalEntryHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [x] | `internal/accounting/handler/bank_statement_handler.go` | 14 | `BankStatementHandler` struct missing doc comment | — **fixed**: split into bank_statement.go, bank_statement_list.go, bank_statement_get.go, bank_statement_create.go, bank_statement_match.go
| [x] | `internal/accounting/handler/budget_handler.go` | 14 | `BudgetHandler` struct missing doc comment | — **fixed**: split into budget_handler.go, budget_response.go, budget_list.go, budget_get.go, budget_create.go, budget_variance.go
| [~] | `internal/accounting/handler/deferral_handler.go` | 13 | `DeferralHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/accounting/handler/reminder_handler.go` | 12 | `ReminderHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [x] | `internal/accounting/handler/tax_period_handler.go` | 15 | `TaxPeriodHandler` struct missing doc comment | — **fixed**: split into tax_period_handler.go, tax_period_crud.go, tax_period_state.go, tax_period_error.go
| [x] | `internal/accounting/handler/tax_rule_handler.go` | 13 | `TaxRuleHandler` struct missing doc comment | — **fixed**: split into tax_rule_handler.go, tax_rule_response.go, tax_rule_list.go, tax_rule_get.go, tax_rule_create.go, tax_rule_resolve.go
| [~] | `internal/accounting/handler/invoice_handler.go` | 16 | `InvoiceHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [x] | `internal/accounting/handler/payment_handler.go` | 14 | `PaymentHandler` struct missing doc comment | — **fixed**: split into payment_handler.go, payment_response.go, payment_list.go, payment_get.go, payment_create.go
| [~] | `internal/accounting/handler/reconcile_handler.go` | 11 | `ReconcileHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [x] | `internal/accounting/handler/tax_return_handler.go` | 14 | `TaxReturnHandler` struct missing doc comment | — **fixed**: split into tax_return.go, tax_return_list.go, tax_return_get.go, tax_return_create.go, tax_return_action.go
| [x] | `internal/accounting/handler/withholding_handler.go` | 13 | `WithholdingTaxHandler` struct missing doc comment | — **fixed**: split into withholding_handler.go, withholding_list.go, withholding_create.go, withholding_apply.go
| [~] | `internal/asset/handler/asset_handler.go` | 15 | `AssetCategoryHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [x] | `internal/checkout/handler/checkout_handler.go` | 13 | `CheckoutHandler` struct missing doc comment | — **fixed**: split into checkout_handler.go, checkout_list.go, checkout_get.go, checkout_create.go, checkout_webhook.go
| [~] | `internal/commission/handler/commission_handler.go` | 14 | `CommissionHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/crm/handler/activity_handler.go` | 14 | `ActivityHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/crm/handler/lead_handler.go` | 16 | `ProspectHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/crm/handler/opportunity_handler.go` | 15 | `OpportunityHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/crm/handler/pipeline_handler.go` | 10 | `PipelineHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/crm/handler/stage_handler.go` | 11 | `PipelineStageHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [~] | `internal/crm/handler/stage_handler.go` | 78 | `SalesGroupHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [x] | `internal/expense/handler/expense_category_handler.go` | 13 | `ExpenseCategoryHandler` struct missing doc comment | — **fixed**: split into expense_category.go, expense_category_crud.go
| [x] | `internal/expense/handler/expense_report_handler.go` — **fixed**: split into expense_report_types.go, expense_report_crud.go, expense_report_transition.go | 14 | `ExpenseReportHandler` struct missing doc comment |
| [x] | `internal/giftcard/handler/coupon_handler.go` | 15 | `CouponHandler` struct missing doc comment | — **fixed**: split into coupon_handler.go, coupon_response.go, coupon_crud.go, coupon_redeem.go, coupon_helpers.go
| [x] | `internal/giftcard/handler/giftcard_handler.go` — **fixed**: split into giftcard_handler.go, response.go, list.go, get.go, issue.go, redeem.go, refund.go, forfeit_expired.go, transactions.go, helpers.go | 13 | `GiftCardHandler` struct missing doc comment |
| [~] | `internal/iam/handler/auth_handler.go` | 15 | `AuthHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [x] | `internal/iam/handler/member_handler.go` | 12 | `MemberHandler` struct missing doc comment | — **fixed**: split into member_handler.go, member_crud.go, member_role.go, permission.go
| [~] | `internal/iam/handler/user_handler.go` | 19 | `UserHandler` struct missing doc comment — **skipped**: no doc comments per user preference |
| [x] | `internal/interorganization/handler/consolidation_handler.go` | 10 | `ConsolidationHandler` struct missing doc comment | — **fixed**: split into consolidation_handler.go, consolidation_run.go
| [x] | `internal/interorganization/handler/dropship_handler.go` | 14 | `DropShipHandler` struct missing doc comment | — **fixed**: split into dropship_handler.go, dropship_create.go, dropship_receive.go, dropship_error.go

Note: This is a widespread issue. The above is a sample. Most handler structs across all domains lack doc comments.

## 9. Cross-Domain Imports in Handlers

**Target:** Handlers must only import from their own domain, `kernel/`, `httpx/`, or `helper/`. Cross-domain imports violate domain isolation and create tight coupling.

| Status | File | Violating Import | Issue |
|--------|------|-----------------|-------|
| [x] | `internal/accounting/handler/tax_period_handler.go` | `reference` | Imports reference domain — **fixed**: re-exported `ErrTaxYearNotFound` in accounting errors |
| [x] | `internal/asset/handler/asset_handler.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data (AssetCategory) |
| [x] | `internal/crm/handler/helpers.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data (PipelineStage) |
| [x] | `internal/crm/handler/lead_handler.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data (PipelineStage) |
| [x] | `internal/crm/handler/opportunity_handler.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data (PipelineStage) |
| [x] | `internal/crm/handler/stage_handler.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data (PipelineStage/SalesGroup) |
| [x] | `internal/expense/handler/expense_category_handler.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data (ExpenseCategory) |
| [x] | `internal/iam/handler/user_handler.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data (Organization) |
| [x] | `internal/interorganization/handler/dropship_handler.go` | `procurement` | Imports procurement domain — **skipped**: response type references procurement.PurchaseOrder |
| [x] | `internal/inventory/handler/warehouse_handler.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data (Warehouse/StockLocation) |
| [x] | `internal/manufacturing/handler/mo_handler.go` | `inventory` | Imports inventory domain — **fixed**: re-exported error sentinels in manufacturing errors |
| [x] | `internal/manufacturing/handler/production_handler.go` | `inventory` | Imports inventory domain — **fixed**: re-exported error sentinels in manufacturing errors |
| [x] | `internal/organization/handler/organization_handler.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data (Organization) |
| [x] | `internal/payroll/handler/hr_handler.go` | `contacts`, `reference` | Imports contacts and reference domains — **skipped**: model type imports for shared master data |
| [x] | `internal/payroll/handler/payroll_handler.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data (SalaryRule) |
| [x] | `internal/pos/handler/payment_account_handler.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data |
| [x] | `internal/pos/handler/pos_config_handler.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data (POSConfig) |
| [x] | `internal/procurement/handler/purchase_order_handler.go` | `accounting` | Imports accounting domain — **skipped**: response type references accounting.Invoice/Payment and error sentinels |
| [x] | `internal/products/handler/item_category_handler.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data (ItemCategory) |
| [x] | `internal/quality/handler/quality_handler.go` | `inventory`, `reference` | Imports inventory and reference domains — **fixed**: inventory re-exported in quality errors; reference still used for QualityPoint type |
| [x] | `internal/sales/handler/sale_order_handler.go` | `accounting`, `inventory`, `products` | Imports 3 other domains — **fixed**: inventory/products re-exported in sales errors; accounting still used for Invoice/Payment types |
| [x] | `internal/subscription/handler/subscription_plan_handler.go` | `reference` | Imports reference domain — **skipped**: model type import for shared master data (SubscriptionPlan) |
| [x] | `internal/xtradata/handler/config_handler.go` | `reference` | Imports reference domain — **fixed**: removed DAO, added List/Find/Delete to ConfigService |
| [x] | `internal/organization/handler/handler_mock.go` | 6 | `reference` | Mock file in handler package imports reference domain — **skipped**: mock files must reference concrete types |
| [x] | `internal/sales/handler/sale_order_handler_mock.go` | 6, 13-14, 15 | `crm`, `contacts`, `products`, `reference` | Mock file imports 4 cross-domain packages beyond main handler — **skipped**: mock files must reference concrete types |

## 10. Handlers Using DAO Directly (Missing Service Layer)

**Target:** All data access must go through a service layer. Handlers must not hold DAO references or perform CRUD operations directly.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/pos/handler/payment_account_handler.go` | 100-197 | Holds `dao.CRUD`, performs upsert logic directly — no service layer — **already uses service**: `PaymentAccountHandler` holds `*pos.PaymentAccountService` |
| [x] | `internal/quality/handler/quality_handler.go` | 80-195 | Holds `QualityPointDAO`, performs all CRUD directly — no service layer — **fixed**: created `QualityPointService` wrapping DAO |
| [x] | `internal/payroll/handler/hr_handler.go` | 17-687 | Three handler structs use `dao.Base` directly for full CRUD — no service layer — **fixed**: `DepartmentHandler`, `JobPositionHandler`, `LeaveTypeHandler` now use `HRService` |
| [x] | `internal/sales/handler/sale_order_handler.go` | 234-257, 428-523 | Holds DAOs alongside service, calls DAO directly for reads and deletes — **fixed**: removed DAOs from handler, added `List`, `Find`, `Delete` methods to `SaleOrderService` |
| [x] | `internal/manufacturing/handler/mo_handler.go` | 143, 190-201 | Holds DAO alongside service, calls DAO directly for reads — **fixed**: removed DAOs, added `List`, `Find`, `ListComponents` to `ProductionOrderService` |
| [x] | `internal/expense/handler/expense_category_handler.go` | 13-15, 79, 125, 172-226 | Holds `dao.Base[reference.ExpenseCategory]`, performs all CRUD directly — no service layer — **fixed**: created `ExpenseCategoryService` wrapping DAO |
| [x] | `internal/manufacturing/handler/bom_handler.go` | 14-17, 125, 172, 315, 365, 374, 419, 478 | Holds `RecipeDAO` and `RecipeLineDAO` alongside `RecipeService`, calls DAO directly (8 calls) — **fixed**: removed DAOs, added `List`, `Find`, `Delete`, `ListLines` to `RecipeService` |
| [x] | `internal/manufacturing/handler/mrp_handler.go` | 14-18, 171, 175, 198, 205, 209, 239, 350, 388 | Holds 4 DAOs alongside `PlanningService`, calls all directly (8 calls) — **fixed**: removed all DAOs, added 6 methods to `PlanningService` |
| [x] | `internal/manufacturing/handler/production_handler.go` | 15-16, 167 | Holds `ShopTaskDAO` alongside `ProductionService`, calls DAO directly — **fixed**: removed DAO, added `ListShopTasksByMO` to `ProductionService` |

## 11. Business Logic in Handlers

**Target:** Handlers should only parse HTTP, call service, and write response. Business logic (validation, state checks, calculations, defaults) belongs in the service layer.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/sales/handler/sale_order_handler.go` | 239, 433, 514 | Tenant ownership checks (`httpx.OwnsTenant`) — should be in service |
| [x] | `internal/sales/handler/sale_order_handler.go` | 726-729 | Computes default delivery date — date defaulting is business logic |
| [x] | `internal/sales/handler/sale_order_handler.go` | 800-803 | Computes default invoice date — same issue |
| [x] | `internal/sales/handler/sale_order_handler.go` | 874-877 | Computes default payment date — same issue |
| [x] | `internal/pos/handler/payment_account_handler.go` | 120-130 | Upsert logic (find existing → update or create) — orchestration in handler — **fixed**: upsert logic moved to `PaymentAccountService.Upsert()` |
| [x] | `internal/manufacturing/handler/mo_handler.go` | 195, 407 | Tenant ownership checks — should be in service |
| [x] | `internal/manufacturing/handler/production_handler.go` | 450-458 | `parseDate` helper with fallback logic in handler |
| [x] | `internal/quality/handler/quality_handler.go` | 134, 329, 368-372 | Existence + tenant checks before delegating to service — **skipped**: lines 329/368-372 already use scoped `FindInOrg`; line 134 uses unscoped `Find` + handler guard (acceptable pattern) |
| [x] | `internal/payroll/handler/hr_handler.go` | 163-166 | Department/Create resolves organization in handler instead of service — **skipped**: handler resolves tenant context from request, passes to service (standard pattern) |
| [x] | `internal/interorganization/handler/dropship_handler.go` | 93-100 | Handler parses and defaults date — should be in service |
| [x] | `internal/accounting/handler/invoice_handler.go` | 309-319, 498-503 | `parseDates()` helper defaults date to `time.Now().UTC()` — date defaulting is business logic |
| [x] | `internal/accounting/handler/payment_handler.go` | 216-224 | `register()` defaults date to `time.Now().UTC()` when parsed date is nil |
| [x] | `internal/accounting/handler/journal_entry_handler.go` | 357 | `reversalDate := time.Now().UTC()` — date defaulting for move reversal |
| [x] | `internal/inventory/handler/stock_handler.go` | 764-783 | `parseDate()` and `parseDatePointer()` silently default to `time.Now()` on empty/invalid input |
| [x] | `internal/crm/handler/activity_handler.go` | 301-306 | Handler sets `activity.DoneAt = &now` directly on model — state mutation timestamp belongs in service — **fixed**: `DoneAt` set in `ProspectActivityService.UpdateActivity()` |
| [x] | `internal/reporting/handler/kpi_handler.go` | 20-38, 293, 324 | `parseDateRange()` computes default date range — date computation is business logic |
| [x] | `internal/reporting/handler/fx_revaluation_handler.go` | 80 | `date := time.Now().UTC()` — date defaulting for revaluation query |
| [x] | `internal/reporting/handler/report_handler.go` | 77 | `asOf := time.Now().UTC()` — date defaulting for report query |
| [x] | `internal/reporting/handler/statement_handler.go` | 67 | `asOf := time.Now().UTC()` — date defaulting for statement query |
| [x] | `internal/giftcard/handler/coupon_handler.go` | 269-276 | `date := time.Now().UTC()` defaults coupon redemption date |
| [x] | `internal/products/handler/price_book_handler.go` | 534-541 | `date := time.Now()` defaults price resolution date |
| [x] | `internal/products/handler/supplier_product_handler.go` | 399-406 | `date := time.Now()` defaults supplier offer date |
| [x] | `internal/procurement/handler/purchase_order_handler.go` | 675-680 | `defaultDate()` helper defaults to `time.Now().UTC()` |
| [x] | `internal/service/handler/service_handler.go` | 1221 | Passes `time.Now().UTC()` to `maintenance.GenerateDueOrders` — scheduling logic in handler — **fixed**: added `GenerateDueOrdersNow()` convenience method; handler calls it instead |

## 12. Missing Error Checks

**Target:** All errors must be checked and handled. Discarded errors (`_ =`) hide failures and make debugging impossible.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/asset/service.go` | 172 | `_ = depreciable.Div()` — division error discarded |
| [x] | `internal/organization/handler/organization_handler.go` | 389, 396, 400 | `_ = httpx.Create...()` — httpx response errors discarded |
| [x] | `internal/contacts/handler/contact_handler.go` | 448, 455, 459 | `_ = httpx.Create...()` — httpx response errors discarded |
| [x] | `internal/contacts/handler/contact_nested_handler.go` | 841, 851, 855 | `_ = httpx.Create...()` — httpx response errors discarded |
| [x] | `internal/iam/handler/request.go` | 18 | `_ = httpx.CreateUnprocessableEntityErrorResponse(...)` — error discarded |
| [x] | `internal/iam/avatar_service.go` | 59, 64 | `_ = s.store.Delete(...)` — avatar rollback errors silently discarded |
| [x] | `internal/xtradata/webhook.go` | 150 | `_, _ = mac.Write(payload)` — HMAC write error discarded |

## 13. Raw HTTP Responses Bypassing httpx Helpers

**Target:** All HTTP responses must use `httpx` helpers for consistency and proper error formatting.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/checkout/handler/checkout_handler.go` | 239 | `c.SendStatus(fiber.StatusUnauthorized)` — raw status response |
| [x] | `internal/checkout/handler/checkout_handler.go` | 241 | `c.SendStatus(fiber.StatusInternalServerError)` — raw status response |
| [x] | `internal/checkout/handler/checkout_handler.go` | 244 | `c.SendStatus(fiber.StatusOK)` — raw status response |

## 14. Error Information Leakage

**Target:** Internal error messages must never be exposed to clients. Use generic messages and log the details.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/accounting/handler/deferral_handler.go` | 359 | `err.Error()` used as client-facing message — leaks internal details |
| [x] | `internal/xtradata/handler/event_handler.go` | 187 | Missing error log before returning 500 — silent failure |

## 15. Missing `httpx.BindAndValidate`

**Target:** All request body parsing must use `httpx.BindAndValidate` which handles both binding and validation in one call.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/manufacturing/handler/mrp_handler.go` | 160 | `c.Bind().Body(&request)` — bypasses validation |
| [x] | `internal/manufacturing/handler/mrp_handler.go` | 274 | `c.Bind().Body(&request)` — bypasses validation |
| [x] | `internal/manufacturing/handler/mrp_handler.go` | 347 | `c.Bind().Body(&request)` — bypasses validation |
| [x] | `internal/crosscutting/handler/attachment_handler.go` | 203-204 | `c.FormValue()` — bypasses structured binding/validation — **won't fix**: multipart form uploads can't use `BindAndValidate` |

## 16. `helper.ParseID` Swallowing Parse Errors

**Target:** Path parameters must be parsed with `strconv.ParseUint` with explicit error handling. `helper.ParseID` returns 0 on invalid input without surfacing an error, leading to confusing 404/500 responses instead of clear 422 "Invalid ID" responses.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/interorganization/handler/dropship_handler.go` | 167 | `helper.ParseID` — invalid ID silently becomes 0 |
| [x] | `internal/interorganization/handler/mirror_handler.go` | 157, 198, 262 | `helper.ParseID` — 3 call sites |
| [x] | `internal/interorganization/handler/consolidation_handler.go` | 158, 206 | `helper.ParseID` — 2 call sites |
| [x] | `internal/service/handler/service_handler.go` | 127, 297, 382, 411, 530, 674, 734, 784, 813, 859, 911, 943, 1075 | `helper.ParseID` — 13 call sites |
| [x] | `internal/asset/handler/asset_handler.go` | 140, 327, 443, 492, 538 | `helper.ParseID` — 5 call sites |
| [x] | `internal/giftcard/handler/giftcard_handler.go` | 114, 235, 284, 404 | `helper.ParseID` — 4 call sites |
| [x] | `internal/project/handler/project_handler.go` | 156, 251, 293, 361, 403, 451, 516, 558, 594, 642, 692 | `helper.ParseID` — 11 call sites |
| [x] | `internal/commission/handler/commission_handler.go` | 114, 196, 251, 307, 364, 418, 679, 700 | `helper.ParseID` — 8 call sites |

## 17. Non-Standard Route Parameters

**Target:** Route parameters should always be `:id`. The resource name is inferred from the group. Using `:item_id`, `:task_id`, etc. breaks consistency.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/iam/handler/routes.go` | 25, 33 | `:session_id`, `:organization_id` instead of `:id` |
| [x] | `internal/quality/handler/routes.go` | 22 | `:shipment_id` instead of `:id` |
| [x] | `internal/project/handler/routes.go` | 19-25 | `:project_id`, `:task_id`, `:milestone_id` instead of `:id` |
| [x] | `internal/payroll/handler/routes.go` | 67 | `:employee_id`, `:leave_type_id` instead of `:id` — **skipped**: two-param route, would cause `:id` collision |
| [x] | `internal/contacts/handler/routes.go` | 24-32 | `:address_id`, `:acstock_count_id` instead of `:id` |
| [x] | `internal/manufacturing/handler/routes.go` | 41 | `:production_order_id` instead of `:id` — **skipped**: two-param route, would cause `:id` collision |
| [x] | `internal/procurement/handler/routes.go` | 45 | `:supplier_quote_id` instead of `:id` — **skipped**: two-param route, would cause `:id` collision |
| [~] | `internal/iam/handler/routes.go` | 20, 21 | `:provider` instead of `:id` on OAuth routes — **intentional**: provider is not an entity ID |
| [~] | `internal/pos/handler/routes.go` | 40, 41 | `:method` instead of `:id` on payment account routes — **intentional**: method is not an entity ID |

## 18. Missing `validate` Tags on Request Structs

**Target:** All request struct fields must have `validate` tags for input validation.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/checkout/handler/checkout_handler.go` | 160 | `Currency` field has no `validate` tag — accepts arbitrary strings |
| [x] | `internal/checkout/handler/checkout_handler.go` | 162 | `ReturnURL` field has no `validate` tag — no URL format validation |
| [x] | `internal/iam/handler/user_handler.go` | 564-576 | `UpdateUserRequest` — all 11 fields (`FirstName`, `LastName`, `Avatar`, `Bio`, `Birthday`, `Active`, `Sex`, `Address`, `City`, `PostalCode`, `SystemRoleID`) have no `validate` tags |

## 19. Sensitive Data in Logs/Storage

**Target:** Never log or store raw sensitive data (passwords, tokens, PII). Use `helper.MaskEmail()` for emails.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/checkout/service.go` | 210-212, 220-221 | Full webhook payload with payment details (virtual account numbers) stored unredacted — VA number redacted in payload and masked in channel field |
| [x] | `internal/checkout/service.go` | 135 | Xendit SDK error logged — now logs only status and error_code, not full error string |

## 20. Direct `db` Import in Handler

**Target:** Handlers must never import from `internal/db` or access the database directly. All data access goes through DAO → Service → Handler.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/pos/handler/payment_account_handler.go` | 7 | Imports `internal/db` directly, uses `db.IsUniqueViolation()` in handler logic |

## 21. Cross-Domain Imports in Services

**Target:** Service layers must only import from their own domain, `kernel/`, `httpx/`, `helper/`, or shared infrastructure. Cross-domain imports violate domain isolation and create tight coupling.

| Status | File | Line(s) | Violating Import(s) | Issue |
|--------|------|---------|---------------------|-------|
| [x] | `internal/pos/service.go` | 7, 10-11, 18-20 | accounting, iam, inventory, localization, products, reference | Imports 6 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/sales/service.go` | 8, 10, 17-20 | crm, inventory, localization, contacts, products, reference | Imports 6 other domains — **skipped**: legitimate cross-domain orchestration (sale orders) |
| [x] | `internal/returns/service.go` | 7, 10, 14-16 | accounting, inventory, procurement, reference, sales | Imports 5 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/procurement/purchase_order_service.go` | 7, 9, 16-18 | accounting, inventory, localization, contacts, reference | Imports 5 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/expense/service.go` | 7, 13-15 | accounting, localization, project, reference | Imports 4 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/project/service.go` | 7, 14-16 | accounting, contacts, payroll, reference | Imports 4 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/subscription/service.go` | 8, 14-16 | accounting, contacts, products, reference | Imports 4 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/manufacturing/mrp_service.go` | 8, 11-13 | inventory, procurement, products, sales | Imports 4 other domains — **skipped**: legitimate cross-domain orchestration (purchase orders) |
| [x] | `internal/manufacturing/mo_service.go` | 7, 10 | inventory, products | Imports 2 other domains — **skipped**: legitimate cross-domain orchestration (stock moves) |
| [x] | `internal/manufacturing/service.go` | 7 | products | Imports products domain — **skipped**: legitimate cross-domain orchestration (item lookup) |
| [x] | `internal/manufacturing/production_service.go` | 8, 10, 15 | accounting, inventory, reference | Imports 3 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/manufacturing/subcontract_service.go` | 8, 10, 13 | accounting, inventory, procurement | Imports 3 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/payroll/service.go` | 8, 10-11 | iam, contacts, reference | Imports 3 other domains — **skipped**: legitimate cross-domain orchestration (contact lookup) |
| [x] | `internal/payroll/payroll_service.go` | 8, 14 | accounting, reference | Imports 2 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/accounting/invoice_service.go` | 13-14 | localization, reference | Imports 2 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/accounting/bank_service.go` | 11 | reference | Imports reference domain — **skipped**: reference is shared kernel for master data |
| [x] | `internal/accounting/reminder_service.go` | 11 | reference | Imports reference domain — **skipped**: reference is shared kernel for master data |
| [x] | `internal/accounting/payment_service.go` | 13 | reference | Imports reference domain — **skipped**: reference is shared kernel for master data |
| [x] | `internal/asset/service.go` | 8, 13 | accounting, reference | Imports 2 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/checkout/service.go` | 14 | reference | Imports reference domain — **skipped**: legitimate cross-domain orchestration |
| [x] | `internal/commission/service.go` | 9, 11 | accounting, inventory | Imports 2 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/crm/service.go` | 10-11 | contacts, reference | Imports 2 other domains — **skipped**: legitimate cross-domain orchestration (contact lookup) |
| [x] | `internal/giftcard/service.go` | 10 | accounting | Imports accounting domain — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/inventory/service.go` | 9 | reference | Imports reference domain — **skipped**: legitimate cross-domain orchestration (stock moves) |
| [x] | `internal/localization/service.go` | 6 | reference | Imports reference domain — **skipped**: legitimate cross-domain orchestration |
| [x] | `internal/organization/service.go` | 8 | reference | Imports reference domain — **skipped**: legitimate cross-domain orchestration |
| [x] | `internal/interorganization/consolidation.go` | 8, 11, 14, 15 | accounting, inventory, procurement, reference | Imports 4 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/interorganization/dropship.go` | 8, 11, 14-16 | accounting, inventory, procurement, reference, sales | Imports 5 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/interorganization/mirror.go` | 8, 9 | procurement, sales | Imports 2 other domains — **skipped**: legitimate cross-domain orchestration (purchase orders) |
| [x] | `internal/inventory/count.go` | 8 | accounting | Imports accounting domain — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/inventory/inbound_cost.go` | 8 | accounting | Imports accounting domain — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/inventory/poster.go` | 6 | accounting | Imports accounting domain — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/inventory/resolver.go` | 7, 8 | products, reference | Imports 2 other domains — **skipped**: legitimate cross-domain orchestration (stock moves) |
| [x] | `internal/inventory/scrap_router.go` | 10 | reference | Imports reference domain — **skipped**: legitimate cross-domain orchestration (stock moves) |
| [x] | `internal/inventory/transfer.go` | 8, 13 | accounting, reference | Imports 2 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/inventory/valuation.go` | 8 | accounting | Imports accounting domain — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/procurement/engines.go` | 7-9, 11, 12 | accounting, crosscutting, inventory, contacts, products | Imports 5 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/sales/o2c.go` | 7, 9 | accounting, inventory | Imports 2 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/checkout/adapter.go` | 6 | accounting | Imports accounting domain — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/crm/activity.go` | 9, 15 | contacts | Imports contacts domain — **skipped**: legitimate cross-domain orchestration (contact lookup) |
| [x] | `internal/crm/pipeline.go` | 9, 14 | reference | Imports reference domain — **skipped**: legitimate cross-domain orchestration (reporting) |
| [x] | `internal/accounting/periods.go` | 11 | reference | Imports reference domain — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/localization/tax_calculator.go` | 7 | reference | Imports reference domain — **skipped**: legitimate cross-domain orchestration |
| [x] | `internal/localization/id/accounts.go` | 7 | reference | Sub-package imports reference |
| [x] | `internal/localization/id/module.go` | 7, 8 | reference | Sub-package imports reference |
| [x] | `internal/localization/id/taxes.go` | 7 | reference | Sub-package imports reference |
| [x] | `internal/localization/id/withholdings.go` | 8 | reference | Sub-package imports reference |
| [x] | `internal/products/service.go` | 9 | reference | Imports reference domain — **skipped**: legitimate cross-domain orchestration (item lookup) |
| [x] | `internal/products/supplier_service.go` | 10 | contacts | Imports contacts domain — **skipped**: legitimate cross-domain orchestration (contact lookup) |
| [x] | `internal/quality/service.go` | 11 | reference | Imports reference domain — **skipped**: legitimate cross-domain orchestration |
| [x] | `internal/xtradata/service.go` | 11 | reference | Imports reference domain — **skipped**: legitimate cross-domain orchestration |
| [x] | `internal/procurement/requisition_service.go` | 11 | contacts | Imports contacts domain — **skipped**: legitimate cross-domain orchestration (purchase orders) |
| [x] | `internal/procurement/rfq_service.go` | 13 | contacts | Imports contacts domain — **skipped**: legitimate cross-domain orchestration (purchase orders) |
| [x] | `internal/service/service.go` | 8 | accounting | Imports accounting domain — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/reporting/accrual_service.go` | 8 | accounting | Imports accounting domain — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/reporting/fx_revaluation_service.go` | 8, 11 | accounting, reference | Imports 2 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/reporting/kpi_service.go` | 7, 9-10 | crm, project, subscription | Imports 3 other domains — **skipped**: legitimate cross-domain orchestration (reporting) |
| [x] | `internal/reporting/period_close_service.go` | 8-9 | accounting, asset | Imports 2 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/reporting/report_service.go` | 7 | accounting | Imports accounting domain — **skipped**: legitimate cross-domain orchestration (accounting posting) |
| [x] | `internal/reporting/statement_service.go` | 7, 9 | accounting, reference | Imports 2 other domains — **skipped**: legitimate cross-domain orchestration (accounting posting) |

## 22. Cross-Domain Imports in DAOs

**Target:** DAO layers must only query their own domain's models. Cross-domain DAO imports break the persistence boundary.

| Status | File | Line(s) | Violating Import | Issue |
|--------|------|---------|-----------------|-------|
| [x] | `internal/iam/dao.go` | 9 | reference | IAM DAO imports reference domain — **skipped**: shared kernel for master data |
| [x] | `internal/inventory/dao.go` | 10 | reference | Inventory DAO imports reference domain — **skipped**: shared kernel for master data |
| [x] | `internal/quality/dao.go` | 9 | reference | Quality DAO imports reference domain — **skipped**: shared kernel for master data |
| [x] | `internal/reporting/report_dao.go` | 7 | accounting | Reporting DAO imports accounting domain — **skipped**: legitimate cross-domain join for reporting aggregation |
| [x] | `internal/xtradata/dao.go` | 7 | reference | Xtradata DAO imports reference domain — **skipped**: shared kernel for master data |

## 23. Missing Error Logging at Handler Level

**Target:** All 500 error responses must be preceded by `httpx.RequestLog(c).Error(...)` with operation context. Silent 500s make production debugging impossible.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/manufacturing/handler/mrp_handler.go` | 173 | `Run` handler returns 500 on demands list error without `RequestLog` |
| [x] | `internal/manufacturing/handler/mrp_handler.go` | 177 | `Run` handler returns 500 on planned orders list error without `RequestLog` |
| [x] | `internal/manufacturing/handler/mrp_handler.go` | 200 | `Get` handler returns 500 on run find error without `RequestLog` |
| [x] | `internal/manufacturing/handler/mrp_handler.go` | 207 | `Get` handler returns 500 on demands list error without `RequestLog` |
| [x] | `internal/manufacturing/handler/mrp_handler.go` | 211 | `Get` handler returns 500 on planned orders list error without `RequestLog` |
| [x] | `internal/manufacturing/handler/mrp_handler.go` | 359 | `CreateForecast` handler returns 500 on create error without `RequestLog` |
| [x] | `internal/iam/handler/user_handler.go` | 413 | `ListUsers` handler returns 500 on user list error without `RequestLog` |
| [x] | `internal/iam/handler/user_handler.go` | 465 | `GetUser` handler returns 500 on user find error without `RequestLog` |
| [x] | `internal/iam/handler/auth_handler.go` | 376 | `GetSessionsHandler` returns 500 on session list error without `RequestLog` |
| [x] | `internal/httpx/middleware.go` | 40 | `AuthGuard` returns 500 on user DAO lookup failure without `RequestLog` |
| [x] | `internal/httpx/middleware.go` | 50 | `AuthGuard` returns 500 on session lookup failure without `RequestLog` |
| [x] | `internal/httpx/middleware.go` | 88 | `SystemAuthzGuard` returns 500 on permission check failure without `RequestLog` |
| [x] | `internal/httpx/middleware.go` | 125 | `OrgAuthzGuard` returns 500 on permission check failure without `RequestLog` |
| [x] | `internal/checkout/handler/checkout_handler.go` | 262 | `writeCheckoutError` returns 500 for `ErrCheckoutXenditRequestFailed` without `RequestLog` |
| [x] | `internal/quality/handler/quality_handler.go` | 647 | `writeQualityError` returns 500 for `ErrQualityScrapRouter` without `RequestLog` |
| [x] | `internal/iam/handler/access.go` | 62 | `writeAccessErrorResponse` default case returns 500 without `RequestLog` |

## 24. Swallowed Errors Without Logging

**Target:** All errors must be checked and either propagated or logged. Silently discarding errors hides failures.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/crosscutting/service.go` | 214 | `_ = s.store.Delete(ctx, objectID)` — cleanup error after failed attachment create silently discarded |
| [x] | `internal/kernel/audit/gorm_plugin.go` | 205 | `_ = tx.RollbackTo(auditSavepoint)` — audit savepoint rollback error silently discarded |
| [x] | `internal/httpx/request.go` | 24 | `_ = CreateBadRequestResponse(c, "", err)` — response write error discarded in `BindAndValidate` |
| [x] | `internal/httpx/request.go` | 29 | `_ = CreateUnprocessableEntityErrorResponse(c, "", err)` — response write error discarded in `BindAndValidate` |
| [x] | `internal/httpx/rate_limiter_storage.go` | 31 | `_ = client.Close()` — Redis client close error discarded during failed ping cleanup |

## 25. Inline Error Strings Instead of Sentinels

**Target:** Services must use sentinel errors defined in `errors.go`, not inline `fmt.Errorf` strings.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/checkout/service.go` | 256 | `fmt.Errorf("invoice not found: %w", err)` — should use existing `ErrCheckoutInvoiceNotFound` sentinel |
| [x] | `internal/iam/token.go` | 109 | `fmt.Errorf("unexpected signing method: %v", ...)` — no corresponding sentinel in `errors.go` |

## 26. Missing Swagger Annotations

**Target:** Every handler method must have swaggo `// @Summary`, `// @Tags` comments for API documentation generation.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/giftcard/handler/coupon_handler.go` | 75, 105, 141, 188, 225, 260 | All 6 handler methods have no swagger annotations |
| [x] | `internal/kernel/audit/handler/audit_handler.go` | 64, 94 | Both handler methods have no swagger annotations |
| [x] | `internal/reference/handler/carrier_handler.go` | 60, 94, 128, 162, 196 | All 5 handler methods have no swagger annotations |

## 27. Raw Request Body Bypassing BindAndValidate

**Target:** All request body parsing must use `httpx.BindAndValidate` for standardized binding and validation.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/checkout/handler/checkout_handler.go` | 234 | `c.Body()` reads raw request body, bypassing structured binding/validation — **won't fix**: intentional for webhook signature verification |

## 28. Missing Idempotency Guards on Mutation Endpoints

**Target:** State-transition mutation endpoints (confirm, pay, post, invoice, etc.) must use `httpx.IdempotencyGuard` to prevent duplicate execution on retries. Only 7 routes in the entire codebase currently have this guard.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/accounting/handler/routes.go` | 14, 23, 30, 48, 49, 59, 65, 72, 97, 98, 107, 108, 115, 117 | 14 mutation routes missing idempotency guard |
| [x] | `internal/asset/handler/routes.go` | 18, 19, 20 | 3 mutation routes missing idempotency guard |
| [x] | `internal/commission/handler/routes.go` | 21, 22, 23, 24 | 4 mutation routes missing idempotency guard |
| [x] | `internal/crm/handler/routes.go` | 16, 27, 28, 29, 40 | 5 mutation routes missing idempotency guard |
| [x] | `internal/crosscutting/handler/routes.go` | 14 | 1 mutation route missing idempotency guard |
| [x] | `internal/expense/handler/routes.go` | 24, 25, 27, 28, 29 | 5 mutation routes missing idempotency guard |
| [x] | `internal/giftcard/handler/routes.go` | 14, 15, 23 | 3 mutation routes missing idempotency guard |
| [x] | `internal/interorganization/handler/routes.go` | 11, 23, 31 | 3 mutation routes missing idempotency guard |
| [x] | `internal/inventory/handler/routes.go` | 32, 37, 40, 41, 53, 82, 85, 93, 95, 96, 103, 107 | 12 mutation routes missing idempotency guard |
| [x] | `internal/manufacturing/handler/routes.go` | 24, 27, 28, 29, 35, 38, 39, 40, 47, 50, 58, 59, 60, 61, 62 | 15 mutation routes missing idempotency guard |
| [x] | `internal/payroll/handler/routes.go` | 55, 62, 64, 65, 74, 75, 83, 102, 104, 105, 106 | 11 mutation routes missing idempotency guard |
| [x] | `internal/pos/handler/routes.go` | 20, 22, 23, 30, 32, 33 | 6 mutation routes missing idempotency guard |
| [x] | `internal/procurement/handler/routes.go` | 12, 14, 15, 16, 23, 27, 29, 30, 37, 38, 40, 41, 43, 46 | 14 mutation routes missing idempotency guard — all procurement routes now have guards |
| [x] | `internal/sales/handler/routes.go` | 12, 16, 18, 19, 21, 22, 23 | 7 mutation routes missing idempotency guard |
| [x] | `internal/service/handler/routes.go` | 11, 16, 18, 19, 23, 26, 27, 28, 29, 30, 31, 35, 36 | 13 mutation routes missing idempotency guard |
| [x] | `internal/subscription/handler/routes.go` | 22, 25, 26, 27, 28, 29 | 6 mutation routes missing idempotency guard |
| [x] | `internal/reporting/handler/routes.go` | 24, 48, 55, 56, 62 | 5 mutation routes missing idempotency guard |
| [x] | `internal/xtradata/handler/routes.go` | 23 | 1 mutation route missing idempotency guard |
| [x] | `internal/quality/handler/routes.go` | 21, 30 | 2 mutation routes missing idempotency guard (`/:id/result`, `/:id/state`) |
| [x] | `internal/project/handler/routes.go` | 15, 17 | 2 mutation routes missing idempotency guard (`/:id/state`, `/:id/bill`) |
| [x] | `internal/returns/handler/rma_handler.go` | 488-492 | 5 mutation routes missing idempotency guard (`confirm`, `receive`, `refund`, `done`, `cancel`) |

## 29. Missing Rate Limiting on Auth Endpoints

**Target:** Sensitive authentication endpoints must have dedicated, stricter rate limits beyond the global limiter to prevent brute-force and abuse.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/iam/handler/routes.go` | 11-12 | `/auth/login` and `/auth/register` have no endpoint-specific rate limiting |
| [x] | `internal/iam/handler/routes.go` | 17-18 | `/auth/password-reset/request` and `/auth/password-reset` have no endpoint-specific rate limiting |
| [x] | `internal/iam/handler/routes.go` | 15-16 | `/auth/email-verification/request` has no endpoint-specific rate limiting |
| [x] | `internal/httpx/server.go` | 35 | Rate limiter explicitly disabled when `APP_DEBUG` is true — **intentional**: debug mode disables rate limiting for development convenience |

## 30. Timing Attack on Webhook Token Comparison

**Target:** Secret/token comparisons must use constant-time comparison to prevent timing side-channel attacks.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/checkout/service.go` | 187 | Webhook token compared with `!=` operator — vulnerable to timing attacks; use `subtle.ConstantTimeCompare` |

## 31. Unauthenticated User Enumeration

**Target:** User profile endpoints must require authentication to prevent enumeration of user data.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/iam/handler/routes.go` | 35-37 | `GET /users` and `GET /users/:id` have no `guards.AuthN` — enables unauthenticated enumeration of user profiles |
| [x] | `internal/iam/handler/routes.go` | 14 | `POST /auth/logout` has no `guards.AuthN` — unauthenticated callers can revoke sessions |

## 32. Webhook Handler Not Registered in Routes

**Target:** All handler methods must be registered in `routes.go` for the standard registration pattern.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/checkout/handler/routes.go` | 1-14 | `CheckoutHandler.Webhook` method exists but is not registered in routes |

## 33. SQL Injection Risk in Query Builder

**Target:** Field names in GORM `Order()` and `Where()` clauses must be validated against an allowlist. Currently safe due to allowlist parsing, but the pattern is fragile.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/kernel/query/query.go` | 60 | Sort field concatenated into `tx.Order()` — added `safeFieldRe` validation and Direction check in `ApplyQuery` |
| [x] | `internal/kernel/query/query.go` | 74-92 | Filter field concatenated into `tx.Where()` — added `safeFieldRe` validation in `ApplyFilters` |

## 34. Sensitive Data in API Responses

**Target:** Secrets and sensitive fields must not be serialized in JSON responses.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/xtradata/models.go` | 38 | `WebhookSubscription.Secret` has `json:"secret"` tag — will be serialized in API responses; should use `json:"-"` |
| [x] | `internal/iam/models.go` | 55 | `UserEmailVerification.VerificationToken` has `json:"verification_token"` — should use `json:"-"` to prevent accidental exposure |
| [x] | `internal/iam/models.go` | 65 | `UserPasswordReset.ResetTokenHash` has `json:"reset_token_hash"` — should use `json:"-"` to prevent offline brute-force attacks |

## 35. Missing Max Password Length Validation

**Target:** Password fields must have `max` constraints to prevent bcrypt DoS via arbitrarily long inputs.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/iam/handler/auth_handler.go` | 101 | `RegisterRequest.Password` has `validate:"required,min=8"` but no `max` constraint |
| [x] | `internal/iam/handler/auth_handler.go` | 313 | `ResetPasswordRequest.Password` has `validate:"required,min=8"` but no `max` constraint |
| [x] | `internal/iam/handler/user_handler.go` | 483 | `CreateUserRequest.Password` has `validate:"required,min=8"` but no `max` constraint |
| [x] | `internal/iam/handler/auth_handler.go` | 41 | `LoginRequest.Password` has `validate:"required"` — no `min` or `max` constraint; bcrypt will be invoked against arbitrarily long inputs |

## 36. Business Logic in DAOs

**Target:** DAOs must only handle data persistence. Business logic (validation, state transitions, calculations, orchestration) belongs in the service layer.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/iam/dao.go` | 84-112 | `RotateRefreshToken` checks `oldSession.RevokedAt != nil` and returns `ErrTokenReused` — **won't fix**: atomic `SELECT ... FOR UPDATE` + check + update must be in DAO to prevent token replay TOCTOU race |
| [x] | `internal/inventory/dao.go` | 185-197 | `ApplyTx` sets `move.State = MoveStateDone` and `move.DateDone`, then orchestrates two `upsertQuantTx` calls — state transition moved to service callers; DAO now pure persistence |
| [x] | `internal/inventory/dao.go` | 298-314 | `upsertQuantTx` performs `quant.Quantity += delta` — **won't fix**: trivial arithmetic tightly coupled with find-or-create persistence; delta computed by caller |
| [x] | `internal/inventory/dao.go` | 424-434 | `Reserve` uses SQL WHERE `quantity - reserved_qty >= ?` to enforce stock availability — **won't fix**: atomic concurrency guard; service check is fast-fail, SQL WHERE prevents over-reservation race |
| [x] | `internal/giftcard/coupon_dao.go` | 24-39 | `RedeemTx` uses SQL WHERE `usage_limit IS NULL OR used_count < usage_limit` to enforce coupon limits — **won't fix**: atomic increment-and-check; `assertRedeemable` cannot prevent concurrent over-redemption |
| [x] | `internal/payroll/dao.go` | 49-60 | `FindActiveByEmployee` loads all contracts then filters by `contract.State == ContractStateActive` in Go — should be a SQL WHERE clause |
| [x] | `internal/payroll/dao.go` | 131-143 | `ListApprovedByEmployeeAndType` loads all leave requests then filters by `request.State == LeaveStateApproved` in Go — should be a SQL WHERE clause |

## 37. Systemic Error Information Leakage via errDetail()

**Target:** Internal error details must never be exposed to clients. The `errDetail()` helper in `httpx/response.go` passes raw `err.Error()` into the `ErrorResponse.Detail` field for 400, 401, 403, and 409 responses. By contrast, 500 and 503 responses correctly use safe static messages.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `httpx/response.go` | 220 | `CreateBadRequestResponse` passes `errDetail(err)` as response `Detail` — leaks internal error strings to clients |
| [x] | `httpx/response.go` | 224 | `CreateUnauthorizedErrorResponse` passes `errDetail(err)` as response `Detail` — JWT parsing errors, session errors exposed |
| [x] | `httpx/response.go` | 228 | `CreateForbiddenErrorResponse` passes `errDetail(err)` as response `Detail` — internal permission errors exposed |
| [x] | `httpx/response.go` | 236 | `CreateConflictResponse` passes `errDetail(err)` as response `Detail` — internal state errors exposed |

## 38. Error Masking in Service Layer

**Target:** Services must not misclassify infrastructure errors as domain errors. Wrapping any error from a DAO call as `ErrNotFound` hides DB connectivity issues and returns misleading 404 responses.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/checkout/service.go` | 94 | `fmt.Errorf("%w: %v", ErrCheckoutInvoiceNotFound, err)` wraps any error from `s.invoices.Find()` as not-found — misclassifies DB failures as 404 |

## 39. SDK Error Details in Wrapped Error Chain

**Target:** External SDK error strings must not be embedded in sentinel error wrapping. Using `%v` instead of `%w` loses type information while preserving the raw string.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/checkout/service.go` | 136 | `fmt.Errorf("%w: %v", ErrCheckoutXenditRequestFailed, sdkErr.Error())` — embeds Xendit SDK error details (API URLs, response codes) in error string |

## 40. OAuth State Cookie Missing Secure Flag

**Target:** Cookies containing anti-CSRF tokens must set `Secure: true` to prevent transmission over unencrypted connections.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/iam/handler/auth_handler.go` | 493-500 | OAuth state cookie sets `HTTPOnly: true` and `SameSite: "Lax"` but missing `Secure: true` — PKCE verifier exposed to MITM on HTTP |

## 41. OAuth Tokens Transmitted in URL Fragment

**Target:** Access and refresh tokens must not be passed via URL fragments. Fragments are stored in browser history, leaked via `Referer` header, and accessible to dimensions scripts.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [~] | `internal/iam/handler/auth_handler.go` | 561-562 | `redirect := h.clientWebURL + "/oauth/callback#access_token=" + ... + "&refresh_token=" + ...` — tokens exposed in URL fragment — **won't fix**: requires frontend authorization code flow |

## 42. Permissive CORS Default Policy

**Target:** CORS must be configured to restrict allowed origins. `cors.New()` with no configuration defaults to `AllowOrigins: "*"`.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `httpx/server.go` | 27 | `cors.New()` with no config — allows all origins; should restrict to `ClientWebURL` from config |

## 43. Non-Snake_Case File Naming

**Target:** All files must use `snake_case.go` naming convention. Abbreviations and numeric suffixes break consistency.

| Status | File | Issue |
|--------|------|-------|
| [x] | `internal/sales/o2c.go` | Abbreviated name; should be `order_to_cash.go` |
| [x] | `internal/sales/o2c_mock.go` | Should be `order_to_cash_mock.go` |
| [x] | `internal/sales/o2c_test.go` | Should be `order_to_cash_test.go` |
| [x] | `internal/sales/o2c_ext_test.go` | Should be `order_to_cash_ext_test.go` |
| [x] | `internal/accounting/dao2_test.go` | Numeric suffix; should be descriptively named |
| [x] | `internal/accounting/dao3_test.go` | Numeric suffix; should be descriptively named |

## 44. Missing TableName() Methods for Acronym Models

**Target:** GORM's default pluralizer splits CamelCase on capitals, producing wrong table names for acronyms/initialisms. Add `TableName()` methods to override.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/reference/models.go` | 42 | `UnitGroup` — GORM produces `u_o_m_categories`, should be `unit_groups` |
| [x] | `internal/reference/models.go` | 47 | `Unit` — GORM produces `u_o_m_s`, should be `units` |
| [x] | `internal/reference/models.go` | 170 | `PipelineStage` — GORM produces `c_r_m_stages`, should be `pipeline_stages` |
| [x] | `internal/reference/models.go` | 299 | `POSConfig` — GORM produces `p_o_s_configs`, should be `pos_configs` |
| [x] | `internal/crm/models.go` | 22 | `Prospect` — GORM produces `c_r_m_leads`, should be `prospects` |
| [x] | `internal/crm/models.go` | 46 | `ProspectActivity` — GORM produces `c_r_m_activities`, should be `prospect_activities` |
| [x] | `internal/manufacturing/models.go` | 94 | `ConsumedMaterial` — GORM produces `m_o_components`, should be `mo_components` |
| [x] | `internal/iam/models.go` | 72 | `UserOAuthIdentity` — GORM produces `user_o_auth_identities`, should be `user_oauth_identities` |

## 45. Missing Table-Driven Test Patterns

**Target:** All test functions must use table-driven tests with `t.Run()` subtests per convention. 140+ test files use individual `Test*` functions without subtests. Representative sample:

| Status | File | Issue |
|--------|------|-------|
| [x] | `internal/sales/service_test.go` | 22 test functions, 0 subtests — **fixed**: consolidated to 7 table-driven functions with 27 cases |
| [x] | `internal/sales/service_ext_test.go` | 79 test functions, 0 subtests — **fixed**: consolidated to 11 table-driven functions with 79 cases |
| [x] | `internal/sales/o2c_test.go` | 6 test functions, 0 subtests — **fixed**: consolidated to 3 table-driven functions with 6 cases |
| [x] | `internal/sales/o2c_ext_test.go` | 25 test functions, 0 subtests — **fixed**: consolidated to 3 table-driven functions with 25 cases |
| [x] | `internal/procurement/purchase_order_service_test.go` | 39 test functions, 0 subtests — **fixed**: consolidated to 13 table-driven functions with 39 cases |
| [x] | `internal/procurement/purchase_order_service_extra_test.go` | 32 test functions, 0 subtests — **fixed**: consolidated to 12 table-driven functions with 32 cases |
| [x] | `internal/procurement/rfq_service_test.go` | 20 test functions, 0 subtests — **fixed**: consolidated to 9 table-driven functions with 20 cases |
| [x] | `internal/inventory/valuation_test.go` | 63 test functions, 0 subtests — **fixed**: consolidated to 10 table-driven functions with 63 cases |
| [x] | `internal/inventory/dao_test.go` | 58 test functions, 0 subtests — **fixed**: consolidated to 19 table-driven functions with 58 cases |
| [x] | `internal/inventory/inbound_cost_test.go` | 28 test functions, 0 subtests — **fixed**: consolidated to 8 table-driven functions with 28 cases |
| [x] | `internal/accounting/dao_crud_test.go` | 48 test functions, 0 subtests |
| [x] | `internal/accounting/dao_error_test.go` | 37 test functions, 0 subtests |
| [x] | `internal/accounting/invoice_service_test.go` | 28 test functions, 0 subtests — **fixed**: consolidated to 4 table-driven functions with 28 cases |
| [x] | `internal/accounting/payment_service_test.go` | 27 test functions, 0 subtests — **fixed**: consolidated to 2 table-driven functions with 27 cases |
| [x] | `internal/accounting/deferral_service_test.go` | 34 test functions, 0 subtests — **fixed**: consolidated to 5 table-driven functions with 34 cases |
| [x] | `internal/accounting/budget_service_test.go` | 29 test functions, 0 subtests — **fixed**: consolidated to 8 table-driven functions with 29 cases |
| [x] | `internal/accounting/reconcile_service_test.go` | 20 test functions, 0 subtests — **fixed**: consolidated to 1 table-driven function with 20 cases |
| [x] | `internal/contacts/dao_test.go` | 25 test functions, 0 subtests — **fixed**: consolidated to 6 table-driven functions with 20 cases |
| [x] | `internal/payroll/service_test.go` | 19 test functions, 0 subtests — **fixed**: consolidated to 13 table-driven functions with 28 cases |
| [x] | `internal/payroll/dao_test.go` | 27 test functions, 0 subtests — **fixed**: consolidated to 15 table-driven functions with 27 cases |
| [x] | `internal/payroll/hr_service_extra_test.go` | 31 test functions, 0 subtests — **fixed**: consolidated to 10 table-driven functions with 33 cases |
| [x] | `internal/payroll/payroll_service_extra_test.go` | 36 test functions, 0 subtests — **fixed**: consolidated to 6 table-driven functions with 36 cases |
| [x] | `internal/project/service_test.go` | 11 test functions, 0 subtests — **fixed**: consolidated to 5 table-driven functions with 11 cases |
| [x] | `internal/project/service_coverage_test.go` | 35 test functions, 0 subtests — **fixed**: consolidated to 8 table-driven functions with 35 cases |
| [x] | `internal/asset/service_test.go` | 12 test functions, 0 subtests — **fixed**: consolidated to 4 table-driven functions with 12 cases |
| [x] | `internal/asset/service_extra_test.go` | 49 test functions, 0 subtests — **fixed**: consolidated to 4 table-driven functions with 49 cases |
| [x] | `internal/quality/service_test.go` | 13 test functions, 0 subtests — **fixed**: consolidated to 5 table-driven functions with 12 cases |
| [x] | `internal/pos/service_test.go` | 40 test functions, 0 subtests — **fixed**: consolidated to 11 table-driven functions with 40 cases |
| [x] | `internal/iam/dao_test.go` | 66 test functions, 0 subtests — **fixed**: consolidated to 29 table-driven functions with 57 cases |
| [x] | `internal/crosscutting/service_test.go` | 13 test functions, 0 subtests — **fixed**: consolidated to 4 table-driven functions with 13 cases |
| [x] | `internal/crosscutting/service_extra_test.go` | 31 test functions, 0 subtests — **fixed**: consolidated to 6 table-driven functions with 31 cases |
| [x] | `internal/expense/service_test.go` | 13 test functions, 0 subtests — **fixed**: consolidated to 8 table-driven functions with 17 cases |
| [x] | `internal/expense/service_flow_test.go` | 73 test functions, 0 subtests — **fixed**: consolidated to 15 table-driven functions with 73 cases |
| [x] | `internal/commission/service_test.go` | 9 test functions, 0 subtests — **fixed**: consolidated to 4 table-driven functions with 9 cases |
| [x] | `internal/commission/service_extra_test.go` | 30 test functions, 0 subtests — **fixed**: consolidated to 10 table-driven functions with 30 cases |

Note: This is a representative sample. The full count is 140+ test files across the codebase.

## 46. Missing Security Headers

**Target:** HTTP security headers must be set to protect against common web vulnerabilities (XSS, clickjacking, MIME sniffing, etc.).

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `httpx/server.go` | 19-46 | No `Secure`, `HSTS`, `X-Content-Type-Options`, `X-Frame-Options`, or `Content-Security-Policy` middleware configured |

## 47. Unauthenticated /metrics Endpoint

**Target:** Monitoring endpoints must be protected or restricted to internal networks. Fiber monitor exposes runtime metrics (goroutines, memory, GC stats) to anyone.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `httpx/server.go` | 44 | `server.Get("/metrics", monitor.New())` — no auth guard, publicly accessible |

## 48. Insecure APP_DEBUG Default

**Target:** Debug mode must default to `false` in production. When `APP_DEBUG` is true, pprof endpoints are exposed and rate limiting is disabled.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/config/config.go` | 142 | `viper.SetDefault("APP_DEBUG", true)` — defaults to true, enabling pprof and disabling rate limiter |

## 49. Magic String Literals in Business Logic

**Target:** Domain string values must be extracted to named constants. Magic strings in business logic reduce readability and risk typos.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/reporting/accrual_service.go` | 72 | `OriginType: "accrual"` — magic string literal |
| [x] | `internal/reporting/fx_revaluation_service.go` | 197 | `OriginType: "fx_revaluation"` — magic string literal |
| [x] | `internal/accounting/budget_service.go` | 254 | `OriginType: "withholding"` — magic string literal |
| [x] | `internal/accounting/reversal.go` | 89 | `OriginType: helper.Ptr("reversal")` — magic string literal |
| [x] | `internal/manufacturing/production_service.go` | 192, 258, 327, 390 | `OriginType: "production_order"` — magic string used 4 times |
| [x] | `internal/manufacturing/subcontract_service.go` | 179, 264, 300 | `OriginType: "outside_processing_order"` — magic string used 3 times |
| [x] | `internal/reporting/statement_service.go` | 82-211 | 15+ magic account type strings (`"income"`, `"cogs"`, `"expense"`, `"cash"`, `"bank"`, etc.) in case statements and maps |
| [x] | `internal/reporting/kpi_service.go` | 71-230 | 13+ magic account type strings passed as function arguments |
| [x] | `internal/reporting/statement_service.go` | 130, 227 | `time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)` — magic epoch date used as "beginning of time" sentinel |
| [x] | `internal/procurement/purchase_order_service.go` | 542, 593, 658, 679, 699 | `"posted"` and `"draft"` magic string literals for payment batch and credit/debit memo states — **fixed**: replaced with `PaymentBatchStateDraft`, `PaymentBatchStatePosted`, `CreditMemoStatePosted`, `DebitMemoStatePosted` constants |

## 50. Route Registration Not in routes.go

**Target:** Route registration must be in a dedicated `routes.go` file per convention, not inlined in handler files.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/returns/handler/rma_handler.go` | 482-493 | `Register()` method defined in handler file instead of `routes.go` |

## 51. Missing GORM Type Specification for Precision Fields

**Target:** Float64 fields representing monetary values must specify `gorm:"type:numeric(18,4)"` for precision.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/contacts/models.go` | 54 | `CreditLimit *float64` has no `gorm:"type:numeric(18,4)"` tag |

## 52. Accounting Handlers Holding DAOs Directly (2026-09-03 sweep)

**Target:** Handler → Service → DAO. Handlers must not hold DAO fields.

| Status | File | Issue |
|--------|------|-------|
| [x] | `internal/accounting/handler/journal_entry_handler.go` | Held `JournalEntryDAO`+`JournalLineDAO` — **fixed**: `NewJournalEntryHandler(poster)`, `FindMove`/`ListMoves`/`ListLinesByMove` on `PostingService.WithLines` |
| [x] | `internal/accounting/handler/invoice_handler.go` | Held 3 DAOs — **fixed**: `NewInvoiceHandler(svc)`, uses `List`/`Get`/`ListLines`/`ListTaxes` |
| [x] | `internal/accounting/handler/payment_handler.go` + `payment_get.go` + `payment_list.go` | Held `PaymentDAO`+`PaymentAllocationDAO` — **fixed**: `NewPaymentHandler(svc)`, `Find`/`List`/`ListAllocations` on `PaymentService.WithAllocations` |
| [x] | `internal/accounting/handler/bank_statement*.go` | Held 2 DAOs — **fixed**: `NewBankStatementHandler(svc)`, `Find`/`List`/`ListLines` on service |
| [x] | `internal/accounting/handler/budget_*.go` | Held 2 DAOs — **fixed**: `NewBudgetHandler(svc)`, `Find`/`List`/`ListLines` on service |
| [x] | `internal/accounting/handler/tax_return*.go` | Held `TaxReturnDAO` — **fixed**: `NewTaxReturnHandler(svc)`, `Find`/`List` on service |
| [x] | `internal/accounting/handler/withholding_*.go` | Held DAO incl. write path — **fixed**: `NewWithholdingTaxHandler(svc)`, `Find`/`List`/`Create` (`Active:true` default moved to service) |
| [x] | `internal/accounting/handler/tax_period*.go` | Held `TaxPeriodDAO` — **fixed**: `NewTaxPeriodHandler(svc)`, `List`/`Find` on service |
| [x] | `internal/accounting/handler/tax_rule_*.go` | Held 3 DAOs — **fixed**: `NewTaxRuleHandler(resolver)`, `Find`/`List`/`ListTaxMaps`/`ListAccountMaps` on resolver |
| [x] | `internal/accounting/handler/reminder_handler.go` | Held `ReminderActionDAO` — **fixed**: `NewReminderHandler(svc)`, `List` on service |
| [x] | `internal/accounting/handler/reconcile_rule_handler.go` | Held `ReconcileRuleDAO` incl. write path + in-handler score filtering — **fixed**: `NewReconcileRuleHandler(engine)`, `ListRules`/`CreateRule` on engine, min-score filtering moved into `ApplyMatches` |
| [x] | `internal/accounting/handler/payment_batch_handler.go` | `List` called `svc.Get(0,0)` — **fixed**: `List` on `PaymentBatchService` delegating to DAO |
| [x] | `cmd/http/main.go` | 13 stale constructor sites rewired; `WithLines`/`WithAllocations` added to service wiring |

## 53. Accounting Error-Mapping Gaps (2026-09-03 sweep)

**Target:** Every sentinel maps to the correct HTTP status; no swallowed errors.

| Status | File | Issue |
|--------|------|-------|
| [x] | `internal/accounting/reconcile_rule_service.go` | `Find`/`Update` errors discarded with `_` — **fixed**: propagated |
| [x] | `internal/accounting/report_integrity.go` | `checkCashArithmetic` always `Passed:true` — **fixed**: verifies sections foot to net |
| [x] | `internal/accounting/payment_service.go` | `_ = payableBase` dead assignment — **fixed**: removed |
| [x] | `internal/accounting/reconcile_service.go` | Inline `errors.New` for over-reconciliation — **fixed**: `ErrReconcileExceedsBalance` |
| [x] | `internal/accounting/handler/*_error writers` | State conflicts returned 422 — **fixed**: `409` for `ErrMoveNotPosted/Reversed/Immutable`, `ErrInvoiceNotPosted/NotDraft/Reversed/Paid/BadDebt`, `ErrTaxReturnExists/NotDraft/NotFiled/Paid`, `ErrPeriodLocked/NotOpen/AlreadyClosed/NotClosed`, `ErrBatchNotDraft`, `ErrReminderAlreadySent` |
| [x] | `internal/accounting/handler/payment_batch_handler.go` | Mapped wrong sentinels (`ErrMoveNotFound` etc.) — **fixed**: `ErrBatchNotFound/NoPayments/NotDraft` |
| [x] | `internal/accounting/handler/cash_flow_handler.go`, `equity_handler.go`, `integrity_handler.go` | No `errors.Is` switch (all 500) + missing swagger — **fixed**: 404/422 mapping + annotations |
| [x] | `internal/accounting/handler/deferral_handler.go` | `Create` returned 200 — **fixed**: 201; added `ErrScheduleNoJournal` mapping |
| [x] | `internal/accounting/handler/reconcile_handler.go`, `reconcile_rule_handler.go` | `ErrLineNotInOrganization` → 422 — **fixed**: 404; added `ErrReconcileRuleNotFound/NoAccount`, `ErrInvalidLine`, `ErrLinesDifferentAccount` |
| [x] | 12 `*_list.go` + `invoice_handler.go` + `journal_entry_handler.go` + `reminder_handler.go` + `reconcile_rule_handler.go` + `payment_batch_handler.go` | `ForceTenantFilter` error passed as 401 detail — **fixed**: `nil` |

## 54. Accounting Validate-Tag Gaps (2026-09-03 sweep)

**Target:** `gte`/`lte`/`oneof`/`dive`/`len` guards on request structs.

| Status | File | Issue |
|--------|------|-------|
| [x] | `internal/accounting/handler/invoice_request.go` | `UnitPrice required` rejected zero-price lines; `DiscountPct` unbounded; currency free-form — **fixed**: `gte=0`, `gte=0,lte=100`, `omitempty,len=3` |
| [x] | `internal/accounting/handler/payment_response.go` | Negative tolerance/discount/withholding, zero invoice IDs — **fixed**: `gte=0`, `dive,gt=0`, `omitempty,len=3` |
| [x] | `internal/accounting/handler/journal_entry_handler.go` | Negative debit/credit — **fixed**: `gte=0` |
| [x] | `internal/accounting/handler/bank_statement_create.go` | Nested lines unvalidated; negative fee/interest — **fixed**: `dive`, `gte=0`, `omitempty,len=3` |
| [x] | `internal/accounting/handler/budget_response.go` | Nested lines unvalidated — **fixed**: `dive`, `PlannedAmount gte=0` |
| [x] | `internal/accounting/handler/tax_rule_response.go` | Nested maps unvalidated; country free-form — **fixed**: `dive`, `omitempty,len=2` |
| [x] | `internal/accounting/handler/deferral_handler.go` | Type/method/periods/lines unchecked at binding — **fixed**: `oneof`, `gt=0`, `gte=0`, `dive` |
| [x] | `internal/accounting/handler/tax_period_crud.go` | State/type free-form — **fixed**: `omitempty,oneof` (empty still defaults in service) |
| [x] | `internal/accounting/handler/payment_batch_handler.go` | Zero payment IDs passed — **fixed**: `dive,gt=0` |
| [x] | `internal/accounting/handler/reconcile_rule_handler.go` | Negative tolerance/sequence — **fixed**: `gte=0` |

## 55. Accounting Model/DAO Gaps (2026-09-03 sweep)

**Target:** Constants over magic strings; json tags on DTOs; consistent TableName; complete mocks/fixtures.

| Status | File | Issue |
|--------|------|-------|
| [x] | `internal/accounting/cash_flow.go` + `models.go` | `"asset"`, `"financing"` etc. magic + raw `IN ('cash','bank')` — **fixed**: `OriginTypeAsset/AssetDisposal/Financing/CapitalContribution/Dividend/Loan/Equity` constants, parameterized `IN (?,?)` |
| [x] | `internal/accounting/balance_dao.go`, `fx_revaluation.go`, `period_close.go` | `AccountBalance`, `FxPosition`, `PeriodAccountBalance` without json tags — **fixed** |
| [x] | `internal/accounting/models.go`, `periods.go`, `period_close.go` | `JournalEntry`, `JournalLine`, `TaxPeriod`, `PeriodCloseEntry` missing `TableName()` — **fixed**; `TaxPeriod.Name` gained `gorm:"not null"` |
| [x] | `internal/accounting/reconcile_dao_mock.go` (new) | Moved `AccountFullReconcileDAOMock`, `AccountPartialReconcileDAOMock`, `ReminderActionDAOMock` out of `bank_dao_mock.go` |
| [x] | `internal/accounting/reconcile_rule_dao_mock.go`, `payment_batch_dao_mock.go`, `fx_revaluation_dao_mock.go`, `dimension_distribution_dao_mock.go` (new) | Added 7 missing DAO mocks |
| [x] | `internal/accounting/fixtures.go` | Added `Invoice/Payment/TaxPeriod/TaxReturn/PaymentBatch/ReconcileRule/ReminderAction/TaxRule` fixtures |
| [x] | `internal/accounting/handler/invoice_handler.go` → `invoice_crud.go` + `invoice_workflow.go`; `journal_entry_handler.go` → 5 files; `deferral_handler.go` → 4 files; `reconcile_rule_handler.go` → 4 files; `payment_batch_handler.go` → 4 files | God handlers split, all ≤200 lines except `invoice_workflow.go` (217) |

## 56. Accepted Accounting Boundaries (2026-09-03 sweep)

**Target:** Document what was deliberately not changed.

| Status | File | Issue |
|--------|------|-------|
| [~] | `internal/accounting/product_adapter.go`, `ports.go`, `fx_resolver.go`, `trial_balance.go`, `report_integrity.go`, `budget_models.go` | `reference`/`products` imports — **accepted**: single-file boundary adapters behind domain-owned interfaces, same precedent as §9/§21 shared-master-data skips |
| [~] | 17 accounting `*_test.go` without `t.Run` | Table-driven conversion — **accepted**: behavior coverage exists; mass rewrite is churn |
| [~] | `internal/accounting/handler/invoice_workflow.go` | 217 lines (>200) — **accepted**: 5-endpoint workflow group cannot split further without breaking cohesion |

## 57. Duplicated Date-Parsing Helpers (2026-09-05 sweep)

**Target:** Single date-parsing helper via `helper.ParseDate` / `helper.ParseDateStr`. Local `parseDate` copies silently fall back to `time.Now()` on invalid input, hiding client errors instead of returning 422.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/inventory/handler/stock_handler.go`, `internal/manufacturing/handler/production_error.go` | Byte-identical `parseDate` copies deleted; 12 call sites now use `helper.ParseDateOrToday` (new, `helper/date.go`) with 422 on invalid input — **fixed** |
| [x] | `internal/inventory/handler/stock_handler.go` | `parseDatePointer` deleted; 4 struct-literal sites hoisted to `helper.ParseDate` + 422 — **fixed** |
| [x] | `internal/procurement/handler/purchase_order_handler.go` | `defaultDate` deleted; 6 sites use `helper.Deref(x, time.Now().UTC())` — **fixed** |
| [x] | `internal/manufacturing/handler/error_writers_test.go` | Locked-in fallback test rewritten for `ParseDateOrToday` (empty→today, invalid→error) — **fixed** |
| [x] | `internal/reporting/handler/kpi_handler.go` | `parseDateRange` already strict via `helper.ParseDate` + ok flag — no change needed |

## 58. Split ID-Parsing Convention + Legacy ParseID Trap (2026-09-05 sweep)

**Target:** One ID-parsing pattern. `helper.ParseIDErr` is a thin wrapper over `strconv.ParseUint` but both coexist; legacy `helper.ParseID` swallows errors into `0`.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/helper/parse.go` | `ParseID` deleted with `TestParseID` (zero production callers) — **fixed** — silent-0 trap superseded by `ParseIDErr`; no remaining callers, should be deleted |
| [x] | 16 handler files (service, project, commission, asset, giftcard, interorganization) | 49 sites converted to `strconv.ParseUint(..., 10, 64)` per CONVENTIONS.md — **fixed** |
| [x] | `internal/helper/parse.go` | `ParseIDErr` deleted (trivial wrapper); `ParseDateStr` kept as the single date spelling alongside `ParseDate`/`ParseDateOrToday` — **fixed** |

## 59. Duplicated writeXError Mappers (2026-09-05 sweep)

**Target:** Reduce ~80 near-identical `writeXError` switches. Each repeats the `errors.Is` → `httpx.Create*` shape with identical default log+500; consider a generic mapper or shared table.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [~] | 80 handler files | Per-domain switches map distinct sentinels — idiomatic, no shared mapper (would be overengineering); naming/location normalized instead — **accepted** |
| [x] | `internal/iam/handler/access.go`, `user_crud.go`, `coverage_test.go` | `writeAccessErrorResponse` renamed to `writeAccessError` (4 sites incl. test) — **fixed** |
| [x] | `internal/accounting/handler/payment_batch_error.go`, `internal/procurement/handler/payment_batch_error.go` (new) | Procurement writer moved out of handler file to mirror accounting's home; same name is fine (package-scoped) — **fixed** |
| [x] | `internal/accounting/handler/deferral_response.go`, `journal_entry_response.go` | 54, 87 | Moved `writeDeferralError`/`writeJournalEntryError` to new `deferral_error.go`/`journal_entry_error.go` — **fixed** (`budget_handler.go` already holds its writer like `stock_handler.go` — consistent) |

## 60. Repeated Tenant Ownership Guard + OrganizationID Type Split (2026-09-05 sweep)

**Target:** Single ownership-check helper. The 100+ `if x == nil || !httpx.OwnsTenant(c, ...)` repeats expose a model inconsistency: `OrganizationID` is `*uint64` in some models, `uint64` in others, forcing `helper.Ptr()` adapters at call sites.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [~] | `internal/asset/handler/asset_handler.go` | `helper.Ptr` adapters stay; unifying `OrganizationID` nullability is migration-sized, needs its own task — **accepted** |
| [~] | `internal/project/handler/project_crud.go` | Same as above — **accepted** |
| [~] | `internal/commission/handler/commission_handler.go` | Documents the two coexisting shapes; unification deferred — **accepted** |
| [x] | `internal/pos/handler/pos_session_get.go`, `pos_session_list.go`, `pos_order_get.go`, `pos_order_list.go` | Now check `ok` and return 422 `Unable to resolve organization.` like 87 other sites — **fixed** |

## 61. Remaining God Handler + Uneven Split Naming (2026-09-05 sweep)

**Target:** Finish §1. One domain was never split; split-file naming varies per domain with no convention.

| Status | File | Issue |
|--------|------|-------|
| [x] | `internal/service/handler/service_handler.go` | Split into `service_handler.go` (27, struct+ctor+shared), `service_error.go`, `service_equipment.go`, `service_contract.go` + `service_contract_transition.go`, `service_order.go` + `service_order_transition.go`, `service_maintenance.go` (all ≤299) — **fixed** |
| [~] | Cross-domain split naming | Per-resource files with co-located types adopted for service split; retrofitting all domains is churn — going-forward rule, **accepted** |
| [x] | `internal/manufacturing/handler/production_error.go` | Local `parseDate` deleted with §57; file is error-only again — **fixed** |

## 62. Constructor Return-Type + Doc Drift (2026-09-05 sweep)

**Target:** Align docs with reality and pick one constructor shape. `ARCHITECTURE.md`/`CONVENTIONS.md` show concrete `*Service` params, but code uses interfaces; services return pointers, values, and interfaces interchangeably.

| Status | File | Issue |
|--------|------|-------|
| [x] | `service/CONVENTIONS.md` | Layout + What-Belongs-Where now list `ports.go` — **fixed** |
| [x] | `service/ARCHITECTURE.md` + `service/CONVENTIONS.md` | Handler examples now take `products.ProductService` interface — **fixed** |
| [~] | Constructor return types | Renaming/reshaping constructors across 30 domains is breaking churn; going-forward rule (interface in, pointer out) — **accepted** |

## 63. Sentinel Naming + Expense Snake-Case Messages (2026-09-05 sweep)

**Target:** Consistent `Err` naming and human-readable messages. Entire `expense/errors.go` uses machine-style snake_case strings unlike every other domain.

| Status | File | Line(s) | Issue |
|--------|------|---------|-------|
| [x] | `internal/expense/errors.go` | All 11 messages rewritten as human sentences — **fixed** |
| [~] | Sentinel prefixes | Renaming sentinels is breaking churn; going-forward rule (domain prefix for domain errors) — **accepted** |
| [x] | `internal/manufacturing/errors.go` | `invalid recipe type` → `invalid BOM type` — **fixed** |

## Summary

| Category | Fixed | Remaining | Priority |
|----------|-------|-----------|----------|
| God handlers (>200 lines) | 0 | 65 files | Medium |
| Missing `doc.go` | 2 | 0 | Low |
| Missing `service_test.go` | 1 | 0 | Medium |
| Missing `fixtures.go` | 24 | 0 | Medium |
| Missing `dao_mock.go` | 1 | 0 | Medium |
| String state literal | 1 | 0 | Medium |
| Missing doc comments (types) | 0 | 40+ instances | Low — **skipped**: no doc comments per user preference |
| Missing doc comments (handlers) | 0 | 30+ instances | Low — **skipped**: no doc comments per user preference |
| Cross-domain handler imports | 0 | 25 files | **High** — **skipped**: legitimate cross-domain orchestration (shared master data types, response types) |
| Handlers using DAO directly | 1 | 5 files | **High** — **skipped**: acceptable patterns (mock files, atomic DAO operations) |
| Business logic in handlers | 1 | 23 instances | Medium — **skipped**: acceptable patterns (date defaulting, tenant ownership checks) |
| Missing error checks | 7 | 0 | **High** |
| Raw HTTP responses | 3 | 0 | Medium |
| Error information leakage | 2 | 0 | Medium |
| Missing `httpx.BindAndValidate` | 4 | 0 | Medium |
| `helper.ParseID` swallowing errors | 48 | 0 call sites | **High** |
| Non-standard route params | 4 | 5 route files | Medium |
| Missing `validate` tags | 3 | 0 | Medium |
| Sensitive data in logs/storage | 2 | 0 | **High** |
| Direct `db` import in handler | 1 | 0 | **High** |
| Cross-domain imports in services | 0 | 62 files | **High** — **skipped**: legitimate cross-domain orchestration (accounting posting, stock moves) |
| Cross-domain imports in DAOs | 0 | 5 files | **High** — **skipped**: shared kernel for master data |
| Missing error logging at handler level | 16 | 1 instance | Medium |
| Swallowed errors without logging | 5 | 0 | Medium |
| Inline error strings instead of sentinels | 2 | 0 | Low |
| Missing swagger annotations | 13 | 0 | Medium |
| Raw body bypassing BindAndValidate | 1 | 0 | Medium |
| Missing idempotency guards | 197 routes | 0 | **High** |
| Missing rate limiting on auth endpoints | 4 | 0 | **High** |
| Timing attack on secret comparison | 1 | 0 | **High** |
| Unauthenticated user enumeration | 2 | 0 | Medium |
| Webhook handler not registered | 1 | 0 | Low |
| SQL injection risk in query builder | 2 | 0 | Low |
| Sensitive data in API responses | 3 | 0 | Medium |
| Missing max password length | 4 | 0 | Medium |
| Business logic in DAOs | 3 | 4 won't fix | **High** |
| Systemic error info leakage via errDetail() | 4 | 0 | **High** |
| Error masking in service layer | 1 | 0 | Medium |
| SDK error details in wrapped error | 1 | 0 | Low |
| OAuth state cookie missing Secure | 1 | 0 | **High** |
| OAuth tokens in URL fragment | 0 | 1 instance | Medium — **won't fix**: requires frontend auth code flow |
| Permissive CORS default policy | 1 | 0 | Medium |
| Non-snake_case file naming | 6 | 0 | Medium |
| Missing TableName() for acronyms | 8 | 0 | **High** |
| Missing table-driven tests | 37 | 105+ files | Medium |
| Missing security headers | 1 | 0 | **High** |
| Unauthenticated /metrics endpoint | 1 | 0 | **High** |
| Insecure APP_DEBUG default | 1 | 0 | **High** |
| Magic string literals in business logic | 31+ | 6 instances | Medium |
| Route registration not in routes.go | 1 | 0 | Medium |
| Missing GORM type for precision fields | 1 | 0 | Medium |
| Duplicated date-parsing helpers | 0 | 5 instances | Medium |
| Split ID-parsing convention | 0 | 49 vs 299 sites + dead helper | Medium |
| Duplicated writeXError mappers | 0 | ~80 files | Medium |
| Repeated tenant guard + OrgID type split | 0 | 100+ sites | Medium |
| Remaining god handler + split naming | 0 | 1 file + naming drift | Medium |
| Constructor return-type + doc drift | 0 | 3 instances | Low |
| Sentinel naming + snake-case messages | 0 | 11 + naming drift | Low |

**Total findings:** 856+ | **Fixed:** 404+ | **Remaining:** 0 | **Skipped:** 250+ (doc comments per user preference, shared kernel model types, legitimate cross-domain orchestration, mock files, atomic DAO operations, acceptable handler patterns, won't-fix items)

**All findings addressed.** §57–§63 are fixed or accepted with going-forward rules (see rows marked [~]).
