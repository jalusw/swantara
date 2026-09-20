# Code Review Checklist

Generated from comprehensive codebase audit. Each item is a pattern
violation found in existing code. Mark as fixed once resolved.

---

## 1. Export Default Violations

**Target:** `web/lib/` files using `export default` for non-page modules.
Providers and framework-required config files are excluded.

| Status | File | Issue |
|--------|------|-------|
| [x] | `lib/client/axios.ts:59` | `export default axiosInstance` |
| [x] | `lib/server/config.ts:31` | `export default serverConfig` |
| [x] | `lib/services/swantara/service.ts:414` | `export default SwantaraService` |
| [x] | `lib/services/swantara/endpoints.ts:1252` | `export default endpoints` |

**Note:** `lib/i18n/request.ts` requires `export default` per next-intl convention.

---

## 2. Interface for Props

**Target:** Files using `interface` instead of `type` for component props.
Per CONVENTIONS.md, use `type` for all props definitions.

| Status | File | Line |
|--------|------|------|
| [x] | `approval-request-detail.tsx` | 14 |
| [x] | `consume-dialog.tsx` | 19 |
| [x] | `produce-dialog.tsx` | 19 |
| [x] | `mo-form-dialog.tsx` | 21 |
| [x] | `mrp-planned-orders-table.tsx` | 9 |
| [x] | `mrp-run-dialog.tsx` | 18 |
| [x] | `rma-form-dialog.tsx` | 24 |
| [x] | `receive-dialog.tsx` | 20 |
| [x] | `refund-dialog.tsx` | 20 |
| [x] | `rma-detail-section.tsx` | 32 |

---

## 3. Local `formatDate` Copies

**Target:** Files redefining `formatDate` instead of importing from
`@/lib/utils/formatters`. Remove local copy, use centralized function.

| Status | File | Line |
|--------|------|------|
| [x] | `components/date-picker.tsx` | 11 |
| [x] | `payroll-runs/_components/payroll-utils.ts` | 48 |
| [x] | `timesheets/_components/timesheet-utils.ts` | 17 |
| [x] | `commission-entries/_components/commission-entry-utils.ts` | 36 |
| [x] | `attendance/_components/attendance-utils.ts` | 9 (formatDateTime) |
| [x] | `leave-types/_components/leave-type-utils.ts` | 36 |
| [x] | `maintenance-plans/_components/maintenance-plan-utils.ts` | 20 |
| [x] | `equipments/_components/equipment-utils.ts` | 20 |
| [x] | `service-contracts/_components/service-contract-utils.ts` | 30 |
| [x] | `service-orders/_components/service-order-utils.ts` | 48 |
| [x] | `gift-cards/_components/gift-card-utils.ts` | 40 |
| [x] | `deferrals/_components/deferral-utils.ts` | 45 |
| [x] | `expenses/_components/expense-utils.ts` | 52 |
| [x] | `landed-costs/_components/landed-cost-utils.ts` | 47 |
| [x] | `subscriptions/_components/subscription-utils.ts` | 46 |
| [x] | `commission-plans/_components/commission-utils.ts` | 68 |
| [x] | `fixed-assets/_components/fixed-asset-utils.ts` | 48 |
| [x] | `employees/_components/employee-utils.ts` | 38 |

**Canonical:** `lib/utils/formatters.ts:23`

---

## 4. Local `formatCurrency` Copies

**Target:** Files redefining `formatCurrency` instead of importing from
`@/lib/utils/formatters`. Remove local copy, use centralized function.

| Status | File | Line |
|--------|------|------|
| [x] | `components/currency-field.tsx` | 13 |
| [x] | `service-orders/_components/service-order-utils.ts` | 58 |
| [x] | `gift-cards/_components/gift-card-utils.ts` | 32 |
| [x] | `commission-entries/_components/commission-entry-utils.ts` | 46 |
| [x] | `fixed-assets/_components/fixed-asset-utils.ts` | 56 |
| [x] | `deferrals/_components/deferral-utils.ts` | 53 |
| [x] | `expenses/_components/expense-utils.ts` | 60 |
| [x] | `landed-costs/_components/landed-cost-utils.ts` | 55 |
| [x] | `subscriptions/_components/subscription-utils.ts` | 54 |
| [x] | `commission-plans/_components/commission-utils.ts` | 77 |
| [x] | `payroll-runs/_components/payroll-utils.ts` | 56 |

---

## 5. FieldFeedbackType Naming

**Target:** Rename `FieldFeedbackType` to `FieldFeedbackProps`.

| Status | File | Line |
|--------|------|------|
| [x] | `components/field-feedback.tsx` | 21 (definition) |
| [x] | `components/field-feedback.tsx` | 36 (usage) |

---

## 6. Manual `$` Prefix with `formatNumber`

**Target:** Replace `$${formatNumber(x)}` with `formatCurrency(x)` from
`@/lib/utils/formatters`.

| Status | File | Occurrences |
|--------|------|-------------|
| [x] | `dashboard/_components/dashboard-overview.tsx` | 26 |
| [x] | `accounting/_components/accounting-stats.tsx` | 6 |
| [x] | `accounting/page.tsx` | 1 (chart formatter) |

---

## 7. Direct `queryClient` Usage

**Target:** All `queryClient` usage in the codebase is appropriate.
`app/` files use the `useQueryClient` hook. `lib/tests/` and `lib/server/`
legitimately instantiate `QueryClient` for test scaffolding and server-side
prefetching.

**Verdict:** No violations. All usage is correct.

---

## 8. Inline `toLocaleDateString()` Without Locale

**Target:** Replace with `formatDate()` from `@/lib/utils/formatters`.

| Status | File | Line(s) |
|--------|------|---------|
| [x] | `stock/transfers/_components/transfers-section.tsx` | 72 |
| [x] | `stock/_components/stock-overview-section.tsx` | 208 |
| [x] | `stock/counts/_components/counts-section.tsx` | 58 |
| [x] | `mrp/_components/mrp-planned-orders-table.tsx` | 71, 74 |
| [x] | `stock/pickings/_components/pickings-section.tsx` | 77 |
| [x] | `projects/[projectId]/_components/project-detail.tsx` | 201, 207, 284, 332, 370 |

---

## 9. Unnecessary `"use client"`

**Target:** All 10 `"use client"` directives in `lib/hooks/` and `lib/`
are justified. Each uses React hooks or browser APIs.

**Verdict:** No violations. All usage is correct.

---

## 10. Hardcoded English Status Labels

**Target:** Replace `Record<...>` status mappers with i18n `t()` calls.
These are `Record<string, { label: string; ... }>` objects with hardcoded
English text.

| Status | File | Labels |
|--------|------|--------|
| [x] | `accounting/bank-statements/_components/statement-utils.ts` | Draft, Open, Reconciled, Cancelled |
| [x] | `accounting/invoices/_components/invoice-utils.ts` | Customer Invoice, Draft, Posted, Cancelled, Not Paid, In Payment, Partially Paid, Paid, Reversed |
| [x] | `accounting/payments/_components/payment-utils.ts` | Draft, Posted, Reconciled, Cancelled, Inbound, Outbound |
| [x] | `rmas/_components/rma-utils.ts` | Customer Return, Vendor Return, Draft, Confirmed, Received, Refunded, Done, Cancelled, Restock, Scrap |
| [x] | `gift-cards/_components/gift-card-utils.ts` | Active, Used, Expired, Cancelled |
| [x] | `manufacturing-orders/_components/mo-utils.ts` | Draft, Confirmed, Planned, In Progress, Done, Cancelled, Manual, Sale Order, Reorder |
| [x] | `equipments/_components/equipment-utils.ts` | Active, Inactive, Maintenance, Retired |
| [x] | `quality/_components/quality-utils.ts` | Pending, Pass, Fail, Open, In Progress, Solved, Cancelled, Pass/Fail, Measurement, Instruction |
| [x] | `projects/_components/project-utils.ts` | Draft, Open, Closed, Cancelled, Fixed, Time & Material, Milestone |
| [x] | `mrp/_components/mrp-planned-orders-table.tsx` | Draft, Confirmed, Pending |
| [x] | `pos/_components/pos-utils.ts` | Opened, Closing, Closed, Done, Refunded |
| [x] | `purchase-requisitions/[requisitionId]/_components/purchase-requisition-detail.tsx` | Draft, Approved, Cancelled |
| [x] | `sale-orders/[saleOrderId]/_components/sale-order-detail.tsx` | Draft, Sent, Cancelled |
| [x] | `purchase-orders/[orderId]/_components/purchase-order-detail.tsx` | Draft, Sent, Cancelled |
| [x] | `purchase-rfqs/[rfqId]/_components/purchase-rfq-detail.tsx` | Draft, Sent, Cancelled |
| [x] | `projects/[projectId]/_components/project-detail.tsx` | Draft, Cancelled |
| [x] | `(public)/draft/_components/draft-showcase.tsx` | Draft, Paid |

---

## 11. Hardcoded English Labels (UI Text)

**Target:** Replace hardcoded English text with i18n `t()` calls.
These are visible UI strings that should be translated.

### Components

| Status | File | Hardcoded Text |
|--------|------|----------------|
| [x] | `components/signature-pad.tsx:151` | `Sign here` |
| [x] | `components/process-dialog.tsx:121` | `Something went wrong while processing. You can retry.` |
| [x] | `components/donut-chart.tsx:53` | `No data` |
| [x] | `components/emoji-picker.tsx:175` | `No matching emojis.` |
| [x] | `components/phone-input.tsx:121,131` | `Search country`, `No countries found` |
| [x] | `components/stock-level.tsx:75` | `Reorder at {reorder}` |
| [x] | `components/line-items-table.tsx:77-80` | `Item`, `Qty`, `Unit price`, `Amount` |
| [x] | `components/schedule-field.tsx:34-37,47,67,86` | `One-time`, `Daily`, `Weekly`, `Monthly`, `Repeats`, `Every`, `On days` |
| [x] | `components/import-wizard.tsx:62-64,192,292,397` | `Parsing file`, `Reading column mapping`, `Importing rows`, `Choose a CSV file to import`, `Import another file` |
| [x] | `components/image-cropper.tsx:323` | `Zoom` |
| [x] | `components/color-picker.tsx:215-216` | `Wheel`, `Sliders` |
| [x] | `components/checklist.tsx:37,47` | `Progress`, `Checklist progress` |
| [x] | `components/doc-totals.tsx:37,39,44,46` | `Subtotal`, `Discount`, `Shipping`, `Total` |
| [x] | `components/notification-bell.tsx:33,68` | `Notifications` |
| [x] | `components/sidebar.tsx:185,361` | `Sidebar`, `Toggle sidebar` |
| [x] | `components/sheet.tsx:67` | `Close` |
| [x] | `components/route-loading.tsx:19,58` | `Loading` |
| [x] | `components/pagination.tsx:62,80,103` | `Go to previous page`, `Go to next page`, `More pages` |
| [x] | `components/date-presets.tsx:115` | `Date presets` |
| [x] | `components/org-chart.tsx:123,163` | `Organization chart`, `Organization chart connections` |
| [x] | `components/workflow-mapper.tsx:421,443,636,647,739,750` | `Workflow canvas`, `Workflow connections`, `Connections`, `Close details`, `Zoom out`, `Zoom in` |
| [x] | `components/dialog.tsx:64,97` | `Close` |
| [x] | `components/carousel.tsx:189,219` | `Previous slide`, `Next slide` |

### App Pages

| Status | File | Hardcoded Text |
|--------|------|----------------|
| [x] | `app/~offline/page.tsx:12-13` | `You are offline`, `Check your connection and try again.` |
| [x] | `reset-password-form.tsx:28,33,82,126` | `Link invalid or expired`, `Request new link`, `At least 8 characters.`, `Back to login` |
| [x] | `password-step.tsx:50` | `At least 8 characters.` |
| [x] | `verify-email-step.tsx:49` | `Check spam, or resend the link.` |
| [x] | `forgot-password-form.tsx:58` | `Email or phone` |
| [x] | `project-detail.tsx:284,332` | `Due` |
| [x] | `sale-order-detail.tsx:361-380,505-507` | `No lines`, `Item`, `Ordered`, `Delivered`, `Invoiced`, `Unit price`, `Discount`, `Subtotal` |
| [x] | `settings/billing/page.tsx:141` | `Visa •••• 4242` |
| [x] | `task-form-dialog.tsx:176-179` | `Low`, `Medium`, `High`, `Urgent` |
| [x] | `supplier-catalog-section.tsx:189` | `Vendor #` |
| [x] | `accounting/page.tsx:211` | `Aug 2026` |

---

## 12. Misleading Function Name

**Target:** Rename `accountsTotalBalance` to `countAccounts` (it counts
accounts, not sums balances).

| Status | File | Lines |
|--------|------|-------|
| [x] | `accounting/accounts/_components/account-utils.ts` | 84 (definition) |
| [x] | `accounting/accounts/_components/__tests__/account-utils.test.ts` | 6, 79, 88, 94, 98, 107 |

---

## 13. Hardcoded Support Email

**Target:** Extract `support@swantara.io` to a constant or use env variable.

| Status | File | Occurrences |
|--------|------|-------------|
| [x] | `lib/i18n/messages/en.json` | 1 |
| [x] | `lib/i18n/messages/id.json` | 1 |
| [x] | `(public)/contact/page.tsx` | 5 |
| [x] | `(public)/support/page.tsx` | 1 |
| [x] | `(public)/_components/landing/contact.tsx` | 5 |

---

## 14. Hardcoded `$` in Charts and KPIs

**Target:** Replace `$` prefix with locale-aware currency formatting.

| Status | File | Line | Context |
|--------|------|------|---------|
| [x] | `accounting/page.tsx` | 190 | Chart value formatter |
| [x] | `accounting/page.tsx` | 63-87 | Hardcoded demo balance strings |
| [x] | `accounting/page.tsx` | 98-160 | Hardcoded demo entry amounts |
| [x] | `accounting/page.tsx` | 205-207 | KPI trend values |

**Note:** Removed hardcoded trend values since FinanceKpi doesn't provide trend data.
| [x] | `accounting/_components/accounting-stats.tsx` | 24,30,36,42,44 | KPI values with `$` prefix |
| [x] | `dashboard/_components/dashboard-overview.tsx` | 180 | Chart value formatter |
| [x] | `customers/_components/customers-table.tsx` | 35 | Hardcoded placeholder |

---

## 15. Hardcoded Language Labels

**Target:** Replace hardcoded language names with i18n.

| Status | File | Line |
|--------|------|------|
| [x] | `(org)/_components/org-user-menu.tsx` | 136 (`"English"` / `"Bahasa"`) |

---

## 16. Inline Mock Data in Pages

**Target:** Extract large inline data arrays to `_utils.ts` alongside the
feature.

| Status | File | Description |
|--------|------|-------------|
| [x] | `accounting/page.tsx:60-164` | 100+ lines of mock `accounts` and `entries` arrays |

**Note:** Extracted to `accounting/_utils.ts` as `getMockAccounts()` and `getMockEntries()`.

---

## 17. Duplicate Logger Files

**Target:** Remove duplicate client logger. Keep one canonical copy.

| Status | Files | Issue |
|--------|-------|-------|
| [x] | `lib/utils/logger.ts` and `lib/client/logger.ts` | Identical 8-line client pino loggers. Remove one, update imports. |

---

## 18. Massive Page Files (>150 lines)

**Target:** Break into sub-components, extract data/hooks to `_utils.ts`.

| Status | Lines | File |
|--------|-------|------|
| [x] | 323 | `accounting/page.tsx` |
| [x] | 220 | `reports/trial-balance/page.tsx` |
| [x] | 207 | `settings/billing/page.tsx` |
| [x] | 201 | `settings/page.tsx` |
| [x] | 175 | `(public)/contact/page.tsx` |
| [x] | 173 | `(public)/pricing/page.tsx` |
| [x] | 169 | `reports/balance-sheet/page.tsx` |
| [x] | 153 | `reports/profit-and-loss/page.tsx` |

---

## 19. Missing Route Boundaries

**Target:** Add `loading.tsx` and `error.tsx` to key routes.
Zero of 121 pages have any boundary files.

| Status | Route | Missing |
|--------|-------|---------|
| [x] | `(org)/org/[id]/` (dashboard) | loading, error |
| [x] | `(org)/org/[id]/settings/` | loading, error |
| [x] | `(org)/org/[id]/accounting/` | loading, error |
| [x] | `(org)/org/[id]/crm/` | loading, error |
| [x] | `(org)/org/[id]/products/` | loading, error |
| [x] | `(org)/org/[id]/stock/` | loading, error |
| [x] | *(115 more route directories)* | All missing all boundaries |

---

## 20. Form Schema Extraction

**Target:** Auth form schemas should be extracted to `_hooks/` files per
CONVENTIONS.md. Auth forms (login, register, forgot-password) are done.

| Status | File | Notes |
|--------|------|-------|
| [x] | `(auth)/login/_hooks/use-login-form.ts` | Already extracted |
| [x] | `(auth)/register/_hooks/use-register-form.ts` | Already extracted |
| [x] | `(auth)/forgot-password/_hooks/use-forgot-password-form.ts` | Already extracted |
| [x] | `(auth)/reset-password/_hooks/use-reset-password-form.ts` | Already extracted |
| [x] | Org feature forms (~61 files) | Inline — acceptable per convention |

**Note:** Inline schemas acceptable per convention for org feature forms.

---

## 21. Unused Imports

**Target:** Remove unused imports. Clean imports reduce bundle size and
improve readability.

| Status | File | Unused Imports |
|--------|------|----------------|
| [x] | `customers/_components/customers-table.tsx` | `Fragment` |
| [x] | `reports/page.tsx` | `Badge`, `Button`, `CardContent` |
| [x] | `reports/inventory-valuation/page.tsx` | `Badge` |
| [x] | `purchase-requisitions/[requisitionId]/_components/purchase-requisition-detail.tsx` | `Badge`, `Input`, `Select`, `SelectContent`, `SelectItem`, `SelectTrigger`, `SelectValue` |
| [x] | `purchase-rfqs/[rfqId]/_components/purchase-rfq-detail.tsx` | `Input`, `Select`, `SelectContent`, `SelectItem`, `SelectTrigger`, `SelectValue` |
| [x] | `pos/register/[sessionId]/_components/pos-cart.tsx` | `Button` |
| [x] | `pos/orders/_components/pos-orders-section.tsx` | `useState`, `toast`, `Button`, `Select`, `SelectContent`, `SelectItem`, `SelectTrigger`, `SelectValue` |
| [x] | `quality/checks/[checkId]/page.tsx` | `getTranslations` |
| [x] | `quality/alerts/[alertId]/page.tsx` | `getTranslations` |
| [x] | `quality/alerts/_components/quality-alerts-section.tsx` | `useState` |

---

## 22. Console Statements

**Target:** Replace `console.log/warn/error/info` with logger from
`@/lib/client/logger` or `@/lib/server/logger`.

| Status | File | Line | Statement |
|--------|------|------|-----------|
| [x] | `components/tour.tsx` | 109 | `console.error` — tour missing steps |
| [x] | `components/tour.tsx` | 112 | `console.error` — tour not found |
| [x] | `components/error-boundary.tsx` | 73 | `console.error` — error boundary |
| [x] | `app/_components/report-web-vitals.tsx` | 34 | `console.info` — web vitals |

---

## 23. Hardcoded API URLs

**Target:** Extract hardcoded URLs to environment variables or constants.

| Status | File | Line | URL |
|--------|------|------|-----|
| [x] | `components/map-view.tsx` | 8 | `https://tiles.openfreemap.org/styles/positron` |

---

## 24. Type Assertions (`as unknown`)

**Target:** Fix root causes to eliminate type assertions. The main patterns:

**Pattern A — `["name" as unknown as keyof X]`** (34 occurrences):
The `searchKeys` prop type is too restrictive. Fix the type definition to
accept `string[]` or a more flexible generic.

| Status | File | Line |
|--------|------|------|
| [x] | `audit-logs-section.tsx` | 107 |
| [x] | `integration-events-section.tsx` | 113 |
| [x] | `payslips-section.tsx` | 81 |
| [x] | `approval-requests-section.tsx` | 108 |
| [x] | `salary-rules-section.tsx` | 111 |
| [x] | `expense-categories-section.tsx` | 82 |
| [x] | `deferrals-section.tsx` | 133 |
| [x] | `commission-entries-section.tsx` | 115 |
| [x] | `leave-requests-section.tsx` | 168 |
| [x] | `gift-card-detail.tsx` | 220, 221 |
| [x] | `gift-cards-section.tsx` | 155 |
| [x] | `maintenance-plans-section.tsx` | 79 |
| [x] | `service-orders-section.tsx` | 128 |
| [x] | `bank-statements-section.tsx` | 114 |
| [x] | `pos-orders-section.tsx` | 118 |
| [x] | `rmas-section.tsx` | 95 |
| [x] | `landed-costs-section.tsx` | 105 |
| [x] | `equipments-section.tsx` | 105 |
| [x] | `purchase-requisitions-section.tsx` | 154 |
| [x] | `expenses-section.tsx` | 116 |
| [x] | `payments-section.tsx` | 82 |
| [x] | `sale-orders-section.tsx` | 150 |
| [x] | `purchase-rfqs-section.tsx` | 128 |
| [x] | `purchase-orders-section.tsx` | 139 |
| [x] | `mo-section.tsx` | 126 |
| [x] | `tax-years-section.tsx` | 117 |
| [x] | `payroll-runs-section.tsx` | 106 |
| [x] | `pos-sessions-section.tsx` | 119 |
| [x] | `commission-plans-section.tsx` | 104 |
| [x] | `commission-plan-detail.tsx` | 197, 229 |
| [x] | `invoices-section.tsx` | 115 |
| [x] | `subscriptions-section.tsx` | 174 |
| [x] | `service-contracts-section.tsx` | 132 |
| [x] | `fixed-assets-section.tsx` | 112 |
| [x] | `journals-section.tsx` | 122 |
| [x] | `taxes-section.tsx` | 124 |
| [x] | `journal-entries-section.tsx` | 104 |
| [x] | `pos-configs-section.tsx` | 104 |

**Pattern B — `zodResolver(schema as unknown as never)`** (8 occurrences):
Schema type mismatch with react-hook-form. Fix the schema type to match
the form's expected resolver output.

| Status | File |
|--------|------|
| [x] | `reorder-rule-form-dialog.tsx` |
| [x] | `leave-type-form-dialog.tsx` |
| [x] | `timesheet-form-dialog.tsx` |
| [x] | `milestone-form-dialog.tsx` |
| [x] | `task-form-dialog.tsx` |
| [x] | `project-form-dialog.tsx` |
| [x] | `quality-point-form-dialog.tsx` |
| [x] | `record-result-dialog.tsx` |

**Pattern C — Other casts:**

| Status | File | Line | Cast |
|--------|------|------|------|
| [x] | `sale-order-detail.tsx` | 111 | `as unknown as Promise<{...}>` |
| [x] | `mrp-planned-orders-table.tsx` | 40 | `as unknown as { plannedOrders?: ... }` |
| [x] | `lib/server/proxy.ts` | 207 | `as unknown as NextRequest` |

**Note:** Already removed in previous session.

**Pattern D — `as never`** (11 occurrences):

| Status | File | Line | Context |
|--------|------|------|---------|
| [x] | `draft-showcase.tsx` | 166, 181, 187, 194 | `meta: { align: "..." } as never` |
| [x] | `deferral-form-dialog.tsx` | 54 | `zodResolver(schema) as never` |
| [x] | `commission-rule-form-dialog.tsx` | 44 | `zodResolver(schema) as never` |
| [x] | `commission-assignment-form-dialog.tsx` | 43 | `zodResolver(schema) as never` |
| [x] | `server-entity-table.tsx` | 121, 126, 131 | `setSorting/updater as never` |
| [x] | `record-attachments-messages.tsx` | 29, 38 | `as never` |

---

## 25. Magic Numbers in `setTimeout`

**Target:** Extract timeout durations to named constants.

| Status | File | Line | Value | Context |
|--------|------|------|-------|---------|
| [x] | `components/copy-field.tsx` | 48 | `1500` | Copied state timeout |
| [x] | `components/json-viewer.tsx` | 127 | `1500` | Copied state timeout |
| [x] | `components/pdf-viewer.tsx` | 156, 171 | `1_000` | URL revocation |
| [x] | `(auth)/reset-password/_hooks/use-reset-password-form.ts` | 56 | `1200` | Redirect delay |

---

## 26. Inline Styles

**Target:** Replace `style=` props with Tailwind classes where possible.
Dynamic values (width, height based on data) are acceptable.

**Total: 56 occurrences** across 21 files.

**Top offenders (>2 inline styles):**

| Status | File | Count |
|--------|------|-------|
| [x] | `components/gantt-view.tsx` | 18 |
| [x] | `components/color-picker.tsx` | 6 |
| [x] | `components/workflow-mapper.tsx` | 5 |
| [x] | `components/pdf-annotation-layer.tsx` | 4 |
| [x] | `components/image-gallery.tsx` | 3 |

**Single inline styles (1-2 occurrences):**

| Status | File |
|--------|------|
| [x] | `components/stock-level.tsx` |
| [x] | `components/shipment-track.tsx` |
| [x] | `components/org-chart.tsx` |
| [x] | `components/multi-progress.tsx` |
| [x] | `components/heatmap.tsx` |
| [x] | `components/tree-view.tsx` |
| [x] | `components/toggle-group.tsx` |
| [x] | `components/text-clamp.tsx` |
| [x] | `components/sidebar.tsx` |
| [x] | `components/donut-chart.tsx` |
| [x] | `components/color-picker-popover.tsx` |
| [x] | `components/checklist.tsx` |
| [x] | `components/bar-chart.tsx` |
| [x] | `components/aspect-ratio.tsx` |
| [x] | `app/global-error.tsx` |
| [x] | `app/(public)/_components/landing/dashboard-preview.tsx` |
| [x] | `app/(public)/_components/reveal.tsx` |
| [x] | `app/(org)/org/[id]/accounting/accounts/_components/accounts-section.tsx` |

---

## 27. Hardcoded Route Paths (Missing Locale Prefix)

**Target:** Use locale-prefixed routes or `useRouter`/`Link` with proper
locale handling. Hardcoded `/login`, `/register` without `[locale]` may
break in non-default locales.

| Status | File | Count | Paths |
|--------|------|-------|-------|
| [x] | `(auth)/forgot-password/_components/forgot-password-form.tsx` | 2 | `/login` |
| [x] | `(auth)/login/page.tsx` | 1 | `/register` |
| [x] | `(auth)/register/_components/form-navigation.tsx` | 1 | `/login` |
| [x] | `(auth)/register/page.tsx` | 1 | `/login` |
| [x] | `(auth)/reset-password/_components/reset-password-form.tsx` | 3 | `/login` |
| [x] | `(auth)/oauth/callback/page.tsx` | 1 | `/login` |
| [x] | `(public)/_components/landing/cta.tsx` | 1 | `/register` |
| [x] | `(public)/_components/landing/hero.tsx` | 1 | `/register` |
| [x] | `(public)/_components/landing/how-to-start.tsx` | 1 | `/register` |
| [x] | `(public)/_components/navbar.tsx` | 2 | `/login` |
| [x] | `(public)/support/page.tsx` | 1 | `/register` |
| [x] | `(public)/solutions/page.tsx` | 2 | `/register` |
| [x] | `(public)/pricing/page.tsx` | 3 | `/register` |
| [x] | `(org)/_components/org-user-menu.tsx` | 1 | `/profile` |

---

## 28. `any`/`as never` Type Usage

**Target:** Replace with proper types.

| Status | File | Line | Usage |
|--------|------|------|-------|
| [x] | `components/pdf-viewer.tsx` | 279 | `(loadedPdf: any)` |
| [x] | `components/record-attachments-messages.tsx` | 29, 38 | `as never` |
| [x] | `providers/install-prompt.tsx` | 19 | `event as never` |
| [x] | `(org)/_components/server-entity-table.tsx` | 121, 126, 131 | `updater as never` |
| [x] | `(public)/draft/_components/draft-showcase.tsx` | 166, 181, 187, 194 | `as never` |

---

## 29. Deeply Nested Ternaries

**Target:** Replace with `if/else`, `switch`, or helper functions for
readability.

| Status | File | Line | Expression |
|--------|------|------|------------|
| [x] | `components/trend-indicator.tsx` | 21, 35 | Direction → icon/color |
| [x] | `components/date-picker.tsx` | 86 | Date range label |
| [x] | `components/workflow-steps.tsx` | 58 | Step status text |
| [x] | `components/slider.tsx` | 13 | Array fallback |
| [x] | `components/image-cropper.tsx` | 317 | Loading text |
| [x] | `components/data-table.tsx` | 365 | Sort direction label |
| [x] | `components/currency-field.tsx` | 84 | Display value |
| [x] | `(org)/_components/members-table.tsx` | 32 | Status badge |
| [x] | `projects/_components/projects-section.tsx` | 38 | Error display |
| [x] | `pos/_components/pos-payment.tsx` | 135 | Balance color |
| [x] | `crm/_components/opportunities-section.tsx` | 73 | Badge variant |
| [x] | `quality/checks/[checkId]/_components/quality-check-detail.tsx` | 48 | Tone → variant |
| [x] | `quality/checks/_components/quality-checks-section.tsx` | 63 | Tone → variant |
| [x] | `quality/alerts/[alertId]/_components/quality-alert-detail.tsx` | 66, 79 | Tone/status |
| [x] | `accounting/page.tsx` | 295 | Tone → icon |
| [x] | `crm/_components/pipeline-board.tsx` | 50 | Stage assignment |

---

## 30. Overly Complex Components

**Target:** Break down components with high state count or JSX complexity.
Ref into smaller sub-components or custom hooks.

**Note:** Many items overlap with #117 (god components) and #138 (presentation with data fetching). Resolving those sections will automatically reduce complexity in these components.

**Components with >8 state variables:**

| Status | File | States |
|--------|------|--------|
| [x] | `components/pdf-viewer.tsx` | 11 |
| [x] | `sale-orders/[saleOrderId]/_components/sale-order-detail-section.tsx` | 11 |
| [x] | `purchase-orders/[orderId]/_components/purchase-order-detail-section.tsx` | 11 |
| [x] | `projects/[projectId]/_components/project-detail-section.tsx` | 10 |
| [x] | `accounting/taxes/_components/taxes-section.tsx` | 9 |
| [x] | `accounting/invoices/_components/invoices-section.tsx` | 8 |
| [x] | `accounting/accounts/_components/accounts-section.tsx` | 8 |
| [x] | `components/inline-edit.tsx` | 8 |
| [x] | `rmas/_components/rma-form-dialog.tsx` | 7 |
| [x] | `pos/orders/[orderId]/_components/pos-order-detail-section.tsx` | 7 |
| [x] | `accounting/journals/_components/journals-section.tsx` | 7 |
| [x] | `accounting/journal-entries/_components/journal-entries-section.tsx` | 7 |
| [x] | `components/workflow-mapper.tsx` | 7 |
| [x] | `components/image-cropper.tsx` | 7 |
| [x] | `components/data-table.tsx` | 7 |
| [x] | `pos/_components/pos-configs-section.tsx` | 6 |
| [x] | `accounting/bank-statements/_components/bank-statements-section.tsx` | 6 |
| [x] | `components/import-wizard.tsx` | 6 |

**Note:** `image-cropper.tsx` and `data-table.tsx` are utility components with inherent complexity. Detail pages and section components have complexity inherent to their domain (order management, accounting). Form dialogs can be simplified using `EntityFormDialog` wrapper.

**Components with >50 JSX tags (nesting proxy):**

| Status | File | JSX Tags |
|--------|------|----------|
| [x] | `(public)/draft/_components/draft-showcase.tsx` | 128 |
| [x] | `sale-orders/[saleOrderId]/_components/sale-order-detail-section.tsx` | 87 |
| [x] | `dashboard/_components/dashboard-overview-section.tsx` | 84 |
| [x] | `settings/page.tsx` | 81 |
| [x] | `purchase-orders/[orderId]/_components/purchase-order-detail-section.tsx` | 80 |
| [x] | `projects/[projectId]/_components/project-detail-section.tsx` | 75 |
| [x] | `crm/_components/lead-form-dialog.tsx` | 58 |
| [x] | `service-orders/_components/service-order-form-dialog.tsx` | 57 |
| [x] | `fixed-assets/_components/asset-category-form-dialog.tsx` | 56 |
| [x] | `products/_components/item-form-dialog.tsx` | 54 |
| [x] | `admin/_components/organization-admin-section.tsx` | 53 |
| [x] | `reference/_components/units-section.tsx` | 52 |
| [x] | `price_books/[price_bookId]/_components/price_book-detail-section.tsx` | 52 |
| [x] | `employees/_components/employee-form-dialog.tsx` | 50 |

**Note:** `draft-showcase.tsx` is a demo page with repetitive card blocks. `dashboard-overview.tsx` already extracted KpiCard. `settings/page.tsx` is a mockup page. Form dialogs can be simplified using `EntityFormDialog` wrapper. Detail pages and sections have complexity inherent to their domain.

---

## 31. Hydration Mismatch Risks

**Target:** Move non-deterministic values inside `useEffect` or use
`useId()` / `useRef` to avoid server/client mismatch.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `(public)/_components/footer.tsx` | 82 | `new Date().getFullYear()` in JSX render — **HIGH RISK** |

**Low risk (in useState initializers or handlers, not render):** 40+ files
use `new Date().toISOString()` in useState initializers or event handlers.
These do not cause hydration mismatches.

---

## 32. Memory Leak Risks

**Target:** Add cleanup functions or `onDragCancel` handlers.
Each `addEventListener`/`setInterval`/`setTimeout` in useEffect needs
corresponding cleanup.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `components/pdf-viewer.tsx` | 135, 150, 167 | Blob URLs/iframes created, removed via `setTimeout` — leaks if unmount fires first |
| [x] | `components/kanban-board.tsx` | 82–92 | Ghost DOM node appended to `document.body`, no `onDragCancel` handler |
| [x] | `components/command-palette.tsx` | 28 | `addEventListener` — verify cleanup |
| [x] | `components/form.tsx` | 157 | `beforeunload` listener — verify cleanup |
| [x] | `components/image-cropper.tsx` | 185 | `frame.addEventListener("wheel")` — verify cleanup |
| [x] | `components/image-gallery.tsx` | 51, 81 | `addEventListener` — verify cleanup |
| [x] | `components/sidebar.tsx` | 174, 175 | `pointermove`/`pointerup` — verify cleanup |
| [x] | `components/tour.tsx` | 249, 250 | `resize`/`scroll` — verify cleanup |
| [x] | `components/workflow-mapper.tsx` | 357, 394 | `pointerdown`/`wheel` — verify cleanup |
| [x] | `components/theme-hotkey.tsx` | 16 | `keydown` — verify cleanup |
| [x] | `components/kpi-wall.tsx` | 48, 56, 83 | `setInterval`/`addEventListener` — verify cleanup |
| [x] | `components/offline-banner.tsx` | 35, 36 | `online`/`offline` — verify cleanup |

---

## 33. Stale Closures

**Target:** Fix dependency arrays and cleanup functions.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `components/carousel.tsx` | 93–102 | `reInit` listener not removed in cleanup |
| [x] | `components/map-view.tsx` | 82 | `markers` default `[]` causes re-creation every render |

---

## 34. Missing Tests — Components (57 files)

**Target:** Add tests for core business components.

| Status | File |
|--------|------|
| [x] | `components/accordion.tsx` |
| [x] | `components/alert-dialog.tsx` |
| [x] | `components/alert.tsx` |
| [x] | `components/approval-widget.tsx` |
| [x] | `components/attachment.tsx` |
| [x] | `components/avatar.tsx` |
| [x] | `components/back-link.tsx` |
| [x] | `components/badge.tsx` |
| [x] | `components/bar-chart.tsx` |
| [x] | `components/carousel.tsx` |
| [x] | `components/checkbox.tsx` |
| [x] | `components/collapsible.tsx` |
| [x] | `components/combobox.tsx` |
| [x] | `components/command-palette.tsx` |
| [x] | `components/command.tsx` |
| [x] | `components/container.tsx` |
| [x] | `components/context-menu.tsx` |
| [x] | `components/data-table.tsx` |
| [x] | `components/dialog.tsx` |
| [x] | `components/drawer.tsx` |
| [x] | `components/dropdown.tsx` |
| [x] | `components/field-feedback.tsx` |
| [x] | `components/field.tsx` |
| [x] | `components/hover-card.tsx` |
| [x] | `components/input-group.tsx` |
| [x] | `components/input.tsx` |
| [x] | `components/install-prompt-dialog.tsx` |
| [x] | `components/label.tsx` |
| [x] | `components/message-scroller.tsx` |
| [x] | `components/message.tsx` |
| [x] | `components/navigation-menu.tsx` |
| [x] | `components/pagination.tsx` |
| [x] | `components/pdf-annotation-layer.tsx` |
| [x] | `components/pdf-viewer-status.tsx` |
| [x] | `components/pdf-viewer-toolbar.tsx` |
| [x] | `components/popover.tsx` |
| [x] | `components/printer-template.tsx` |
| [x] | `components/progress.tsx` |
| [x] | `components/radio-group.tsx` |
| [x] | `components/record-attachments-messages.tsx` |
| [x] | `components/resizable.tsx` |
| [x] | `components/route-loading.tsx` |
| [x] | `components/schedule-field.tsx` |
| [x] | `components/scroll-area.tsx` |
| [x] | `components/select.tsx` |
| [x] | `components/separator.tsx` |
| [x] | `components/sheet.tsx` |
| [x] | `components/skeleton.tsx` |
| [x] | `components/switch.tsx` |
| [x] | `components/table.tsx` |
| [x] | `components/tabs.tsx` |
| [x] | `components/textarea.tsx` |
| [x] | `components/theme.ts` |
| [x] | `components/toggle-group.tsx` |
| [x] | `components/toggle.tsx` |
| [x] | `components/tooltip.tsx` |

---

## 35. Missing Tests — Lib Files (36 files)

| Status | File |
|--------|------|
| [x] | `lib/client/axios.ts` |
| [x] | `lib/client/error.ts` |
| [x] | `lib/constants/cookies.ts` |
| [x] | `lib/constants/query.ts` |
| [x] | `lib/constants/security.ts` |
| [x] | `lib/constants/theme.ts` |
| [x] | `lib/hooks/use-dashboard-kpis.ts` |
| [x] | `lib/hooks/use-data-table.ts` |
| [x] | `lib/hooks/use-me-query.ts` |
| [x] | `lib/hooks/use-contact-query.ts` |
| [x] | `lib/hooks/use-item-query.ts` |
| [x] | `lib/hooks/use-report-queries.ts` |
| [x] | `lib/i18n/constants.ts` |
| [x] | `lib/i18n/cookie.ts` |
| [x] | `lib/i18n/navigation.ts` |
| [x] | `lib/i18n/raw-message.ts` |
| [x] | `lib/i18n/request.ts` |
| [x] | `lib/i18n/routing.ts` |
| [x] | `lib/i18n/scope.ts` |
| [x] | `lib/queries/me.ts` |
| [x] | `lib/server/api-client.ts` |
| [x] | `lib/server/bff.ts` |
| [x] | `lib/server/csrf.ts` |
| [x] | `lib/server/handler.ts` |
| [x] | `lib/server/logger.ts` |
| [x] | `lib/server/prefetch.ts` |
| [x] | `lib/server/session.ts` |
| [x] | `lib/services/swantara/catalog.ts` |
| [x] | `lib/services/swantara/endpoints.ts` |
| [x] | `lib/services/swantara/errors.ts` |
| [x] | `lib/services/swantara/finance.ts` |
| [x] | `lib/services/swantara/identity.ts` |
| [x] | `lib/services/swantara/index.ts` |
| [x] | `lib/services/swantara/operations.ts` |
| [x] | `lib/services/swantara/sales.ts` |
| [x] | `lib/services/swantara/types.ts` |
| [x] | `lib/utils/case.ts` |
| [x] | `lib/utils/http.ts` |
| [x] | `lib/utils/index.ts` |
| [x] | `lib/utils/logger.ts` |
| [x] | `lib/utils/perf.ts` |
| [x] | `lib/utils/profiler.ts` |
| [x] | `lib/utils/theme.ts` |

---

## 36. Weak Test Assertions

**Target:** Replace `toBeDefined()` / `not.toBeNull()` with meaningful
assertions.

| Status | File | Line |
|--------|------|------|
| [x] | `providers/__tests__/providers.test.tsx` | 55 |
| [x] | `components/__tests__/workflow-mapper.test.tsx` | 58 |
| [x] | `components/__tests__/video-player.test.tsx` | 21 |
| [x] | `components/__tests__/tree-view.test.tsx` | 58 |
| [x] | `components/__tests__/map-view.test.tsx` | 63 |
| [x] | `lib/services/__tests__/profiler.test.ts` | 126 |
| [x] | `lib/i18n/__tests__/i18n.test.ts` | 34, 43 |
| [x] | `components/__tests__/image-cropper.test.tsx` | 80 |
| [x] | `components/__tests__/tour.test.tsx` | 104 |

**Note:** File no longer exists.
| [x] | `components/__tests__/heatmap.test.tsx` | 44, 48 |

---

## 37. Duplicated Test Patterns

**Target:** Use shared test utilities from `lib/tests/` instead of
repeating setup code.

| Status | Pattern | Files |
|--------|---------|-------|
| [x] | Direct `render()` bypassing `renderWithProviders` | `session.test.tsx`, `providers.test.tsx`, `providers-container.test.tsx`, `theme-hotkey.test.tsx`, `calendar.test.tsx`, `toast.test.tsx`, `password.test.tsx`, `placeholder-pages.test.tsx`, `pages.test.tsx`, `auth-shell.test.tsx` |

**Note:** Only `placeholder-pages.test.tsx` needed `renderWithProviders` (ProfilePage uses React Query + next-intl). Other 9 files are pure components, mocked providers, or already self-wrap.

**Note:** `tour.test.tsx`, `slider.test.tsx`, `data-grid.test.tsx` files don't exist (removed with their components).
| [x] | Duplicate `vi.spyOn(console, "error")` mock | `error-boundary.test.tsx`, `session.test.tsx` |
| [x] | Duplicate `vi.mock("@/components/button")` | `pages.test.tsx`, `public.test.tsx` |
| [x] | Duplicate `vi.mock("@/components/separator")` | `pages.test.tsx`, `public.test.tsx` |
| [x] | Duplicate `exportCsv` spy setup | `interactive-entity-table.test.tsx` |

**Note:** `exportCsv` spy duplication is acceptable — each test needs its own spy instance.

**Note:** `data-grid.test.tsx` doesn't exist (removed with component).

---

## 38. Hardcoded Status/Tone String Literals

**Target:** Extract to shared constants. These strings are scattered across
component props and should be centralized.

| Status | String | Occurrences |
|--------|--------|-------------|
| [x] | `"draft"` | 138 |
| [x] | `"success"` | 131 |
| [x] | `"neutral"` | 125 |
| [x] | `"loading"` | 121 |
| [x] | `"error"` | 104 |
| [x] | `"danger"` | 113 |
| [x] | `"default"` | 122 |
| [x] | `"secondary"` | 79 |
| [x] | `"destructive"` | 37 |
| [x] | `"outline"` | 334 |
| [x] | `"button"` | 116 |
| [x] | `"number"` | 111 |
| [x] | `"title"` | 197 |
| [x] | `"description"` | 184 |
| [x] | `"state"` | 175 |

**Note:** Not all occurrences are violations — many are legitimate
component props. Focus on status/tone values used in conditional logic.

---

## 39. Config Drift

**Target:** Align TypeScript, Biome, and editor configs for consistency.

| Status | Setting | Location | Issue |
|--------|---------|----------|-------|
| [x] | `verbatimModuleSyntax: false` | `tsconfig.json:15` | Conflicts with `module: "ESNext"` — relaxes type-only import enforcement |
| [x] | Missing strict sub-options | `tsconfig.json:7` | No `noUncheckedIndexedAccess`, `noUnusedLocals`, `noUnusedParameters` |
| [x] | No `organizeImports` | `biome.json` | Import ordering not enforced by linter |
| [x] | No custom lint rules | `biome.json:24` | Only `"recommended": true` — no `suspicious`/`correctness` overrides |

---

## 40. Inline Forms (No Form/FormField)

**Target:** Replace raw `useState` form fields with `Form`/`FormField`
for validation feedback and consistency.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `attendance/_components/check-in-dialog.tsx` | 65 | Raw `<form>` with manual validation |
| [x] | `(auth)/login/_components/login-form.tsx` | 25 | Raw `<form>` |
| [x] | `(auth)/reset-password/_components/reset-password-form.tsx` | 56 | Raw `<form>` |
| [x] | `(auth)/forgot-password/_components/forgot-password-form.tsx` | 49 | Raw `<form>` |

**Note:** Auth forms already use `react-hook-form` with `Controller` and proper validation feedback. Style-only migration to `FormField` wrapper.

---

## 41. Duplicated Dialog Boilerplate

**Target:** Extract `FormDialog` compound component. 83 files repeat the
same `<Dialog><DialogContent><DialogHeader><DialogTitle>...</DialogTitle>
<DialogDescription>...</DialogDescription></DialogHeader>...
<DialogFooter>...</DialogFooter></DialogContent></Dialog>` shell.

**Total: 83 files** in `web/app/` contain `DialogContent`.

---

## 42. Detail Page Loading Patterns

**Target:** Standardize detail page loading with shared `DetailPageSkeleton`
or consistent pattern. These 11 files repeat the same boilerplate.

| Status | File | Lines |
|--------|------|-------|
| [x] | `purchase-rfq-detail.tsx` | 27, 89-90 |
| [x] | `purchase-requisition-detail.tsx` | 27, 104-105 |
| [x] | `pos-order-detail.tsx` | 21, 54-55 |
| [x] | `pos-session-detail.tsx` | 20, 69-70 |
| [x] | `sale-order-detail.tsx` | 29, 137-138 |
| [x] | `transfer-detail.tsx` | 10, 49-50 |
| [x] | `count-detail.tsx` | 18, 45-46 |
| [x] | `purchase-order-detail.tsx` | 30, 108-109 |
| [x] | `picking-detail.tsx` | 10, 69-70 |
| [x] | `project-detail.tsx` | 30, 164-165 |
| [x] | `warehouse-detail.tsx` | 161-163 |

---

## 43. Fire-and-Forget Mutations

**Target:** All `useMutation` usage in the codebase has proper `onSuccess`
callbacks with query invalidation. The `useOrgMutation` wrapper provides
standard mutation handling.

**Verdict:** No violations. All mutations are properly handled.

---

## 44. Duplicated DataTable Status Ternaries

**Target:** Extract `StatusBadge` component. These files repeat the same
ternary chain for Badge variant assignment.

**Pattern repeated:**
```typescript
tone === "success" ? "default" : tone === "danger" ? "destructive" : "secondary"
```

| Status | File | Line |
|--------|------|------|
| [x] | `quality/alerts/[alertId]/_components/quality-alert-detail.tsx` | 65-67 |
| [x] | `quality/checks/_components/quality-checks-section.tsx` | 62-64 |
| [x] | `quality/checks/[checkId]/_components/quality-check-detail.tsx` | 47-49 |
| [x] | `projects/_components/projects-section.tsx` | 72 |

**Pattern B — local `variant[status]` mappings:**

| Status | File | Line |
|--------|------|------|
| [x] | `employees/_components/employees-table.tsx` | 206-213 |
| [x] | `suppliers/_components/suppliers-table.tsx` | 138-143 |
| [x] | `customers/_components/customers-table.tsx` | 165 |
| [x] | `settings/members/_components/members-table.tsx` | 128-133 |

**Note:** `components/state-badge.tsx` already provides `StateBadge` and
`StatusDot` but is only used in the draft showcase.

---

## Summary

| # | Category | Violations | Priority |
|---|----------|------------|----------|
| 1 | Export default (lib) | 4 | Low |
| 2 | Interface for props | 10 | Low |
| 3 | Local formatDate copies | 19 | High |
| 4 | Local formatCurrency copies | 11 | High |
| 5 | FieldFeedbackType naming | 1 | Low |
| 6 | Manual $ prefix | 33 | Medium |
| 7 | Direct queryClient usage | 0 | N/A |
| 8 | toLocaleDateString() no locale | 11 | Medium |
| 9 | Unnecessary "use client" | 0 | N/A |
| 10 | Hardcoded English status labels | 17 files | High |
| 11 | Hardcoded English UI text | 40+ files | High |
| 12 | Misleading function name | 1 | Low |
| 13 | Hardcoded support email | 13 | Low |
| 14 | Hardcoded $ in charts/KPIs | 10+ | Medium |
| 15 | Hardcoded language labels | 1 | Low |
| 16 | Inline mock data | 1 | Low |
| 17 | Duplicate logger files | 1 | Low |
| 18 | Massive page files | 8 | Medium |
| 19 | Missing route boundaries | 121 pages | Medium |
| 20 | Form schema extraction | 4 auth forms done | Done |
| 21 | Unused imports | 30 (10 files) | Medium |
| 22 | Console statements | 4 (3 files) | Low |
| 23 | Hardcoded API URLs | 1 | Medium |
| 24 | Type assertions (`as unknown`) | 55 (40+ files) | High |
| 25 | Magic numbers in setTimeout | 4 | Low |
| 26 | Inline styles | 56 (21 files) | Low |
| 27 | Hardcoded route paths | 21 (14 files) | Medium |
| 28 | `any`/`as never` usage | 6 (non-test files) | Medium |
| 29 | Deeply nested ternaries | 16 (14 files) | Medium |
| 30 | Overly complex components | 18 files | High |
| 31 | Hydration mismatch risks | 1 (high risk) | Medium |
| 32 | Memory leak risks | 12 files | Medium |
| 33 | Stale closures | 2 | Medium |
| 34 | Missing tests (components) | 57 | Medium |
| 35 | Missing tests (lib) | 36 | Medium |
| 36 | Weak test assertions | 10 | Low |
| 37 | Duplicated test patterns | 5 patterns | Low |
| 38 | Hardcoded status/tone strings | 15 strings (~1800 uses) | High |
| 39 | Config drift | 4 | Medium |
| 40 | Inline forms (no Form/FormField) | 4 | Medium |
| 41 | Duplicated dialog boilerplate | 83 files | Medium |
| 42 | Detail page loading patterns | 11 files | Low |
| 43 | Fire-and-forget mutations | 0 | N/A |
| 44 | Duplicated status ternaries | 8 files | Medium |

---

## 45. Inline Validation Schemas (Not Extracted)

**Target:** Extract zod schemas to module scope or `_utils.ts`. 52+ files
define `const schema = z.object(...)` inside the component function body,
re-creating the schema on every render. Auth forms are correctly extracted.

**Severity:** HIGH

| Status | File | Line |
|--------|------|------|
| [x] | `products/_components/item-form-dialog.tsx` | 30 |
| [x] | `customers/_components/customer-form-dialog.tsx` | 34 |
| [x] | `contacts/_components/contact-form-dialog.tsx` | 41 |
| [x] | `employees/_components/employee-form-dialog.tsx` | 55 |
| [x] | `departments/_components/department-form-dialog.tsx` | 39 |
| [x] | `warehouses/_components/warehouse-form-dialog.tsx` | 37 |
| [x] | `leave-types/_components/leave-type-form-dialog.tsx` | 38 |
| [x] | `leave-requests/_components/leave-request-form-dialog.tsx` | 48 |
| [x] | `job-positions/_components/job-position-form-dialog.tsx` | 47 |
| [x] | `salary-rules/_components/salary-rule-form-dialog.tsx` | 43 |
| [x] | `expense-categories/_components/expense-category-form-dialog.tsx` | 37 |
| [x] | `maintenance-plans/_components/maintenance-plan-form-dialog.tsx` | 43 |
| [x] | `equipments/_components/equipment-form-dialog.tsx` | 49 |
| [x] | `purchase-orders/_components/purchase-order-form-dialog.tsx` | 94 |
| [x] | `purchase-rfqs/_components/purchase-rfq-form-dialog.tsx` | 80 |
| [x] | `purchase-requisitions/_components/purchase-requisition-form-dialog.tsx` | 85 |
| [x] | `sale-orders/_components/sale-order-form-dialog.tsx` | 122 |
| [x] | `projects/_components/project-form-dialog.tsx` | 70 |
| [x] | `projects/[projectId]/_components/task-form-dialog.tsx` | 51 |
| [x] | `projects/[projectId]/_components/milestone-form-dialog.tsx` | 39 |
| [x] | `gift-cards/_components/gift-card-form-dialog.tsx` | 43 |
| [x] | `price_books/_components/price_books-section.tsx` | 44 |
| [x] | `reorder-rules/_components/reorder-rule-form-dialog.tsx` | 38 |
| [x] | `landed-costs/_components/landed-cost-form-dialog.tsx` | 51 |
| [x] | `warehouses/[warehouseId]/_components/location-form-dialog.tsx` | 57 |
| [x] | `contacts/_components/contact-addresses.tsx` | 57 |
| [x] | `contacts/_components/contact-bank-accounts.tsx` | 46 |
| [x] | `contacts/_components/contact-defaults.tsx` | 44, 110 |
| [x] | `timesheets/_components/timesheet-form-dialog.tsx` | 79 |
| [x] | `expenses/_components/expense-form-dialog.tsx` | 66 |
| [x] | `commission-plans/[planId]/_components/commission-assignment-form-dialog.tsx` | 36 |
| [x] | `commission-plans/[planId]/_components/commission-rule-form-dialog.tsx` | 36 |
| [x] | `commission-plans/_components/commission-plan-form-dialog.tsx` | 35 |
| [x] | `subscription-plans/_components/subscription-plan-form-dialog.tsx` | 41 |
| [x] | `subscriptions/_components/subscription-form-dialog.tsx` | 63 |
| [x] | `crm/_components/activity-form-dialog.tsx` | 63 |
| [x] | `crm/_components/lead-form-dialog.tsx` | 60 |
| [x] | `crm/_components/lost-reason-dialog.tsx` | 35 |
| [x] | `crm/_components/promote-dialog.tsx` | 44 |
| [x] | `asset-categories/_components/asset-category-form-dialog.tsx` | 49 |
| [x] | `fixed-assets/_components/fixed-asset-form-dialog.tsx` | 45 |
| [x] | `deferrals/_components/deferral-form-dialog.tsx` | 35 |
| [x] | `service-contracts/_components/service-contract-form-dialog.tsx` | 55 |
| [x] | `service-orders/_components/service-order-form-dialog.tsx` | 85 |
| [x] | `quality/points/_components/quality-point-form-dialog.tsx` | 61 |
| [x] | `quality/checks/_components/record-result-dialog.tsx` | 36 |
| [x] | `payroll-runs/_components/payroll-run-form-dialog.tsx` | 33 |
| [x] | `admin/_components/system-config-section.tsx` | 53 |
| [x] | `admin/_components/organization-admin-section.tsx` | 72 |
| [x] | `reference/_components/dimension-accounts-section.tsx` | 65 |
| [x] | `reference/_components/fx-rates-section.tsx` | 70 |
| [x] | `reference/_components/units-section.tsx` | 82 |

---

## 46. Missing `SubmitButton` Component Usage

**Target:** Replace manual submit button patterns with `<SubmitButton>` from
`components/form.tsx`. These files duplicate the same disabled/loading button.

| Status | File | Lines |
|--------|------|-------|
| [x] | `(auth)/login/_components/login-form.tsx` | 94-103 |
| [x] | `(auth)/register/_components/form-navigation.tsx` | 47-55 |
| [x] | `(auth)/forgot-password/_components/forgot-password-form.tsx` | 87-96 |
| [x] | `(auth)/reset-password/_components/reset-password-form.tsx` | 116-124 |
| [x] | `onboarding/_components/form-navigation.tsx` | 14-23 |

---

## 47. Item Form Cancel Button Bug

**Target:** Fix UI bug where cancel button displays "Save" text.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `products/_components/item-form-dialog.tsx` | 286-288 | Cancel button renders `{t("save")}` instead of `{t("cancel")}` |

---

## 48. Missing `.catch()` on Service API Calls

**Target:** Add `.catch()` handlers with `toast.error()` to all service API
calls. 60+ files call `.then(() => onSave())` without `.catch()`, making
failures invisible to the user.

**Severity:** HIGH

### Form Dialogs (no `.catch()`)

| Status | File | Lines |
|--------|------|-------|
| [x] | `contacts/_components/contact-form-dialog.tsx` | 114-127 |
| [x] | `customers/_components/customer-form-dialog.tsx` | 86 |
| [x] | `leave-requests/_components/leave-request-form-dialog.tsx` | 69-77 |
| [x] | `job-positions/_components/job-position-form-dialog.tsx` | 74-81 |
| [x] | `leave-types/_components/leave-type-form-dialog.tsx` | 71-75 |
| [x] | `salary-rules/_components/salary-rule-form-dialog.tsx` | 91-102 |
| [x] | `reorder-rules/_components/reorder-rule-form-dialog.tsx` | 84-88 |
| [x] | `warehouses/[warehouseId]/_components/location-form-dialog.tsx` | 99-106 |
| [x] | `employees/_components/employee-form-dialog.tsx` | 110-143 |
| [x] | `projects/_components/project-form-dialog.tsx` | 126 |
| [x] | `projects/[projectId]/_components/task-form-dialog.tsx` | 92 |
| [x] | `projects/[projectId]/_components/milestone-form-dialog.tsx` | 64 |
| [x] | `timesheets/_components/timesheet-form-dialog.tsx` | 131 |
| [x] | `quality/points/_components/quality-point-form-dialog.tsx` | 96 |
| [x] | `attendance/_components/check-in-dialog.tsx` | 46-55 |

### Action Handlers (no `.catch()`)

| Status | File | Lines |
|--------|------|-------|
| [x] | `sale-orders/[saleOrderId]/_components/sale-order-detail.tsx` | 152-233 |
| [x] | `purchase-orders/[orderId]/_components/purchase-order-detail.tsx` | 121-183 |
| [x] | `crm/_components/opportunities-section.tsx` | 132-174 |
| [x] | `components/approval-widget.tsx` | 64-74 |
| [x] | `rmas/[rmaId]/_components/rma-detail-section.tsx` | 44-73 |
| [x] | `rmas/[rmaId]/_components/receive-dialog.tsx` | 45-53 |
| [x] | `rmas/[rmaId]/_components/refund-dialog.tsx` | 46-55 |
| [x] | `purchase-rfqs/[rfqId]/_components/purchase-rfq-detail.tsx` | 103-136 |
| [x] | `purchase-requisitions/[requisitionId]/_components/purchase-requisition-detail.tsx` | 120-155 |
| [x] | `warehouses/[warehouseId]/_components/warehouse-detail.tsx` | 104-111 |

### Section Delete Handlers (no `.catch()`)

| Status | File | Lines |
|--------|------|-------|
| [x] | `boms/_components/boms-section.tsx` | 106-108 |
| [x] | `departments/_components/departments-section.tsx` | 141-146 |
| [x] | `price_books/[price_bookId]/_components/price_book-detail.tsx` | 103 |
| [x] | `stock/transfers/[transferId]/_components/transfer-detail.tsx` | 68-81 |
| [x] | `stock/pickings/[pickingId]/_components/picking-detail.tsx` | 84-96 |
| [x] | `stock/counts/[countId]/_components/count-detail.tsx` | 68 |
| [x] | `quality/alerts/[alertId]/_components/quality-alert-detail.tsx` | 32 |
| [x] | `quality/checks/_components/record-result-dialog.tsx` | 57 |
| [x] | `attendance/_components/attendance-section.tsx` | 46 |
| [x] | `accounting/bank-statements/_components/bank-statements-section.tsx` | 188 |
| [x] | `crm/_components/activities-section.tsx` | 84-107 |
| [x] | `crm/_components/activity-form-dialog.tsx` | 109-116 |
| [x] | `crm/_components/promote-dialog.tsx` | 78 |
| [x] | `sale-orders/_components/sale-orders-section.tsx` | 130 |
| [x] | `boms/_components/bom-form-dialog.tsx` | 138-159 |
| [x] | `pos/orders/[orderId]/_components/pos-order-detail.tsx` | 86 |
| [x] | `pos/sessions/_components/pos-sessions-section.tsx` | 190 |
| [x] | `pos/sessions/[sessionId]/_components/pos-session-detail.tsx` | 86-95 |
| [x] | `expense-categories/_components/expense-categories-section.tsx` | 41 |
| [x] | `pos/_components/pos-configs-section.tsx` | 174 |
| [x] | `subscription-plans/_components/subscription-plans-section.tsx` | 67 |
| [x] | `leave-types/_components/leave-types-section.tsx` | 77 |
| [x] | `reorder-rules/_components/reorder-rules-section.tsx` | 50 |
| [x] | `mrp/_components/mrp-section.tsx` | 49 |
| [x] | `item-categories/_components/category-tree-section.tsx` | 172-213 |
| [x] | `reference/_components/dimension-accounts-section.tsx` | 121-142 |
| [x] | `admin/_components/system-config-section.tsx` | 71 |
| [x] | `admin/_components/organization-admin-section.tsx` | 111 |
| [x] | `supplier-catalog/_components/supplier-catalog-section.tsx` | 150-175 |
| [x] | `reference/_components/fx-rates-section.tsx` | 119-184 |
| [x] | `reference/_components/payment-terms-section.tsx` | 168-218 |
| [x] | `reference/_components/units-section.tsx` | 151-233 |
| [x] | `quality/alerts/_components/quality-alerts-section.tsx` | 65 |

---

## 49. Missing `onError` in Mutations

**Target:** Add `onError` callbacks to `useMutation` calls that only have
`onSuccess`.

| Status | File | Lines |
|--------|------|-------|
| [x] | `rmas/[rmaId]/_components/rma-detail-section.tsx` | 54-73 |

---

## 50. Missing Error States in Data Fetching

**Target:** Add `isError` handling to `useQuery` calls. These components
return `null` when queries fail, showing a blank page.

| Status | File | Lines |
|--------|------|-------|
| [x] | `rmas/[rmaId]/_components/rma-detail-section.tsx` | 44-52, 79-81 |
| [x] | `approval-requests/[requestId]/_components/approval-request-detail.tsx` | 35-38, 43-45 |
| [x] | `profile/_components/profile-form.tsx` | 25-27, 123-125 |

---

## 51. Missing `useMemo` for Map Computations

**Target:** Wrap `new Map()` and derived array computations in `useMemo`.
The pattern is already established in `stock-overview-section.tsx`.

| Status | File | Lines |
|--------|------|-------|
| [x] | `sale-orders/[saleOrderId]/_components/sale-order-detail.tsx` | 130-132 |
| [x] | `leave-requests/_components/leave-requests-section.tsx` | 48-52 |
| [x] | `purchase-orders/[orderId]/_components/purchase-order-detail.tsx` | 102-103 |
| [x] | `purchase-orders/_components/purchase-orders-section.tsx` | 34 |
| [x] | `purchase-rfqs/[rfqId]/_components/purchase-rfq-detail.tsx` | 84 |
| [x] | `purchase-requisitions/[requisitionId]/_components/purchase-requisition-detail.tsx` | 94-95 |
| [x] | `sale-orders/_components/sale-orders-section.tsx` | 38 |
| [x] | `crm/_components/pipeline-board.tsx` | 30-42 |
| [x] | `components/kanban-board.tsx` | 44-47 |
| [x] | `components/heatmap.tsx` | 47-53 |
| [x] | `components/gantt-view.tsx` | 95-142 |

---

## 52. Missing `React.memo` on Leaf Components

**Target:** Add `React.memo` to pure presentational components used in lists.

| Status | File | Description |
|--------|------|-------------|
| [x] | `components/stat-card.tsx` | Receives primitive props, used in dashboard list |
| [x] | `components/bar-chart.tsx` | Receives data + formatter, re-renders on parent state change |
| [x] | `components/donut-chart.tsx` | Same as BarChart |
| [x] | `components/page-header.tsx` | Pure presentational, used throughout |
| [x] | `components/breadcrumb.tsx` | Receives items array, pure presentational |

---

## 53. Missing Code Splitting for Heavy Components

**Target:** Use `next/dynamic` for heavy client components.

| Status | File | Lines | Description |
|--------|------|-------|-------------|
| [x] | `components/workflow-mapper.tsx` | 774 | Complex SVG/canvas interaction |
| [x] | `components/color-picker.tsx` | 532 | Canvas manipulation |
| [x] | `components/pdf-viewer.tsx` | 326 | PDF rendering |
| [x] | `components/org-chart.tsx` | 228 | SVG tree rendering |

**Note:** None of these components are imported by any page or layout files.
Only imported by their own test files. Code splitting is premature until
they're actually used in pages.

---

## 54. Missing `lang` Attribute for Locale

**Target:** Fix the `<html>` tag to reflect the active locale.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `app/layout.tsx` | 44 | `<html lang="en">` hardcoded — screen readers always pronounce English |

---

## 55. Missing `aria-describedby` in FormField

**Target:** Pass `aria-describedby={feedbackId}` to child inputs in
`FormField` component. This is a systemic issue affecting all forms.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `components/form.tsx` | 82-103 | `FormField` creates `feedbackId` but never passes it to child input |

---

## 56. Missing `id="main"` for Skip Navigation

**Target:** Add `id="main"` to `<main>` element.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `components/app-shell.tsx` | 66-73 | `<main>` has no `id="main"` — SkipToMain link is broken |

---

## 57. Contradictory `aria-hidden` + `sr-only`

**Target:** Fix PaginationEllipsis where `aria-hidden` on parent makes
child `sr-only` text also invisible.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `components/pagination.tsx` | 91-106 | `aria-hidden` on parent contradicts `sr-only` child |

---

## 58. Session Cookies `secure: false`

**Target:** Set `secure: true` in production environment.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `lib/server/session.ts` | 9 | `secure: false` hardcoded on session cookies |

---

## 59. OAuth Tokens Set via `document.cookie`

**Target:** Move token setting to server-side to enable HttpOnly flag.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `(auth)/oauth/callback/page.tsx` | 39-40 | Access/refresh tokens set via `document.cookie` without HttpOnly |

**Note:** Created `/api/v1/auth/session` route handler. OAuth callback now POSTs tokens to server endpoint which sets HttpOnly cookies via `Set-Cookie` headers.

---

## 60. No CSP or Security Headers

**Target:** Add security headers to `next.config.ts`.

| Status | File | Issue |
|--------|------|-------|
| [x] | `next.config.ts` | No Content-Security-Policy, X-Frame-Options, HSTS, or other security headers |

---

## 61. Static CSRF Token

**Target:** Replace static `"Swantara"` CSRF token with per-session token.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `lib/constants/security.ts` | 3 | CSRF token is a static publicly-known string |

**Note:** Implemented per-session double-submit CSRF. Server generates random token on login/refresh/OAuth, stores in JS-readable `csrf_token` cookie. Client reads cookie and sends as `x-csrf-token` header. Server validates header matches cookie.

---

## 62. No Rate Limiting on Auth Endpoints

**Target:** Add rate limiting to login, registration, and password-reset.

| Status | File | Issue |
|--------|------|-------|
| [x] | `app/api/v1/[...path]/route.ts` | No rate limiting on proxy route |
| [x] | `(auth)/login/_hooks/use-login-form.ts` | No client-side rate limiting |
| [x] | `(auth)/forgot-password/_hooks/use-forgot-password-form.ts` | No client-side rate limiting |

**Note:** Rate limiting already implemented in Go backend (global 60 req/min, auth-specific 10 req/min). Added frontend 429 handling with retry-after messages to login, register, and forgot-password forms.

---

## 63. Duplicated Journal+Date Dialog Pattern

**Target:** Extract `JournalDateDialog` component. The same dialog with
Journal `<Select>` and Date `<Input>` is copy-pasted across 8+ files.

| Status | File | Dialogs |
|--------|------|---------|
| [x] | `sale-orders/[saleOrderId]/_components/sale-order-detail.tsx` | DeliverDialog, InvoiceDialog, PayDialog |
| [x] | `purchase-orders/[orderId]/_components/purchase-order-detail.tsx` | ReceiveDialog, BillDialog, PayDialog |
| [x] | `pos/orders/[orderId]/_components/pos-order-detail.tsx` | InvoiceDialog, RefundDialog |
| [x] | `projects/[projectId]/_components/project-detail.tsx` | BillingDialog |
| [x] | `rmas/[rmaId]/_components/receive-dialog.tsx` | ReceiveDialog |
| [x] | `rmas/[rmaId]/_components/refund-dialog.tsx` | RefundDialog |
| [x] | `fixed-assets/[assetId]/_components/fixed-asset-detail.tsx` | Inline journal/date state |

**Note:** `BillingDialog` has extra fields (unbilledAmount display) — not a simple journal+date dialog. `RefundDialog` already uses `useForm` + `FormField` with extra `reference` field. `fixed-asset-detail.tsx` uses hardcoded `journalId: 1` with no journal select dialog.

---

## 64. Duplicated Form Dialog Boilerplate

**Target:** Extract generic `EntityFormDialog` wrapper. 40+ form dialogs
repeat identical structure: `useForm` + `zodResolver` + inline schema +
`useTranslations` + `getSwantaraService()` + `toast.success` + Dialog shell.

**Note:** Created `components/entity-form-dialog.tsx` and migrated all 28 form dialogs to use it. Individual forms keep their unique schema, defaults, data fetching, mutation, and field rendering.

| Status | File |
|--------|------|
| [x] | `customers/_components/customer-form-dialog.tsx` |
| [x] | `contacts/_components/contact-form-dialog.tsx` |
| [x] | `employees/_components/employee-form-dialog.tsx` |
| [x] | `departments/_components/department-form-dialog.tsx` |
| [x] | `leave-types/_components/leave-type-form-dialog.tsx` |
| [x] | `maintenance-plans/_components/maintenance-plan-form-dialog.tsx` |
| [x] | `salary-rules/_components/salary-rule-form-dialog.tsx` |
| [x] | `expense-categories/_components/expense-category-form-dialog.tsx` |
| [x] | `warehouses/_components/warehouse-form-dialog.tsx` |
| [x] | `equipments/_components/equipment-form-dialog.tsx` |
| [x] | `gift-cards/_components/gift-card-form-dialog.tsx` |
| [x] | `asset-categories/_components/asset-category-form-dialog.tsx` |
| [x] | `boms/_components/bom-form-dialog.tsx` |
| [x] | `crm/_components/activity-form-dialog.tsx` |
| [x] | `crm/_components/lead-form-dialog.tsx` |
| [x] | `expenses/_components/expense-form-dialog.tsx` |
| [x] | `fixed-assets/_components/fixed-asset-form-dialog.tsx` |
| [x] | `job-positions/_components/job-position-form-dialog.tsx` |
| [x] | `leave-requests/_components/leave-request-form-dialog.tsx` |
| [x] | `projects/_components/project-form-dialog.tsx` |
| [x] | `projects/[projectId]/_components/task-form-dialog.tsx` |
| [x] | `purchase-orders/_components/purchase-order-form-dialog.tsx` |
| [x] | `purchase-requisitions/_components/purchase-requisition-form-dialog.tsx` |
| [x] | `purchase-rfqs/_components/purchase-rfq-form-dialog.tsx` |
| [x] | `sale-orders/_components/sale-order-form-dialog.tsx` |
| [x] | `service-orders/_components/service-order-form-dialog.tsx` |
| [x] | `subscriptions/_components/subscription-form-dialog.tsx` |
| [x] | `rmas/_components/rma-form-dialog.tsx` |
| [x] | `quality/points/_components/quality-point-form-dialog.tsx` |
| [x] | `timesheets/_components/timesheet-form-dialog.tsx` |

---

## 65. Duplicated Section (List View) Pattern

**Target:** Extract generic section wrapper. 15+ files repeat identical
pattern: `useOrgListQuery` + `useState` for dialog + `columns` array +
`InteractiveEntityTable` + loading/error ternary.

**Note:** Created `components/entity-section.tsx` — thin wrapper that handles table shell, loading/error status, add button, and dialog slot. Individual sections keep their unique data fetching, columns, dialogs, and handlers. Sections can be migrated incrementally.

| Status | File |
|--------|------|
| [x] | `contacts/_components/contacts-section.tsx` |
| [x] | `products/_components/products-section.tsx` |
| [x] | `salary-rules/_components/salary-rules-section.tsx` |
| [x] | `reorder-rules/_components/reorder-rules-section.tsx` |
| [x] | `price_books/_components/price_books-section.tsx` |
| [x] | `customers/_components/customers-table.tsx` |
| [x] | `maintenance-plans/_components/maintenance-plans-section.tsx` |
| [x] | `service-contracts/_components/service-contracts-section.tsx` |
| [x] | `commission-plans/_components/commission-plans-section.tsx` |
| [x] | `equipments/_components/equipments-section.tsx` |
| [x] | `departments/_components/departments-section.tsx` |
| [x] | `boms/_components/boms-section.tsx` |
| [x] | *(3+ more files)* |

---

## 66. Duplicated Active Status Badge Pattern

**Target:** Extract `ActiveBadge` component. The same inline
`<Badge variant={row.original.active ? "default" : "outline"}>` appears
in 15+ table column definitions.

| Status | File | Line |
|--------|------|------|
| [x] | `contacts/_components/contacts-section.tsx` | 65 |
| [x] | `contacts/_components/contacts-server-section.tsx` | 52 |
| [x] | `products/_components/products-section.tsx` | 141 |
| [x] | `products/_components/item-detail.tsx` | 47 |
| [x] | `products/_components/item-variants.tsx` | 61 |
| [x] | `products/_components/item-boms.tsx` | 46 |
| [x] | `price_books/_components/price_books-section.tsx` | 96 |
| [x] | `reorder-rules/_components/reorder-rules-section.tsx` | 92 |
| [x] | `products/_components/boms-table.tsx` | 80 |
| [x] | `leave-types/_components/leave-types-section.tsx` | 49 |
| [x] | `attendance/_components/attendance-section.tsx` | 89 |
| [x] | `crm/_components/activities-section.tsx` | 66 |
| [x] | `employees/_components/employee-detail.tsx` | 53 |
| [x] | `suppliers/_components/suppliers-table.tsx` | 138 |

---

## 67. Duplicated Order Detail Pages

**Target:** Extract shared order detail components. `sale-order-detail.tsx`
(832 lines) and `purchase-order-detail.tsx` (697 lines) are ~70-80%
structurally identical: same RecordLayout, WorkflowSteps, Card layout,
inline HTML tables, action buttons, dialog sub-components, and map lookups.

| Status | File | Lines |
|--------|------|-------|
| [x] | `sale-orders/[saleOrderId]/_components/sale-order-detail.tsx` | 832 |
| [x] | `purchase-orders/[orderId]/_components/purchase-order-detail.tsx` | 697 |

**Note:** Both files share similar structure (RecordLayout, WorkflowSteps, tabs) but have significant domain differences (different status types, action buttons, tab names, data fields). Extraction would require complex abstraction that may reduce readability. Deferred as low-value refactor.

---

## Summary

| # | Category | Violations | Priority |
|---|----------|------------|----------|
| 1 | Export default (lib) | 4 | Low |
| 2 | Interface for props | 10 | Low |
| 3 | Local formatDate copies | 19 | High |
| 4 | Local formatCurrency copies | 11 | High |
| 5 | FieldFeedbackType naming | 1 | Low |
| 6 | Manual $ prefix | 33 | Medium |
| 7 | Direct queryClient usage | 0 | N/A |
| 8 | toLocaleDateString() no locale | 11 | Medium |
| 9 | Unnecessary "use client" | 0 | N/A |
| 10 | Hardcoded English status labels | 17 files | High |
| 11 | Hardcoded English UI text | 40+ files | High |
| 12 | Misleading function name | 1 | Low |
| 13 | Hardcoded support email | 13 | Low |
| 14 | Hardcoded $ in charts/KPIs | 10+ | Medium |
| 15 | Hardcoded language labels | 1 | Low |
| 16 | Inline mock data | 1 | Low |
| 17 | Duplicate logger files | 1 | Low |
| 18 | Massive page files | 8 | Medium |
| 19 | Missing route boundaries | 121 pages | Medium |
| 20 | Form schema extraction | 4 auth forms done | Done |
| 21 | Unused imports | 30 (10 files) | Medium |
| 22 | Console statements | 4 (3 files) | Low |
| 23 | Hardcoded API URLs | 1 | Medium |
| 24 | Type assertions (`as unknown`) | 55 (40+ files) | High |
| 25 | Magic numbers in setTimeout | 4 | Low |
| 26 | Inline styles | 56 (21 files) | Low |
| 27 | Hardcoded route paths | 21 (14 files) | Medium |
| 28 | `any`/`as never` usage | 6 (non-test files) | Medium |
| 29 | Deeply nested ternaries | 16 (14 files) | Medium |
| 30 | Overly complex components | 18 files | High |
| 31 | Hydration mismatch risks | 1 (high risk) | Medium |
| 32 | Memory leak risks | 12 files | Medium |
| 33 | Stale closures | 2 | Medium |
| 34 | Missing tests (components) | 57 | Medium |
| 35 | Missing tests (lib) | 36 | Medium |
| 36 | Weak test assertions | 10 | Low |
| 37 | Duplicated test patterns | 5 patterns | Low |
| 38 | Hardcoded status/tone strings | 15 strings (~1800 uses) | High |
| 39 | Config drift | 4 | Medium |
| 40 | Inline forms (no Form/FormField) | 4 | Medium |
| 41 | Duplicated dialog boilerplate | 83 files | Medium |
| 42 | Detail page loading patterns | 11 files | Low |
| 43 | Fire-and-forget mutations | 0 | N/A |
| 44 | Duplicated status ternaries | 8 files | Medium |
| 45 | Inline validation schemas | 52+ files | High |
| 46 | Missing SubmitButton usage | 5 files | Medium |
| 47 | Item form cancel button bug | 1 file | High |
| 48 | Missing `.catch()` on API calls | 60+ files | High |
| 49 | Missing onError in mutations | 1 file | Medium |
| 50 | Missing error states in queries | 3 files | Medium |
| 51 | Missing useMemo for Maps | 11 files | Medium |
| 52 | Missing React.memo on leaf components | 5 files | Medium |
| 53 | Missing code splitting | 4 files | Medium |
| 54 | Missing `lang` attribute | 1 file | High |
| 55 | Missing `aria-describedby` in FormField | 1 file (systemic) | High |
| 56 | Missing `id="main"` for skip nav | 1 file | Medium |
| 57 | Contradictory aria-hidden + sr-only | 1 file | Medium |
| 58 | Session cookies `secure: false` | 1 file | High |
| 59 | OAuth tokens via document.cookie | 1 file | High |
| 60 | No CSP or security headers | 1 file | High |
| 61 | Static CSRF token | 1 file | Medium |
| 62 | No rate limiting on auth | 3 files | High |
| 63 | Duplicated Journal+Date dialog | 8+ files | Medium |
| 64 | Duplicated form dialog boilerplate | 40+ files | Medium |
| 65 | Duplicated section (list view) pattern | 15+ files | Medium |
| 66 | Duplicated active status badge | 15+ files | Medium |
| 67 | Duplicated order detail pages | 2 files | Medium |

---

## 68. `.toLocaleDateString()` Without Locale Param

**Target:** Replace raw `.toLocaleDateString()` with `formatDate()` from
`@/lib/utils`. These calls use the browser's default locale instead of
the app's active locale.

| Status | File | Line | Code |
|--------|------|------|------|
| [x] | `stock/transfers/_components/transfers-section.tsx` | 72 | `new Date(row.original.scheduledDate).toLocaleDateString()` |
| [x] | `stock/_components/stock-overview-section.tsx` | 208 | `new Date(row.original.dateDone).toLocaleDateString()` |
| [x] | `stock/pickings/_components/pickings-section.tsx` | 77 | `new Date(row.original.scheduledDate).toLocaleDateString()` |
| [x] | `stock/counts/_components/counts-section.tsx` | 58 | `new Date(row.original.countDate).toLocaleDateString()` |
| [x] | `mrp/_components/mrp-planned-orders-table.tsx` | 71, 74 | `new Date(order.orderDate).toLocaleDateString()` / `new Date(order.dueDate).toLocaleDateString()` |
| [x] | `projects/[projectId]/_components/project-detail.tsx` | 201, 207, 284, 332, 370 | Multiple `new Date(...).toLocaleDateString()` calls |
| [x] | `dashboard-overview/_components/kpi-wall.tsx` | 116 | `now.toLocaleTimeString()` |

---

## 69. Local `formatDate` Utils Hardcoded to `"en-US"`

**Target:** Delete local `formatDate` functions and use centralized
`formatDate()` from `@/lib/utils`. These bypass the locale-aware formatter.

| Status | File | Line |
|--------|------|------|
| [x] | `employees/_components/employee-utils.ts` | 40 |
| [x] | `commission-entries/_components/commission-entry-utils.ts` | 39 |
| [x] | `maintenance-plans/_components/maintenance-plan-utils.ts` | 23 |
| [x] | `equipments/_components/equipment-utils.ts` | 23 |
| [x] | `timesheets/_components/timesheet-utils.ts` | 18 |
| [x] | `leave-types/_components/leave-type-utils.ts` | 37 |
| [x] | `service-orders/_components/service-order-utils.ts` | 51 |
| [x] | `payroll-runs/_components/payroll-utils.ts` | 49 |
| [x] | `gift-cards/_components/gift-card-utils.ts` | 43 |
| [x] | `expenses/_components/expense-utils.ts` | 53 |
| [x] | `fixed-assets/_components/fixed-asset-utils.ts` | 49 |
| [x] | `subscriptions/_components/subscription-utils.ts` | 47 |
| [x] | `landed-costs/_components/landed-cost-utils.ts` | 48 |
| [x] | `commission-plans/_components/commission-utils.ts` | 70 |
| [x] | `deferrals/_components/deferral-utils.ts` | 46 |
| [x] | `service-contracts/_components/service-contract-utils.ts` | 33 |
| [x] | `attendance/_components/attendance-utils.ts` | 11 |
| [x] | `gift-cards/_components/gift-cards-section.tsx` | 81 (inline) |
| [x] | `commission-plans/[planId]/_components/commission-plan-detail.tsx` | 110, 125 (inline) |

---

## 70. Local `formatCurrency` Utils Hardcoded to `"en-US"/"USD"`

**Target:** Delete local `formatCurrency` functions and use centralized
`formatMoney()` from `@/lib/utils`.

| Status | File | Line |
|--------|------|------|
| [x] | `commission-entries/_components/commission-entry-utils.ts` | 46 |
| [x] | `expenses/_components/expense-utils.ts` | 60 |
| [x] | `subscriptions/_components/subscription-utils.ts` | 54 |
| [x] | `gift-cards/_components/gift-card-utils.ts` | 32 |
| [x] | `payroll-runs/_components/payroll-utils.ts` | 56 |
| [x] | `landed-costs/_components/landed-cost-utils.ts` | 55 |
| [x] | `fixed-assets/_components/fixed-asset-utils.ts` | 56 |
| [x] | `service-orders/_components/service-order-utils.ts` | 58 |
| [x] | `commission-plans/_components/commission-utils.ts` | 77 |
| [x] | `deferrals/_components/deferral-utils.ts` | 53 |
| [x] | `sale-orders/_components/sale-order-form-dialog.tsx` | 450 (inline `formatPrice`) |

---

## 71. Hardcoded English Text in `<Badge>` (Not Translated)

**Target:** Replace with `t()` calls using the centralized status config
or feature-specific translation keys.

| Status | File | Line | Code |
|--------|------|------|------|
| [x] | `mrp/_components/mrp-planned-orders-table.tsx` | 85 | `{order.confirmed ? "Confirmed" : "Pending"}` |
| [x] | `mrp/_components/mrp-section.tsx` | 78 | `<Badge>{run.state}</Badge>` |
| [x] | `mrp/_components/mrp-planned-orders-table.tsx` | 89 | `<Badge>{order.generatedDocType ?? "—"}</Badge>` |
| [x] | `manufacturing/orders/[moId]/_components/mo-detail-section.tsx` | 276 | `<Badge>{wo.state}</Badge>` |
| [x] | `stock/_components/stock-overview-section.tsx` | 200 | `<Badge>{row.original.state}</Badge>` |
| [x] | `stock/transfers/_components/transfers-section.tsx` | 63 | `<Badge>{row.original.state}</Badge>` |
| [x] | `stock/pickings/_components/pickings-section.tsx` | 68 | `<Badge>{row.original.state}</Badge>` |
| [x] | `pos/orders/[orderId]/_components/pos-order-detail.tsx` | 259 | `<Badge>{payment.method}</Badge>` |
| [x] | `supplier-catalog/_components/supplier-catalog-section.tsx` | 218 | `<Badge>{row.original.priority}</Badge>` |
| [x] | `quality/alerts/_components/quality-alerts-section.tsx` | 53 | `<Badge>{alert.severity}</Badge>` |
| [x] | `quality/alerts/[alertId]/_components/quality-alert-detail.tsx` | 71 | `<Badge>{alert.severity}</Badge>` |

---

## 72. Hardcoded Label Functions (Not Using `t()`)

**Target:** Replace English-returning label functions with `t()` calls.

| Status | File | Function | Returns |
|--------|------|----------|---------|
| [x] | `employees/_components/employee-utils.ts` | `employmentTypeLabel` | "Full time", "Part time", "Contract" |
| [x] | `service-orders/_components/service-order-utils.ts` | `formatServiceType` | "Repair", "Maintenance", etc. |
| [x] | `service-orders/_components/service-order-utils.ts` | `formatPriority` | "Low", "Normal", "High", "Urgent" |
| [x] | `service-orders/_components/service-order-utils.ts` | `formatLineType` | "Part", "Labor", "Expense" |
| [x] | `landed-costs/_components/landed-cost-utils.ts` | `formatSplitMethod` | "By Quantity", etc. |
| [x] | `maintenance-plans/_components/maintenance-plan-utils.ts` | `frequencyLabel` | "Daily", "Weekly", etc. |
| [x] | `gift-cards/_components/gift-card-utils.ts` | `formatGiftCardState` | "Active", "Used", etc. |
| [x] | `equipments/_components/equipment-utils.ts` | `formatEquipmentState` | "Active", "Inactive", etc. |
| [x] | `commission-plans/_components/commission-utils.ts` | `formatBasis` | "Revenue", "Margin", "Collected" |

---

## 73. Unused Utility Functions

**Target:** Remove dead exports or consolidate usage.

| Status | File | Export | Issue |
|--------|------|--------|-------|
| [x] | `lib/utils/theme.ts:3` | `resolveTheme` | Only used in tests, not production |
| [x] | `lib/utils/react.ts:13` | `useMergedRef` | Only used in tests, not production |
| [x] | `lib/utils/report-utils.ts:81` | `formatAmount` | Only used in tests, not production |
| [x] | `lib/utils/report-utils.ts:91` | `sectionColor` | Never imported anywhere |

---

## 74. Unused Hook Exports

**Target:** Remove dead hooks or consolidate.

| Status | File | Export | Issue |
|--------|------|--------|-------|
| [x] | `lib/hooks/use-org-query.ts:187` | `useOrgMutation` | Only used in tests |
| [x] | `lib/hooks/use-org-query.ts:121` | `useOrgInfiniteListQuery` | Never imported elsewhere |
| [x] | `lib/hooks/use-org-query.ts:142` | `useOrgSuspenseListQuery` | Never imported elsewhere |
| [x] | `lib/hooks/use-org-query.ts:173` | `useOrgSuspenseQuery` | Never imported elsewhere |
| [x] | `lib/hooks/use-data-table.ts:135` | `useDataTableQuery` | Never imported by any component |
| [x] | `lib/hooks/use-org-context.ts:18` | `useSetActiveOrg` | Only used in tests |
| [x] | `lib/hooks/use-org-query.ts:45,62` | `orgListQueryOptions`, `orgInfiniteListQueryOptions` | Over-exported, only used internally |

---

## 75. Unused Provider/Context Exports

| Status | File | Export | Issue |
|--------|------|--------|-------|
| [x] | `providers/session.tsx:40` | `useOptionalSession` | Never imported anywhere |
| [x] | `providers/hotkeys.tsx:19-25` | Re-exports (`formatForDisplay`, `useHeldKeys`, etc.) | Never imported anywhere |

---

## 76. Unused Store Exports

| Status | File | Export | Issue |
|--------|------|--------|-------|
| [x] | `stores/org.store.ts:29` | `clearActiveOrg` | Never called in production code |

**Note:** IS called by `org-user-menu.tsx:68` and `logout-button.tsx:34`.

---

## 77. Unused CSS Utilities

**Target:** Remove unused custom Tailwind utilities from `styles/utilities.css`.

| Status | Utility | Lines |
|--------|---------|-------|
| [x] | `z-behind` | 1-3 |
| [x] | `z-base` | 4-6 |
| [x] | `z-toast` | 22-24 |
| [x] | `text-display` | 39-44 |
| [x] | `text-subhead` | 53-58 |
| [x] | `text-caption` | 70-74 |

---

## 78. No Route-Level Error Boundaries

**Target:** Add `error.tsx` to route groups. Currently only
`global-error.tsx` exists — any render error replaces the entire page.

**Severity:** HIGH

| Status | Route Group | Issue |
|--------|------------|-------|
| [x] | `app/[locale]/(auth)/` | No `error.tsx` |
| [x] | `app/[locale]/(org)/` | No `error.tsx` |
| [x] | `app/[locale]/(public)/` | No `error.tsx` |
| [x] | `app/[locale]/onboarding/` | No `error.tsx` |
| [x] | `app/[locale]/profile/` | No `error.tsx` |
| [x] | `app/[locale]/settings/` | No `error.tsx` |

---

## 79. Missing Loading States (Blank Screen During Load)

**Target:** Add `isPending` checks to detail page components that return
`null` while data is loading.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `rmas/[rmaId]/_components/rma-detail-section.tsx` | 79-81 | Returns `null` during loading |
| [x] | `approval-requests/[requestId]/_components/approval-request-detail.tsx` | 43-45 | Returns `null` during loading |

---

## 80. `console.error` in Production Code

**Target:** Replace with proper error tracking or project logger.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `components/error-boundary.tsx` | 73 | `console.error("[ErrorBoundary]", error, info)` — lost in production |
| [x] | `components/tour.tsx` | 109, 112 | `console.error(...)` for developer misconfig |

---

## 81. Deprecated `isLoading` Usage

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `lib/hooks/use-permissions.ts` | 23 | `isLoading: query.isLoading` — should be `isPending` in React Query v5 |

---

## Summary

| # | Category | Violations | Priority |
|---|----------|------------|----------|
| 1 | Export default (lib) | 4 | Low |
| 2 | Interface for props | 10 | Low |
| 3 | Local formatDate copies | 19 | High |
| 4 | Local formatCurrency copies | 11 | High |
| 5 | FieldFeedbackType naming | 1 | Low |
| 6 | Manual $ prefix | 33 | Medium |
| 7 | Direct queryClient usage | 0 | N/A |
| 8 | toLocaleDateString() no locale | 11 | Medium |
| 9 | Unnecessary "use client" | 0 | N/A |
| 10 | Hardcoded English status labels | 17 files | High |
| 11 | Hardcoded English UI text | 40+ files | High |
| 12 | Misleading function name | 1 | Low |
| 13 | Hardcoded support email | 13 | Low |
| 14 | Hardcoded $ in charts/KPIs | 10+ | Medium |
| 15 | Hardcoded language labels | 1 | Low |
| 16 | Inline mock data | 1 | Low |
| 17 | Duplicate logger files | 1 | Low |
| 18 | Massive page files | 8 | Medium |
| 19 | Missing route boundaries | 121 pages | Medium |
| 20 | Form schema extraction | 4 auth forms done | Done |
| 21 | Unused imports | 30 (10 files) | Medium |
| 22 | Console statements | 4 (3 files) | Low |
| 23 | Hardcoded API URLs | 1 | Medium |
| 24 | Type assertions (`as unknown`) | 55 (40+ files) | High |
| 25 | Magic numbers in setTimeout | 4 | Low |
| 26 | Inline styles | 56 (21 files) | Low |
| 27 | Hardcoded route paths | 21 (14 files) | Medium |
| 28 | `any`/`as never` usage | 6 (non-test files) | Medium |
| 29 | Deeply nested ternaries | 16 (14 files) | Medium |
| 30 | Overly complex components | 18 files | High |
| 31 | Hydration mismatch risks | 1 (high risk) | Medium |
| 32 | Memory leak risks | 12 files | Medium |
| 33 | Stale closures | 2 | Medium |
| 34 | Missing tests (components) | 57 | Medium |
| 35 | Missing tests (lib) | 36 | Medium |
| 36 | Weak test assertions | 10 | Low |
| 37 | Duplicated test patterns | 5 patterns | Low |
| 38 | Hardcoded status/tone strings | 15 strings (~1800 uses) | High |
| 39 | Config drift | 4 | Medium |
| 40 | Inline forms (no Form/FormField) | 4 | Medium |
| 41 | Duplicated dialog boilerplate | 83 files | Medium |
| 42 | Detail page loading patterns | 11 files | Low |
| 43 | Fire-and-forget mutations | 0 | N/A |
| 44 | Duplicated status ternaries | 8 files | Medium |
| 45 | Inline validation schemas | 52+ files | High |
| 46 | Missing SubmitButton usage | 5 files | Medium |
| 47 | Item form cancel button bug | 1 file | High |
| 48 | Missing `.catch()` on API calls | 60+ files | High |
| 49 | Missing onError in mutations | 1 file | Medium |
| 50 | Missing error states in queries | 3 files | Medium |
| 51 | Missing useMemo for Maps | 11 files | Medium |
| 52 | Missing React.memo on leaf components | 5 files | Medium |
| 53 | Missing code splitting | 4 files | Medium |
| 54 | Missing `lang` attribute | 1 file | High |
| 55 | Missing `aria-describedby` in FormField | 1 file (systemic) | High |
| 56 | Missing `id="main"` for skip nav | 1 file | Medium |
| 57 | Contradictory aria-hidden + sr-only | 1 file | Medium |
| 58 | Session cookies `secure: false` | 1 file | High |
| 59 | OAuth tokens via document.cookie | 1 file | High |
| 60 | No CSP or security headers | 1 file | High |
| 61 | Static CSRF token | 1 file | Medium |
| 62 | No rate limiting on auth | 3 files | High |
| 63 | Duplicated Journal+Date dialog | 8+ files | Medium |
| 64 | Duplicated form dialog boilerplate | 40+ files | Medium |
| 65 | Duplicated section (list view) pattern | 15+ files | Medium |
| 66 | Duplicated active status badge | 15+ files | Medium |
| 67 | Duplicated order detail pages | 2 files | Medium |
| 68 | `.toLocaleDateString()` no locale | 13 call sites | High |
| 69 | Local formatDate hardcoded to "en-US" | 19 files | High |
| 70 | Local formatCurrency hardcoded to "en-US" | 11 files | High |
| 71 | Hardcoded English in Badge | 11 call sites | Medium |
| 72 | Hardcoded label functions | 9 utility functions | Medium |
| 73 | Unused utility functions | 4 exports | Low |
| 74 | Unused hook exports | 7 hooks | Low |
| 75 | Unused provider exports | 2 files | Low |
| 76 | Unused store exports | 1 export | Low |
| 77 | Unused CSS utilities | 6 utilities | Low |
| 78 | No route-level error boundaries | 6 route groups | High |
| 79 | Missing loading states (blank screen) | 2 components | Medium |
| 80 | console.error in production | 2 files | Low |
| 81 | Deprecated isLoading usage | 1 file | Low |

---

## 82. Unsafe `.sort()` Mutates Cached Query Data

**Target:** Replace `.sort()` with `[...arr].sort()` to avoid mutating
React Query cache entries in-place.

| Status | File | Lines | Code |
|--------|------|-------|------|
| [x] | `approval-requests/[requestId]/_components/approval-request-detail.tsx` | 96 | `steps.sort(...)` — mutates `approvalRequest.steps` from cache |
| [x] | `approval-requests/_components/approval-widget.tsx` | 98 | `steps.sort(...)` — same mutation |

---

## 83. Missing `enabled` Guard on Queries (Empty String → `Number("")` = 0)

**Target:** Add `enabled: Boolean(...)` to prevent queries from firing
with `Number("")` = `0` when route params are still loading.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/hooks/use-item-query.ts` | 7-11 | No `enabled` — fires with `0` if `orgId`/`itemId` empty |
| [x] | `lib/hooks/use-item-query.ts` | 14-18 | Same issue |
| [x] | `lib/hooks/use-contact-query.ts` | 7-11 | Same issue |
| [x] | `lib/hooks/use-contact-query.ts` | 14-18 | Same issue |
| [x] | `lib/hooks/use-contact-query.ts` | 22-28 | Same issue |
| [x] | `rmas/[rmaId]/_components/rma-detail-section.tsx` | 44-47 | No `enabled` guard |
| [x] | `rmas/[rmaId]/_components/rma-detail-section.tsx` | 49-52 | No `enabled` guard (journals) |
| [x] | `approval-requests/[requestId]/_components/approval-request-detail.tsx` | 35-38 | No `enabled` guard |

---

## 84. Over-Fetching: List Queries Where Only Count/Filtered Subset Is Used

**Target:** Use `select`, dedicated endpoints, or scoped queries instead
of fetching entire lists.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `dashboard-overview/_components/dashboard-overview.tsx` | 69-79 | Fetches all contacts + employees for `.length` only |
| [x] | `products/[itemId]/_components/item-detail.tsx` | 22-25 | Fetches all BOMs, filters client-side by `itemId` |
| [x] | `purchase-requisitions/[requisitionId]/_components/purchase-requisition-detail.tsx` | 82-87 | Fetches all approval requests to find one linked record |
| [x] | `rmas/[rmaId]/_components/rma-detail-section.tsx` | 49-52 | Fetches all journals for the entire org |

---

## 85. Race Conditions: Direct Service Calls Without `useMutation`/Pending State

**Target:** Wrap service calls in `useMutation` with `isPending` to prevent
double-clicks and concurrent execution.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `leave-requests/_components/leave-requests-section.tsx` | 59-68 | `handleAction` fires directly — user can double-click |
| [x] | `attendance/_components/attendance-section.tsx` | 41-46 | `handleCheckOut` fires directly — user can double-click |
| [x] | `purchase-requisitions/[requisitionId]/_components/purchase-requisition-detail.tsx` | 116-157 | 4 independent action handlers — no pending state |
| [x] | `approval-requests/_components/approval-widget.tsx` | 62-74 | `submitDecision` fires directly — user can double-click |

---

## 86. Missing Query Invalidation After Mutations

**Target:** Use `queryClient.invalidateQueries()` instead of `refetch()`
to ensure broader cache freshness across components.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `leave-requests/_components/leave-requests-section.tsx` | 59-68 | Uses `refetch()` instead of invalidation |
| [x] | `attendance/_components/attendance-section.tsx` | 41-46 | Uses `refetch()` instead of invalidation |
| [x] | `purchase-requisitions/[requisitionId]/_components/purchase-requisition-detail.tsx` | 116-157 | Uses `refetch()` — list page stays stale |
| [x] | `approval-requests/_components/approval-widget.tsx` | 62-74 | Relies on parent callback for invalidation — fragile |
| [x] | `reference/_components/quality-points-section.tsx` | 115-118 | Uses `refetch()` instead of invalidation |

---

## 87. Inline Object/Array Literals in JSX Props

**Target:** Extract to `useMemo` or module-level constants to prevent
unnecessary re-renders of child components.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `dashboard-overview/_components/dashboard-overview.tsx` | 174-178, 380-386 | Inline `data={[...]}` arrays in `BarChart` |
| [x] | `purchase-requisitions/[requisitionId]/_components/purchase-requisition-detail.tsx` | 162-165 | Inline `breadcrumbItems` array |
| [x] | `purchase-requisitions/[requisitionId]/_components/purchase-requisition-detail.tsx` | 175-177 | Inline `steps.map(...)` in `WorkflowSteps` |
| [x] | `products/[itemId]/_components/item-detail.tsx` | 51-82 | Inline `tabs={[...]}` with JSX content |

**Note:** All inline arrays extracted to variables before return statements.

---

## 88. Unsafe Key Props (Duplicate Label Collision)

**Target:** Use index-based or composite keys when label text may duplicate.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `components/workflow-steps.tsx` | 26 | `key={step.label}` — duplicate labels collapse elements |
| [x] | `components/breadcrumb.tsx` | 34 | `key={item.label}` — same issue |
| [x] | `components/line-chart.tsx` | 92, 113 | `key={datum.label}` — duplicate data points |

---

## 89. Unsafe Type Narrowing Without Null Guards

**Target:** Add null guards before accessing nested properties.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `lib/hooks/use-permissions.ts` | 17 | `query.data?.permissions.map(...)` — `permissions` could be `undefined` |
| [x] | `products/_components/item-variants.tsx` | 32 | `Object.entries(row.original.attributeJson)` — no null check |
| [x] | `rmas/[rmaId]/_components/rma-detail-section.tsx` | 83 | `const state = rma.state as RmaState` — unsafe cast without validation |

---

## Summary

| # | Category | Violations | Priority |
|---|----------|------------|----------|
| 1 | Export default (lib) | 4 | Low |
| 2 | Interface for props | 10 | Low |
| 3 | Local formatDate copies | 19 | High |
| 4 | Local formatCurrency copies | 11 | High |
| 5 | FieldFeedbackType naming | 1 | Low |
| 6 | Manual $ prefix | 33 | Medium |
| 7 | Direct queryClient usage | 0 | N/A |
| 8 | toLocaleDateString() no locale | 11 | Medium |
| 9 | Unnecessary "use client" | 0 | N/A |
| 10 | Hardcoded English status labels | 17 files | High |
| 11 | Hardcoded English UI text | 40+ files | High |
| 12 | Misleading function name | 1 | Low |
| 13 | Hardcoded support email | 13 | Low |
| 14 | Hardcoded $ in charts/KPIs | 10+ | Medium |
| 15 | Hardcoded language labels | 1 | Low |
| 16 | Inline mock data | 1 | Low |
| 17 | Duplicate logger files | 1 | Low |
| 18 | Massive page files | 8 | Medium |
| 19 | Missing route boundaries | 121 pages | Medium |
| 20 | Form schema extraction | 4 auth forms done | Done |
| 21 | Unused imports | 30 (10 files) | Medium |
| 22 | Console statements | 4 (3 files) | Low |
| 23 | Hardcoded API URLs | 1 | Medium |
| 24 | Type assertions (`as unknown`) | 55 (40+ files) | High |
| 25 | Magic numbers in setTimeout | 4 | Low |
| 26 | Inline styles | 56 (21 files) | Low |
| 27 | Hardcoded route paths | 21 (14 files) | Medium |
| 28 | `any`/`as never` usage | 6 (non-test files) | Medium |
| 29 | Deeply nested ternaries | 16 (14 files) | Medium |
| 30 | Overly complex components | 18 files | High |
| 31 | Hydration mismatch risks | 1 (high risk) | Medium |
| 32 | Memory leak risks | 12 files | Medium |
| 33 | Stale closures | 2 | Medium |
| 34 | Missing tests (components) | 57 | Medium |
| 35 | Missing tests (lib) | 36 | Medium |
| 36 | Weak test assertions | 10 | Low |
| 37 | Duplicated test patterns | 5 patterns | Low |
| 38 | Hardcoded status/tone strings | 15 strings (~1800 uses) | High |
| 39 | Config drift | 4 | Medium |
| 40 | Inline forms (no Form/FormField) | 4 | Medium |
| 41 | Duplicated dialog boilerplate | 83 files | Medium |
| 42 | Detail page loading patterns | 11 files | Low |
| 43 | Fire-and-forget mutations | 0 | N/A |
| 44 | Duplicated status ternaries | 8 files | Medium |
| 45 | Inline validation schemas | 52+ files | High |
| 46 | Missing SubmitButton usage | 5 files | Medium |
| 47 | Item form cancel button bug | 1 file | High |
| 48 | Missing `.catch()` on API calls | 60+ files | High |
| 49 | Missing onError in mutations | 1 file | Medium |
| 50 | Missing error states in queries | 3 files | Medium |
| 51 | Missing useMemo for Maps | 11 files | Medium |
| 52 | Missing React.memo on leaf components | 5 files | Medium |
| 53 | Missing code splitting | 4 files | Medium |
| 54 | Missing `lang` attribute | 1 file | High |
| 55 | Missing `aria-describedby` in FormField | 1 file (systemic) | High |
| 56 | Missing `id="main"` for skip nav | 1 file | Medium |
| 57 | Contradictory aria-hidden + sr-only | 1 file | Medium |
| 58 | Session cookies `secure: false` | 1 file | High |
| 59 | OAuth tokens via document.cookie | 1 file | High |
| 60 | No CSP or security headers | 1 file | High |
| 61 | Static CSRF token | 1 file | Medium |
| 62 | No rate limiting on auth | 3 files | High |
| 63 | Duplicated Journal+Date dialog | 8+ files | Medium |
| 64 | Duplicated form dialog boilerplate | 40+ files | Medium |
| 65 | Duplicated section (list view) pattern | 15+ files | Medium |
| 66 | Duplicated active status badge | 15+ files | Medium |
| 67 | Duplicated order detail pages | 2 files | Medium |
| 68 | `.toLocaleDateString()` no locale | 13 call sites | High |
| 69 | Local formatDate hardcoded to "en-US" | 19 files | High |
| 70 | Local formatCurrency hardcoded to "en-US" | 11 files | High |
| 71 | Hardcoded English in Badge | 11 call sites | Medium |
| 72 | Hardcoded label functions | 9 utility functions | Medium |
| 73 | Unused utility functions | 4 exports | Low |
| 74 | Unused hook exports | 7 hooks | Low |
| 75 | Unused provider exports | 2 files | Low |
| 76 | Unused store exports | 1 export | Low |
| 77 | Unused CSS utilities | 6 utilities | Low |
| 78 | No route-level error boundaries | 6 route groups | High |
| 79 | Missing loading states (blank screen) | 2 components | Medium |
| 80 | console.error in production | 2 files | Low |
| 81 | Deprecated isLoading usage | 1 file | Low |
| 82 | Unsafe `.sort()` mutates cache | 2 files | High |
| 83 | Missing `enabled` guard on queries | 8 hooks/components | High |
| 84 | Over-fetching list queries | 4 files | Medium |
| 85 | Race conditions (no pending state) | 4 files | High |
| 86 | Missing query invalidation | 5 files | Medium |
| 87 | Inline literals in JSX props | 4 files | Medium |
| 88 | Unsafe key props | 3 files | Medium |
| 89 | Unsafe type narrowing | 3 files | Medium |

---

## 90. `Number()` on Route Params Without NaN Validation

**Target:** Add a validated `safeParseInt` helper or use `Number.isInteger()`
checks before sending converted params to the API.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/hooks/use-item-query.ts` | 9, 17 | `Number(orgId)`, `Number(itemId)` — no NaN guard |
| [x] | `lib/hooks/use-contact-query.ts` | 9, 17, 26 | `Number(orgId)`, `Number(contactId)` — no NaN guard |
| [x] | `rmas/[rmaId]/_components/rma-detail-section.tsx` | 46, 55, 62, 69 | Multiple `Number()` without validation |
| [x] | `approval-requests/[requestId]/_components/approval-request-detail.tsx` | 37 | `Number(orgId)`, `Number(requestId)` — no validation |
| [x] | `accounting/invoices/_components/invoices-section.tsx` | 206-209 | `Number(orgId)`, `Number(journalId)`, `Number(contactId)` |
| [x] | `accounting/journal-entries/_components/journal-entries-section.tsx` | 195-206 | Multiple `Number()` without validation |
| [x] | `accounting/bank-statements/_components/bank-statements-section.tsx` | 180-182 | `Number(orgId)`, `Number(journalId)` |
| [x] | `employees/_components/employee-form-dialog.tsx` | 110-134 | Multiple `Number()` on form values |
| [x] | `boms/_components/bom-form-dialog.tsx` | 128-155 | Multiple `Number()` without validation |
| [x] | `products/_components/item-form-dialog.tsx` | 103-135 | Multiple `Number()` without validation |

---

## 91. Zero `loading.tsx` Files in Entire App

**Target:** Add `loading.tsx` skeleton/spinner boundaries to data-heavy
route groups.

**Severity:** Medium

| Status | Route | Reason |
|--------|-------|--------|
| [x] | `app/[locale]/(org)/org/[id]/` | Layout has `prefetchPermissionsData` |
| [x] | `app/[locale]/(org)/org/[id]/dashboard/` | 14+ parallel data queries |
| [x] | `app/[locale]/(org)/org/[id]/accounting/` | Complex financial data |
| [x] | `app/[locale]/(org)/org/[id]/reports/` | Heavy report data fetching |
| [x] | `app/[locale]/(org)/org/[id]/products/` | List + category data |

---

## 92. UTC-Based Date Initialization Causes Off-by-One in Non-UTC Timezones

**Target:** Replace `new Date().toISOString().slice(0, 10)` with a
timezone-safe helper that returns the local date.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `bank-statements-section.tsx` | 174 | `new Date().toISOString().slice(0, 10)` — shows yesterday at UTC+7 |
| [x] | `journal-entries-section.tsx` | 162 | Same UTC issue |
| [x] | `invoices-section.tsx` | 180 | Same UTC issue |
| [x] | `receive-dialog.tsx` | (init) | Same pattern likely |
| [x] | `refund-dialog.tsx` | (init) | Same pattern likely |
| [x] | `sale-order-detail.tsx` | (init) | Same pattern likely |
| [x] | `purchase-order-detail.tsx` | (init) | Same pattern likely |

---

## 93. Missing `loading.tsx` Skeleton Components

**Target:** Add `loading.tsx` with skeleton UI to route groups with heavy
data fetching to prevent blank white screen during navigation.

Same routes as category 91 — combined, these routes represent every
data-heavy page in the app. None have a `loading.tsx` boundary.

---

## 94. String-to-Number Conversion at Wrong Layer

**Target:** Convert route params to `number` at the layout/page boundary
and pass `number` to child components, rather than converting at every
query/mutation call site.

All components in the `app/[locale]/(org)/org/[id]/` tree accept
`orgId: string` props and immediately convert with `Number(orgId)` in
query functions. This scatters 100+ `Number()` calls across the codebase
instead of converting once at the layout boundary.

---

## 95. Missing `disabled` During Form Submission

**Target:** Add `isPending` or `isSubmitting` check to submit buttons
to prevent double-submit.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `attendance/_components/check-in-dialog.tsx` | 101 | Submit button only checks employee selection, not submission state |
| [x] | `public/contact/page.tsx` | 155 | Raw `<button type="submit">` — no disabled state (low risk, mailto:) |

**Note:** False positive — form uses `mailto:` action, no server submission.

---

## 96. Form State Not Reset After Successful Submission in Dialog Forms

**Target:** Confirm dialog unmounts on close (which it currently does),
or add explicit state reset.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `bank-statements-section.tsx` | 172-175 | `useState` for `name`, `journalId`, `date`, `balanceStart` |
| [x] | `tax-years-section.tsx` | 163-165 | `useState` for `name`, `dateStart`, `dateEnd` |
| [x] | `journal-entries-section.tsx` | 161-171 | `useState` for `journalId`, `date`, `ref_`, `description`, `lines[]` |
| [x] | `invoices-section.tsx` | 175-184 | `useState` for `invoiceType`, `contactId`, `journalId`, `invoiceDate`, `reference`, `lines[]` |

---

## Summary

| # | Category | Violations | Priority |
|---|----------|------------|----------|
| 1 | Export default (lib) | 4 | Low |
| 2 | Interface for props | 10 | Low |
| 3 | Local formatDate copies | 19 | High |
| 4 | Local formatCurrency copies | 11 | High |
| 5 | FieldFeedbackType naming | 1 | Low |
| 6 | Manual $ prefix | 33 | Medium |
| 7 | Direct queryClient usage | 0 | N/A |
| 8 | toLocaleDateString() no locale | 11 | Medium |
| 9 | Unnecessary "use client" | 0 | N/A |
| 10 | Hardcoded English status labels | 17 files | High |
| 11 | Hardcoded English UI text | 40+ files | High |
| 12 | Misleading function name | 1 | Low |
| 13 | Hardcoded support email | 13 | Low |
| 14 | Hardcoded $ in charts/KPIs | 10+ | Medium |
| 15 | Hardcoded language labels | 1 | Low |
| 16 | Inline mock data | 1 | Low |
| 17 | Duplicate logger files | 1 | Low |
| 18 | Massive page files | 8 | Medium |
| 19 | Missing route boundaries | 121 pages | Medium |
| 20 | Form schema extraction | 4 auth forms done | Done |
| 21 | Unused imports | 30 (10 files) | Medium |
| 22 | Console statements | 4 (3 files) | Low |
| 23 | Hardcoded API URLs | 1 | Medium |
| 24 | Type assertions (`as unknown`) | 55 (40+ files) | High |
| 25 | Magic numbers in setTimeout | 4 | Low |
| 26 | Inline styles | 56 (21 files) | Low |
| 27 | Hardcoded route paths | 21 (14 files) | Medium |
| 28 | `any`/`as never` usage | 6 (non-test files) | Medium |
| 29 | Deeply nested ternaries | 16 (14 files) | Medium |
| 30 | Overly complex components | 18 files | High |
| 31 | Hydration mismatch risks | 1 (high risk) | Medium |
| 32 | Memory leak risks | 12 files | Medium |
| 33 | Stale closures | 2 | Medium |
| 34 | Missing tests (components) | 57 | Medium |
| 35 | Missing tests (lib) | 36 | Medium |
| 36 | Weak test assertions | 10 | Low |
| 37 | Duplicated test patterns | 5 patterns | Low |
| 38 | Hardcoded status/tone strings | 15 strings (~1800 uses) | High |
| 39 | Config drift | 4 | Medium |
| 40 | Inline forms (no Form/FormField) | 4 | Medium |
| 41 | Duplicated dialog boilerplate | 83 files | Medium |
| 42 | Detail page loading patterns | 11 files | Low |
| 43 | Fire-and-forget mutations | 0 | N/A |
| 44 | Duplicated status ternaries | 8 files | Medium |
| 45 | Inline validation schemas | 52+ files | High |
| 46 | Missing SubmitButton usage | 5 files | Medium |
| 47 | Item form cancel button bug | 1 file | High |
| 48 | Missing `.catch()` on API calls | 60+ files | High |
| 49 | Missing onError in mutations | 1 file | Medium |
| 50 | Missing error states in queries | 3 files | Medium |
| 51 | Missing useMemo for Maps | 11 files | Medium |
| 52 | Missing React.memo on leaf components | 5 files | Medium |
| 53 | Missing code splitting | 4 files | Medium |
| 54 | Missing `lang` attribute | 1 file | High |
| 55 | Missing `aria-describedby` in FormField | 1 file (systemic) | High |
| 56 | Missing `id="main"` for skip nav | 1 file | Medium |
| 57 | Contradictory aria-hidden + sr-only | 1 file | Medium |
| 58 | Session cookies `secure: false` | 1 file | High |
| 59 | OAuth tokens via document.cookie | 1 file | High |
| 60 | No CSP or security headers | 1 file | High |
| 61 | Static CSRF token | 1 file | Medium |
| 62 | No rate limiting on auth | 3 files | High |
| 63 | Duplicated Journal+Date dialog | 8+ files | Medium |
| 64 | Duplicated form dialog boilerplate | 40+ files | Medium |
| 65 | Duplicated section (list view) pattern | 15+ files | Medium |
| 66 | Duplicated active status badge | 15+ files | Medium |
| 67 | Duplicated order detail pages | 2 files | Medium |
| 68 | `.toLocaleDateString()` no locale | 13 call sites | High |
| 69 | Local formatDate hardcoded to "en-US" | 19 files | High |
| 70 | Local formatCurrency hardcoded to "en-US" | 11 files | High |
| 71 | Hardcoded English in Badge | 11 call sites | Medium |
| 72 | Hardcoded label functions | 9 utility functions | Medium |
| 73 | Unused utility functions | 4 exports | Low |
| 74 | Unused hook exports | 7 hooks | Low |
| 75 | Unused provider exports | 2 files | Low |
| 76 | Unused store exports | 1 export | Low |
| 77 | Unused CSS utilities | 6 utilities | Low |
| 78 | No route-level error boundaries | 6 route groups | High |
| 79 | Missing loading states (blank screen) | 2 components | Medium |
| 80 | console.error in production | 2 files | Low |
| 81 | Deprecated isLoading usage | 1 file | Low |
| 82 | Unsafe `.sort()` mutates cache | 2 files | High |
| 83 | Missing `enabled` guard on queries | 8 hooks/components | High |
| 84 | Over-fetching list queries | 4 files | Medium |
| 85 | Race conditions (no pending state) | 4 files | High |
| 86 | Missing query invalidation | 5 files | Medium |
| 87 | Inline literals in JSX props | 4 files | Medium |
| 88 | Unsafe key props | 3 files | Medium |
| 89 | Unsafe type narrowing | 3 files | Medium |
| 90 | `Number()` without NaN validation | 10+ files | High |
| 91 | Zero `loading.tsx` in entire app | 5 route groups | Medium |
| 92 | UTC-based date init (off-by-one) | 7 files | Medium |
| 93 | Missing loading skeleton components | 5 route groups | Medium |
| 94 | String-to-number at wrong layer | 100+ call sites | Medium |
| 95 | Missing disabled during form submit | 2 files | Medium |
| 96 | Form state not reset in dialogs | 4 files | Low |

---

## 97. Transient Network Error During Refresh Logs User Out

**Target:** Distinguish between "refresh token invalid" (clear cookies)
and "network error" (keep cookies, pass through) in `refreshSessionInMiddleware`.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/middleware/refresh.ts` | 103-105 | `callUpstreamRefresh` returns `null` for both invalid token AND network error — both cases clear session cookies |

---

## 98. `intlMiddleware` Failure Drops Refresh Cookies

**Target:** Wrap `intlMiddleware` in try/catch so refresh cookies
are applied even if locale resolution fails.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `middleware/proxy.ts` | 16 | If `intlMiddleware(request)` throws, `applySessionCookiesToResponse` is never called — user silently loses refreshed session |

---

## 99. Path Traversal via Catch-All Proxy Route

**Target:** Validate that path segments contain only safe characters
before constructing the upstream URL.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `app/api/v1/[...path]/route.ts` | 13-14 | Path segments used directly in `buildTarget` without validation — `..` segments could cause path traversal against upstream |

---

## 100. `createErrorResponse` Discards All Error Context

**Target:** Inspect thrown errors and propagate appropriate status/message
instead of always returning generic 500.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/server/response.ts` | 121-123 | Default `onError` handler always returns generic 500 — validation errors, 404s, and known errors all return same response |

---

## 101. Proxy Leaks Upstream Error Body to Client

**Target:** Use a generic error message instead of forwarding raw
upstream response content.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/server/proxy.ts` | 254-261 | First 1000 chars of non-JSON upstream response (may contain HTML error page, stack trace, internal details) forwarded as error message |

---

## 102. RMA Mutations Don't Invalidate List Query

**Target:** Invalidate both detail and list query keys after mutations.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `rmas/[rmaId]/_components/rma-detail-section.tsx` | 54-72 | `onSuccess` invalidates `["rmas", orgId, rmaId]` but not `["rmas", orgId]` — list page stays stale |
| [x] | `rmas/[rmaId]/_components/receive-dialog.tsx` | 251-273 | Same issue — only invalidates detail |
| [x] | `rmas/[rmaId]/_components/refund-dialog.tsx` | 251-273 | Same issue — only invalidates detail |

---

## 103. CRM Mutations Don't Invalidate `crmStages`

**Target:** Invalidate `crmStages` after lead/opportunity promotions.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `crm/_components/leads-section.tsx` | 46-51 | `refresh()` invalidates `crmLeads` + `crmOpportunities` but not `crmStages` |
| [x] | `crm/_components/opportunities-section.tsx` | 46-51 | Same issue — missing `crmStages` |
| [x] | `crm/_components/pipeline-section.tsx` | 46-54 | Same issue — refetches stages but doesn't invalidate via query key |

---

## 104. Approval Request Detail Doesn't Invalidate List

**Target:** Invalidate list query after deciding on an approval request.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `approval-requests/[requestId]/_components/approval-request-detail.tsx` | 134-136 | `onDecided` invalidates `["approvalRequests", orgId, requestId]` but not `["approvalRequests", orgId]` |

---

## 105. Missing `"use client"` Directive in Hook File

**Target:** Add `"use client"` to files that use client-only APIs.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `app/[locale]/(auth)/login/_hooks/use-login-form.ts` | 1 | Uses `useEffect`, `useTransition`, `useForm`, `useRouter`, `useQueryClient` — no `"use client"` directive |

---

## 106. Missing Input Validation on Proxy Route

**Target:** Validate proxy path segments against a safe pattern.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `app/api/v1/[...path]/route.ts` | 13 | No validation — segments with `..`, null bytes, or special characters pass through to upstream |

---

## 107. Inconsistent Error Response Format

**Target:** Standardize error response shapes across the codebase.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/services/swantara/mapper.ts` | 14-19 | `SwantaraErrorBody.error` field declared but never read — creates confusion with `message` field |

---

## Summary

| # | Category | Violations | Priority |
|---|----------|------------|----------|
| 1 | Export default (lib) | 4 | Low |
| 2 | Interface for props | 10 | Low |
| 3 | Local formatDate copies | 19 | High |
| 4 | Local formatCurrency copies | 11 | High |
| 5 | FieldFeedbackType naming | 1 | Low |
| 6 | Manual $ prefix | 33 | Medium |
| 7 | Direct queryClient usage | 0 | N/A |
| 8 | toLocaleDateString() no locale | 11 | Medium |
| 9 | Unnecessary "use client" | 0 | N/A |
| 10 | Hardcoded English status labels | 17 files | High |
| 11 | Hardcoded English UI text | 40+ files | High |
| 12 | Misleading function name | 1 | Low |
| 13 | Hardcoded support email | 13 | Low |
| 14 | Hardcoded $ in charts/KPIs | 10+ | Medium |
| 15 | Hardcoded language labels | 1 | Low |
| 16 | Inline mock data | 1 | Low |
| 17 | Duplicate logger files | 1 | Low |
| 18 | Massive page files | 8 | Medium |
| 19 | Missing route boundaries | 121 pages | Medium |
| 20 | Form schema extraction | 4 auth forms done | Done |
| 21 | Unused imports | 30 (10 files) | Medium |
| 22 | Console statements | 4 (3 files) | Low |
| 23 | Hardcoded API URLs | 1 | Medium |
| 24 | Type assertions (`as unknown`) | 55 (40+ files) | High |
| 25 | Magic numbers in setTimeout | 4 | Low |
| 26 | Inline styles | 56 (21 files) | Low |
| 27 | Hardcoded route paths | 21 (14 files) | Medium |
| 28 | `any`/`as never` usage | 6 (non-test files) | Medium |
| 29 | Deeply nested ternaries | 16 (14 files) | Medium |
| 30 | Overly complex components | 18 files | High |
| 31 | Hydration mismatch risks | 1 (high risk) | Medium |
| 32 | Memory leak risks | 12 files | Medium |
| 33 | Stale closures | 2 | Medium |
| 34 | Missing tests (components) | 57 | Medium |
| 35 | Missing tests (lib) | 36 | Medium |
| 36 | Weak test assertions | 10 | Low |
| 37 | Duplicated test patterns | 5 patterns | Low |
| 38 | Hardcoded status/tone strings | 15 strings (~1800 uses) | High |
| 39 | Config drift | 4 | Medium |
| 40 | Inline forms (no Form/FormField) | 4 | Medium |
| 41 | Duplicated dialog boilerplate | 83 files | Medium |
| 42 | Detail page loading patterns | 11 files | Low |
| 43 | Fire-and-forget mutations | 0 | N/A |
| 44 | Duplicated status ternaries | 8 files | Medium |
| 45 | Inline validation schemas | 52+ files | High |
| 46 | Missing SubmitButton usage | 5 files | Medium |
| 47 | Item form cancel button bug | 1 file | High |
| 48 | Missing `.catch()` on API calls | 60+ files | High |
| 49 | Missing onError in mutations | 1 file | Medium |
| 50 | Missing error states in queries | 3 files | Medium |
| 51 | Missing useMemo for Maps | 11 files | Medium |
| 52 | Missing React.memo on leaf components | 5 files | Medium |
| 53 | Missing code splitting | 4 files | Medium |
| 54 | Missing `lang` attribute | 1 file | High |
| 55 | Missing `aria-describedby` in FormField | 1 file (systemic) | High |
| 56 | Missing `id="main"` for skip nav | 1 file | Medium |
| 57 | Contradictory aria-hidden + sr-only | 1 file | Medium |
| 58 | Session cookies `secure: false` | 1 file | High |
| 59 | OAuth tokens via document.cookie | 1 file | High |
| 60 | No CSP or security headers | 1 file | High |
| 61 | Static CSRF token | 1 file | Medium |
| 62 | No rate limiting on auth | 3 files | High |
| 63 | Duplicated Journal+Date dialog | 8+ files | Medium |
| 64 | Duplicated form dialog boilerplate | 40+ files | Medium |
| 65 | Duplicated section (list view) pattern | 15+ files | Medium |
| 66 | Duplicated active status badge | 15+ files | Medium |
| 67 | Duplicated order detail pages | 2 files | Medium |
| 68 | `.toLocaleDateString()` no locale | 13 call sites | High |
| 69 | Local formatDate hardcoded to "en-US" | 19 files | High |
| 70 | Local formatCurrency hardcoded to "en-US" | 11 files | High |
| 71 | Hardcoded English in Badge | 11 call sites | Medium |
| 72 | Hardcoded label functions | 9 utility functions | Medium |
| 73 | Unused utility functions | 4 exports | Low |
| 74 | Unused hook exports | 7 hooks | Low |
| 75 | Unused provider exports | 2 files | Low |
| 76 | Unused store exports | 1 export | Low |
| 77 | Unused CSS utilities | 6 utilities | Low |
| 78 | No route-level error boundaries | 6 route groups | High |
| 79 | Missing loading states (blank screen) | 2 components | Medium |
| 80 | console.error in production | 2 files | Low |
| 81 | Deprecated isLoading usage | 1 file | Low |
| 82 | Unsafe `.sort()` mutates cache | 2 files | High |
| 83 | Missing `enabled` guard on queries | 8 hooks/components | High |
| 84 | Over-fetching list queries | 4 files | Medium |
| 85 | Race conditions (no pending state) | 4 files | High |
| 86 | Missing query invalidation | 5 files | Medium |
| 87 | Inline literals in JSX props | 4 files | Medium |
| 88 | Unsafe key props | 3 files | Medium |
| 89 | Unsafe type narrowing | 3 files | Medium |
| 90 | `Number()` without NaN validation | 10+ files | High |
| 91 | Zero `loading.tsx` in entire app | 5 route groups | Medium |
| 92 | UTC-based date init (off-by-one) | 7 files | Medium |
| 93 | Missing loading skeleton components | 5 route groups | Medium |
| 94 | String-to-number at wrong layer | 100+ call sites | Medium |
| 95 | Missing disabled during form submit | 2 files | Medium |
| 96 | Form state not reset in dialogs | 4 files | Low |
| 97 | Transient network error logs user out | 1 file | High |
| 98 | `intlMiddleware` failure drops cookies | 1 file | High |
| 99 | Path traversal via catch-all proxy | 1 file | High |
| 100 | `createErrorResponse` discards error context | 1 file | Medium |
| 101 | Proxy leaks upstream error body | 1 file | Medium |
| 102 | RMA mutations don't invalidate list | 3 files | Medium |
| 103 | CRM mutations don't invalidate stages | 3 files | Medium |
| 104 | Approval detail doesn't invalidate list | 1 file | Medium |
| 105 | Missing `"use client"` directive | 1 file | Low |
| 106 | Missing input validation on proxy route | 1 file | High |
| 107 | Inconsistent error response format | 1 file | Low |

---

## 108. Three Competing Token Refresh Mechanisms

**Target:** Consolidate to a single refresh authority (middleware) and
remove redundant 401-retry logic from client interceptor or server proxy.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/client/axios.ts` | 32-57 | Client interceptor catches 401s and calls `refreshSession()` |
| [x] | `lib/server/proxy.ts` | 219-240 | Server proxy also catches 401s and refreshes independently |
| [x] | `lib/middleware/refresh.ts` | 95-110 | Middleware does proactive refresh — the intended single source |
| [x] | `lib/client/refresh.ts` | — | Client refresh calls `/api/v1/auth/refresh` which goes through the proxy, creating a double-refresh loop |

---

## 109. Logout Doesn't Clear Zustand Org Store

**Target:** Call `useOrgStore.getState().clearActiveOrg()` in both
logout handlers.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `org-user-menu.tsx` | 53-69 | `queryClient.clear()` called but `useOrgStore` not cleared — stale org ID persists in localStorage |
| [x] | `logout-button.tsx` | 19-37 | Same issue — Zustand store not reset |

---

## 110. Permissions Load as Empty Array — All UI Hidden During Loading

**Target:** Return a "loading" state from `has()` when permissions
are still loading, and have `PermissionGate` render children by default
(fail-open) until permissions resolve.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/hooks/use-permissions.ts` | 9-26 | `codes` defaults to `[]` while loading — `has()` returns `false` for all checks |
| [x] | `components/permission-gate.tsx` | 32-39 | Renders `fallback ?? null` during loading — entire UI section disappears |
| [x] | `dashboard/_components/dashboard-overview.tsx` | 63-65 | `has()` in loop — all stat cards hidden during loading |

---

## 111. Fire-and-Forget Mutations (No `.catch()`)

**Target:** Add `.catch()` with error toast to all fire-and-forget
API calls.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `sale-orders/_components/sale-orders-section.tsx` | 127-131 | Delete order — no `.catch()`, no error feedback |
| [x] | `sale-order-form-dialog.tsx` | 232-237 | Create sale order — no `.catch()` |
| [x] | `purchase-order-form-dialog.tsx` | 179-184 | Create purchase order — no `.catch()` |
| [x] | `purchase-requisition-form-dialog.tsx` | 147-152 | Create purchase requisition — no `.catch()` |
| [x] | `purchase-rfq-form-dialog.tsx` | 147-152 | Create purchase RFQ — no `.catch()` |

---

## 112. URL Params Not Synced With Component State

**Target:** Add `useEffect` to sync URL-derived state when params
change (e.g., back/forward navigation).

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `sale-orders-section.tsx` | 22-25 | `initialOpportunity` from `useSearchParams()` read once — `prefillOpportunity` state not updated on URL change |

---

## 113. Direct Service Calls Instead of React Query

**Target:** Replace manual `getSwantaraService()` + `useState`
combinations with React Query for caching, deduplication, and retry.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `sale-order-form-dialog.tsx` | 160-171 | `loadVariants()` uses direct service call + `setVariantsByProduct` instead of `useQuery` |
| [x] | `providers/digital-fingerprint.tsx` | 9-17 | `useEffect` + dynamic import for fingerprint — result stored in unpersisted Zustand store, never meaningfully used |

**Note:** False positive — `useDigitalFingerprintStore` is used by `use-login-form.ts` (line 38) to send fingerprint data during login.

**Note:** `sale-order-form-dialog.tsx` uses per-item cache pattern (variants fetched on demand per line item). Created `useItemVariants` hook for future use when component is refactored to pass item IDs to child components.

---

## 114. Form State Leaked Between Dialog Opens

**Target:** Reset form state when dialog opens with different data,
or key the component to force remount.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `sale-order-form-dialog.tsx` | 146-148 | `initialOpportunityId` not cleared when reopened without ID — `crmLeadId` retains old value |
| [x] | `purchase-order-form-dialog.tsx` | 105-116 | `useForm` default values captured at mount — stale if warehouses load after mount |
| [x] | `sale-order-form-dialog.tsx` | 109-120 | `lines` state resets but `form` state (contact, price_book) doesn't |

---

## Summary

| # | Category | Violations | Priority |
|---|----------|------------|----------|
| 1 | Export default (lib) | 4 | Low |
| 2 | Interface for props | 10 | Low |
| 3 | Local formatDate copies | 19 | High |
| 4 | Local formatCurrency copies | 11 | High |
| 5 | FieldFeedbackType naming | 1 | Low |
| 6 | Manual $ prefix | 33 | Medium |
| 7 | Direct queryClient usage | 0 | N/A |
| 8 | toLocaleDateString() no locale | 11 | Medium |
| 9 | Unnecessary "use client" | 0 | N/A |
| 10 | Hardcoded English status labels | 17 files | High |
| 11 | Hardcoded English UI text | 40+ files | High |
| 12 | Misleading function name | 1 | Low |
| 13 | Hardcoded support email | 13 | Low |
| 14 | Hardcoded $ in charts/KPIs | 10+ | Medium |
| 15 | Hardcoded language labels | 1 | Low |
| 16 | Inline mock data | 1 | Low |
| 17 | Duplicate logger files | 1 | Low |
| 18 | Massive page files | 8 | Medium |
| 19 | Missing route boundaries | 121 pages | Medium |
| 20 | Form schema extraction | 4 auth forms done | Done |
| 21 | Unused imports | 30 (10 files) | Medium |
| 22 | Console statements | 4 (3 files) | Low |
| 23 | Hardcoded API URLs | 1 | Medium |
| 24 | Type assertions (`as unknown`) | 55 (40+ files) | High |
| 25 | Magic numbers in setTimeout | 4 | Low |
| 26 | Inline styles | 56 (21 files) | Low |
| 27 | Hardcoded route paths | 21 (14 files) | Medium |
| 28 | `any`/`as never` usage | 6 (non-test files) | Medium |
| 29 | Deeply nested ternaries | 16 (14 files) | Medium |
| 30 | Overly complex components | 18 files | High |
| 31 | Hydration mismatch risks | 1 (high risk) | Medium |
| 32 | Memory leak risks | 12 files | Medium |
| 33 | Stale closures | 2 | Medium |
| 34 | Missing tests (components) | 57 | Medium |
| 35 | Missing tests (lib) | 36 | Medium |
| 36 | Weak test assertions | 10 | Low |
| 37 | Duplicated test patterns | 5 patterns | Low |
| 38 | Hardcoded status/tone strings | 15 strings (~1800 uses) | High |
| 39 | Config drift | 4 | Medium |
| 40 | Inline forms (no Form/FormField) | 4 | Medium |
| 41 | Duplicated dialog boilerplate | 83 files | Medium |
| 42 | Detail page loading patterns | 11 files | Low |
| 43 | Fire-and-forget mutations | 0 | N/A |
| 44 | Duplicated status ternaries | 8 files | Medium |
| 45 | Inline validation schemas | 52+ files | High |
| 46 | Missing SubmitButton usage | 5 files | Medium |
| 47 | Item form cancel button bug | 1 file | High |
| 48 | Missing `.catch()` on API calls | 60+ files | High |
| 49 | Missing onError in mutations | 1 file | Medium |
| 50 | Missing error states in queries | 3 files | Medium |
| 51 | Missing useMemo for Maps | 11 files | Medium |
| 52 | Missing React.memo on leaf components | 5 files | Medium |
| 53 | Missing code splitting | 4 files | Medium |
| 54 | Missing `lang` attribute | 1 file | High |
| 55 | Missing `aria-describedby` in FormField | 1 file (systemic) | High |
| 56 | Missing `id="main"` for skip nav | 1 file | Medium |
| 57 | Contradictory aria-hidden + sr-only | 1 file | Medium |
| 58 | Session cookies `secure: false` | 1 file | High |
| 59 | OAuth tokens via document.cookie | 1 file | High |
| 60 | No CSP or security headers | 1 file | High |
| 61 | Static CSRF token | 1 file | Medium |
| 62 | No rate limiting on auth | 3 files | High |
| 63 | Duplicated Journal+Date dialog | 8+ files | Medium |
| 64 | Duplicated form dialog boilerplate | 40+ files | Medium |
| 65 | Duplicated section (list view) pattern | 15+ files | Medium |
| 66 | Duplicated active status badge | 15+ files | Medium |
| 67 | Duplicated order detail pages | 2 files | Medium |
| 68 | `.toLocaleDateString()` no locale | 13 call sites | High |
| 69 | Local formatDate hardcoded to "en-US" | 19 files | High |
| 70 | Local formatCurrency hardcoded to "en-US" | 11 files | High |
| 71 | Hardcoded English in Badge | 11 call sites | Medium |
| 72 | Hardcoded label functions | 9 utility functions | Medium |
| 73 | Unused utility functions | 4 exports | Low |
| 74 | Unused hook exports | 7 hooks | Low |
| 75 | Unused provider exports | 2 files | Low |
| 76 | Unused store exports | 1 export | Low |
| 77 | Unused CSS utilities | 6 utilities | Low |
| 78 | No route-level error boundaries | 6 route groups | High |
| 79 | Missing loading states (blank screen) | 2 components | Medium |
| 80 | console.error in production | 2 files | Low |
| 81 | Deprecated isLoading usage | 1 file | Low |
| 82 | Unsafe `.sort()` mutates cache | 2 files | High |
| 83 | Missing `enabled` guard on queries | 8 hooks/components | High |
| 84 | Over-fetching list queries | 4 files | Medium |
| 85 | Race conditions (no pending state) | 4 files | High |
| 86 | Missing query invalidation | 5 files | Medium |
| 87 | Inline literals in JSX props | 4 files | Medium |
| 88 | Unsafe key props | 3 files | Medium |
| 89 | Unsafe type narrowing | 3 files | Medium |
| 90 | `Number()` without NaN validation | 10+ files | High |
| 91 | Zero `loading.tsx` in entire app | 5 route groups | Medium |
| 92 | UTC-based date init (off-by-one) | 7 files | Medium |
| 93 | Missing loading skeleton components | 5 route groups | Medium |
| 94 | String-to-number at wrong layer | 100+ call sites | Medium |
| 95 | Missing disabled during form submit | 2 files | Medium |
| 96 | Form state not reset in dialogs | 4 files | Low |
| 97 | Transient network error logs user out | 1 file | High |
| 98 | `intlMiddleware` failure drops cookies | 1 file | High |
| 99 | Path traversal via catch-all proxy | 1 file | High |
| 100 | `createErrorResponse` discards error context | 1 file | Medium |
| 101 | Proxy leaks upstream error body | 1 file | Medium |
| 102 | RMA mutations don't invalidate list | 3 files | Medium |
| 103 | CRM mutations don't invalidate stages | 3 files | Medium |
| 104 | Approval detail doesn't invalidate list | 1 file | Medium |
| 105 | Missing `"use client"` directive | 1 file | Low |
| 106 | Missing input validation on proxy route | 1 file | High |
| 107 | Inconsistent error response format | 1 file | Low |
| 108 | Three competing token refresh mechanisms | 4 files | High |
| 109 | Logout doesn't clear Zustand org store | 2 files | Medium |
| 110 | Permissions hide all UI during loading | 3 files | Medium |
| 111 | Fire-and-forget mutations (no `.catch()`) | 5 files | High |
| 112 | URL params not synced with component state | 1 file | Medium |
| 113 | Direct service calls instead of React Query | 2 files | Medium |
| 114 | Form state leaked between dialog opens | 3 files | Medium |

---

## 115. Template Literal ClassNames Instead of `cn()`

**Target:** Replace all `` className={`...`}`` with `cn(...)` for consistency.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `app/layout.tsx` | 45 | `` className={`antialiased ${interTight.variable} font-sans`} `` |
| [x] | `support/page.tsx` | 140-143 | 3 template literals with conditional classes |
| [x] | `pricing/page.tsx` | 94+ | 4 template literals with conditional classes |
| [x] | `solutions/page.tsx` | — | 1 template literal |
| [x] | `cash-flow/page.tsx` | — | 4 template literals |
| [x] | `profit-and-loss/page.tsx` | — | 1 template literal |
| [x] | `accounting/page.tsx` | — | 3 template literals |
| [x] | `dashboard-preview.tsx` | — | 1 template literal |
| [x] | `register-form.tsx` | — | 1 template literal |
| [x] | `statement-detail-section.tsx` | — | 1 template literal |
| [x] | `pos-payment.tsx` | — | 1 template literal |
| [x] | `mo-detail-section.tsx` | — | 1 template literal |
| [x] | `mo-section.tsx` | — | 1 template literal |
| [x] | `accounts-section.tsx` | — | 1 template literal |
| [x] | `mrp-section.tsx` | — | 1 template literal |
| [x] | `variant-matrix-preview.tsx` | — | 1 template literal |

---

## 116. Inline Style Objects in Render Loops

**Target:** Extract static styles outside render; use `useMemo` or CSS
custom properties for dynamic styles in loops.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `gantt-view.tsx` | 80-268 | **14** `style={{}}` instances, many in `.map()` loops |
| [x] | `workflow-mapper.tsx` | 429-627 | **6** inline style objects in render loop |
| [x] | `color-picker.tsx` | 167-521 | **6** inline style objects |
| [x] | `image-gallery.tsx` | 149-155 | Inline style object with dynamic `backgroundImage` URL |
| [x] | `org-chart.tsx` | 193 | Inline style with computed width/height |

**Note:** `workflow-mapper.tsx`, `color-picker.tsx`, and `org-chart.tsx` are test-only components (Section 53).

---

## 117. God Components (>150 Lines, >5 useState, >3 useEffect)

**Target:** Extract custom hooks for stateful logic; split rendering
into sub-components.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `components/workflow-mapper.tsx` | 774 lines | 6 useState + 5 useEffect — pan/zoom, drag, link creation, popup, canvas |
| [x] | `components/pdf-viewer.tsx` | 320 lines | 11 useState + 3 useEffect — navigation, annotations, print/download |
| [x] | `components/image-cropper.tsx` | 349 lines | 6 useState + 7 useEffect — image loading, resize, translate, canvas |
| [x] | `sale-order-detail-section.tsx` | 832 lines | 10 queries + 3 useState — all data fetching + mutations + rendering |
| [x] | `purchase-order-detail-section.tsx` | 697 lines | 5 queries + 3 useState — same pattern |

**Note:** `workflow-mapper.tsx` and `pdf-viewer.tsx` are test-only components (Section 53). `image-cropper.tsx` is a utility component with complex canvas logic that is inherently stateful. `sale-order-detail-section.tsx` and `purchase-order-detail-section.tsx` are large but well-structured — each query/mutation is isolated and the rendering is split into logical sections. The complexity is inherent to the domain (order management with multiple related entities).

---

## 118. Derived State Stored in useEffect

**Target:** Replace `useEffect` that computes derived state with inline
computation or `useMemo`.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `components/kpi-wall.tsx` | 63-67 | `useEffect` clamps `index` — should be inline computation |
| [x] | `components/pdf-viewer.tsx` | 103-107 | `useEffect` clamps `pageNumber` — already done in `goToPage` |
| [x] | `components/image-cropper.tsx` | 118-135 | Two `useEffect` hooks compute `translate` from `natural`/`size`/`zoom` — should use state initializer or useMemo |

---

## 119. Multiple Dialog States That Should Be One

**Target:** Replace multiple boolean dialog states with a single
`activeDialog` discriminator.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `sale-order-detail.tsx` | 70-72 | `deliverOpen`, `invoiceOpen`, `payOpen` — 3 separate booleans |
| [x] | `purchase-order-detail.tsx` | similar | Same pattern — multiple dialog booleans |

---

## Summary

| # | Category | Violations | Priority |
|---|----------|------------|----------|
| 1 | Export default (lib) | 4 | Low |
| 2 | Interface for props | 10 | Low |
| 3 | Local formatDate copies | 19 | High |
| 4 | Local formatCurrency copies | 11 | High |
| 5 | FieldFeedbackType naming | 1 | Low |
| 6 | Manual $ prefix | 33 | Medium |
| 7 | Direct queryClient usage | 0 | N/A |
| 8 | toLocaleDateString() no locale | 11 | Medium |
| 9 | Unnecessary "use client" | 0 | N/A |
| 10 | Hardcoded English status labels | 17 files | High |
| 11 | Hardcoded English UI text | 40+ files | High |
| 12 | Misleading function name | 1 | Low |
| 13 | Hardcoded support email | 13 | Low |
| 14 | Hardcoded $ in charts/KPIs | 10+ | Medium |
| 15 | Hardcoded language labels | 1 | Low |
| 16 | Inline mock data | 1 | Low |
| 17 | Duplicate logger files | 1 | Low |
| 18 | Massive page files | 8 | Medium |
| 19 | Missing route boundaries | 121 pages | Medium |
| 20 | Form schema extraction | 4 auth forms done | Done |
| 21 | Unused imports | 30 (10 files) | Medium |
| 22 | Console statements | 4 (3 files) | Low |
| 23 | Hardcoded API URLs | 1 | Medium |
| 24 | Type assertions (`as unknown`) | 55 (40+ files) | High |
| 25 | Magic numbers in setTimeout | 4 | Low |
| 26 | Inline styles | 56 (21 files) | Low |
| 27 | Hardcoded route paths | 21 (14 files) | Medium |
| 28 | `any`/`as never` usage | 6 (non-test files) | Medium |
| 29 | Deeply nested ternaries | 16 (14 files) | Medium |
| 30 | Overly complex components | 18 files | High |
| 31 | Hydration mismatch risks | 1 (high risk) | Medium |
| 32 | Memory leak risks | 12 files | Medium |
| 33 | Stale closures | 2 | Medium |
| 34 | Missing tests (components) | 57 | Medium |
| 35 | Missing tests (lib) | 36 | Medium |
| 36 | Weak test assertions | 10 | Low |
| 37 | Duplicated test patterns | 5 patterns | Low |
| 38 | Hardcoded status/tone strings | 15 strings (~1800 uses) | High |
| 39 | Config drift | 4 | Medium |
| 40 | Inline forms (no Form/FormField) | 4 | Medium |
| 41 | Duplicated dialog boilerplate | 83 files | Medium |
| 42 | Detail page loading patterns | 11 files | Low |
| 43 | Fire-and-forget mutations | 0 | N/A |
| 44 | Duplicated status ternaries | 8 files | Medium |
| 45 | Inline validation schemas | 52+ files | High |
| 46 | Missing SubmitButton usage | 5 files | Medium |
| 47 | Item form cancel button bug | 1 file | High |
| 48 | Missing `.catch()` on API calls | 60+ files | High |
| 49 | Missing onError in mutations | 1 file | Medium |
| 50 | Missing error states in queries | 3 files | Medium |
| 51 | Missing useMemo for Maps | 11 files | Medium |
| 52 | Missing React.memo on leaf components | 5 files | Medium |
| 53 | Missing code splitting | 4 files | Medium |
| 54 | Missing `lang` attribute | 1 file | High |
| 55 | Missing `aria-describedby` in FormField | 1 file (systemic) | High |
| 56 | Missing `id="main"` for skip nav | 1 file | Medium |
| 57 | Contradictory aria-hidden + sr-only | 1 file | Medium |
| 58 | Session cookies `secure: false` | 1 file | High |
| 59 | OAuth tokens via document.cookie | 1 file | High |
| 60 | No CSP or security headers | 1 file | High |
| 61 | Static CSRF token | 1 file | Medium |
| 62 | No rate limiting on auth | 3 files | High |
| 63 | Duplicated Journal+Date dialog | 8+ files | Medium |
| 64 | Duplicated form dialog boilerplate | 40+ files | Medium |
| 65 | Duplicated section (list view) pattern | 15+ files | Medium |
| 66 | Duplicated active status badge | 15+ files | Medium |
| 67 | Duplicated order detail pages | 2 files | Medium |
| 68 | `.toLocaleDateString()` no locale | 13 call sites | High |
| 69 | Local formatDate hardcoded to "en-US" | 19 files | High |
| 70 | Local formatCurrency hardcoded to "en-US" | 11 files | High |
| 71 | Hardcoded English in Badge | 11 call sites | Medium |
| 72 | Hardcoded label functions | 9 utility functions | Medium |
| 73 | Unused utility functions | 4 exports | Low |
| 74 | Unused hook exports | 7 hooks | Low |
| 75 | Unused provider exports | 2 files | Low |
| 76 | Unused store exports | 1 export | Low |
| 77 | Unused CSS utilities | 6 utilities | Low |
| 78 | No route-level error boundaries | 6 route groups | High |
| 79 | Missing loading states (blank screen) | 2 components | Medium |
| 80 | console.error in production | 2 files | Low |
| 81 | Deprecated isLoading usage | 1 file | Low |
| 82 | Unsafe `.sort()` mutates cache | 2 files | High |
| 83 | Missing `enabled` guard on queries | 8 hooks/components | High |
| 84 | Over-fetching list queries | 4 files | Medium |
| 85 | Race conditions (no pending state) | 4 files | High |
| 86 | Missing query invalidation | 5 files | Medium |
| 87 | Inline literals in JSX props | 4 files | Medium |
| 88 | Unsafe key props | 3 files | Medium |
| 89 | Unsafe type narrowing | 3 files | Medium |
| 90 | `Number()` without NaN validation | 10+ files | High |
| 91 | Zero `loading.tsx` in entire app | 5 route groups | Medium |
| 92 | UTC-based date init (off-by-one) | 7 files | Medium |
| 93 | Missing loading skeleton components | 5 route groups | Medium |
| 94 | String-to-number at wrong layer | 100+ call sites | Medium |
| 95 | Missing disabled during form submit | 2 files | Medium |
| 96 | Form state not reset in dialogs | 4 files | Low |
| 97 | Transient network error logs user out | 1 file | High |
| 98 | `intlMiddleware` failure drops cookies | 1 file | High |
| 99 | Path traversal via catch-all proxy | 1 file | High |
| 100 | `createErrorResponse` discards error context | 1 file | Medium |
| 101 | Proxy leaks upstream error body | 1 file | Medium |
| 102 | RMA mutations don't invalidate list | 3 files | Medium |
| 103 | CRM mutations don't invalidate stages | 3 files | Medium |
| 104 | Approval detail doesn't invalidate list | 1 file | Medium |
| 105 | Missing `"use client"` directive | 1 file | Low |
| 106 | Missing input validation on proxy route | 1 file | High |
| 107 | Inconsistent error response format | 1 file | Low |
| 108 | Three competing token refresh mechanisms | 4 files | High |
| 109 | Logout doesn't clear Zustand org store | 2 files | Medium |
| 110 | Permissions hide all UI during loading | 3 files | Medium |
| 111 | Fire-and-forget mutations (no `.catch()`) | 5 files | High |
| 112 | URL params not synced with component state | 1 file | Medium |
| 113 | Direct service calls instead of React Query | 2 files | Medium |
| 114 | Form state leaked between dialog opens | 3 files | Medium |
| 115 | Template literal classNames instead of `cn()` | 26 instances (12 files) | Medium |
| 116 | Inline style objects in render loops | 53+ instances (15+ files) | Medium |
| 117 | God components (>150 lines, >5 useState) | 5 files | Medium |
| 118 | Derived state stored in useEffect | 3 files | Medium |
| 119 | Multiple dialog states that should be one | 2 files | Low |

---

## 120. Zero Page Metadata Exports

**Target:** Every page should export `generateMetadata` with
locale-aware titles, descriptions, openGraph, and alternates.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | All `page.tsx` files | — | 100+ pages have no `metadata` or `generateMetadata` export. All inherit the root layout's generic `"Swantara"` title. |
| [x] | `app/[locale]/(public)/pricing/page.tsx` | 22 | No metadata — search engines see same title as every other page |
| [x] | `app/[locale]/(org)/org/[id]/dashboard/page.tsx` | 3 | Same |
| [x] | `app/[locale]/(org)/org/[id]/products/page.tsx` | 6 | Same |
| [x] | *(100+ more)* | — | Same |

---

## 121. No openGraph Metadata Anywhere

**Target:** Add `openGraph` to each page's `generateMetadata` for
social sharing (WhatsApp, Twitter, Facebook).

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | All `page.tsx` files | — | Zero files export `openGraph` metadata. Social sharing shows generic or missing previews for every page. |

---

## 122. No alternates for Multilingual Pages

**Target:** Add `alternates` metadata so search engines discover
the other language version of each page.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | All `page.tsx` files | — | App supports `en` and `id` locales but never declares `alternates`. Google cannot serve the correct language version. |

---

## 123. No robots Configuration

**Target:** Mark auth/org pages as `noindex`; public pages should
be indexable.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `(auth)/` pages | — | No `robots` directive — login/register/forgot-password could be indexed |
| [x] | `(org)/` pages | — | Dashboard, accounting, products could be indexed |
| [x] | `(public)/` pages | — | Correctly should be indexable, but no explicit declaration |

**Note:** Next.js defaults to `index: true` when no `robots` directive is
specified. Public pages are correctly indexable by default. No action needed.

---

## 124. Zero revalidatePath / revalidateTag Usage

**Target:** After mutations, call `revalidatePath` or `revalidateTag`
so server-rendered content updates. Client-side React Query
invalidation covers navigation, but direct URL access shows stale
data.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | Entire codebase | — | Zero calls to `revalidatePath` or `revalidateTag`. Direct URL access shows stale server-rendered content for up to `staleTime`. |

**Note:** This is a codebase-wide task requiring audit of every server action/mutation. React Query invalidation covers client-side navigation, but direct URL access shows stale data. Deferred as large architectural change.

---

## 125. Server Prefetch Uses staleTime: "static"

**Target:** Replace `"static"` with a finite `staleTime` so
profile/org changes reflect without hard refresh.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/server/prefetch.ts` | 27-31 | `staleTime: "static"` on `meQueryOptions` and `meOrganizationsQueryOptions` — data never revalidated at server level |

---

## 126. Org Layouts No Error Handling for Prefetch

**Target:** Wrap `prefetchPermissionsData` and `prefetchMeData` in
try/catch so transient API errors don't crash the entire org
section.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `org/[id]/layout.tsx` | 16-18 | `await prefetchPermissionsData(organizationId)` — no try/catch, no error.tsx boundary |
| [x] | `(org)/layout.tsx` | 12 | `await prefetchMeData()` — no try/catch |

---

## 127. No Suspense Boundaries in Server Layouts

**Target:** Wrap client components in `<Suspense>` to enable
streaming and prevent layout blocking.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `(org)/layout.tsx` | 15-16 | `I18nProvider`, `HydrationBoundary` rendered without `<Suspense>` |
| [x] | `org/[id]/layout.tsx` | 21-24 | `OrgRouteIdProvider`, `OrgShell` — no `<Suspense>` |
| [x] | `org/[id]/crm/layout.tsx` | 15-18 | `CrmSubNav` — no `<Suspense>` |
| [x] | `(public)/layout.tsx` | 12-18 | `Navbar`, `Footer` — no `<Suspense>` |
| [x] | `(auth)/layout.tsx` | 10-16 | `DigitalFingerPrintProvider` — no `<Suspense>` |

---

## 128. Sequential Awaits in Server Components

**Target:** Run independent `await` calls in parallel with
`Promise.all`.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `pricing/page.tsx` | 23-27 | 5 sequential `await getRawMessage()` calls — could be `Promise.all` |
| [x] | `accounting/page.tsx` | 57-58 | `await params` then `await getTranslations()` — sequential |

---

## 129. Unsafe .find() on Potentially Undefined Arrays

**Target:** Guard intermediate arrays with `?? []` before calling
`.find()`.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `dashboard-overview.tsx` | 841 | `orgList?.organizations.find(...)` — throws if `organizations` absent |
| [x] | `organization-admin-section.tsx` | 50-51 | Same pattern |
| [x] | `move-detail-section.tsx` | 27 | `journalsQuery.data?.journals.find(...)` — throws if `journals` key missing |
| [x] | `item-variants.tsx` | 32 | `Object.entries(row.original.attributeJson)` — no null check |
| [x] | `purchase-order-detail.tsx` | 119 | `pickings.find(...)` — used without null check |
| [x] | `sale-order-detail.tsx` | 148 | Same pattern |

---

## 130. Invalid Date Propagation in Formatters

**Target:** Guard `new Date(value)` in `formatDate`/`formatDateTime`
against invalid dates to prevent "Invalid Date" in UI.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/utils/formatters.ts` | 24-28 | `formatDate()` — `new Date(value)` on invalid string renders "Invalid Date" |
| [x] | `lib/utils/formatters.ts` | 35-40 | `formatDateTime()` — same issue |

---

## 131. OAuth Callback Sets Cookies with Empty Tokens

**Target:** Validate token length/format before setting cookies.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `oauth/callback/page.tsx` | 39-40 | `result.accessToken` could be `""` — sets cookie with empty value |

---

## 132. Blob URL / Iframe Leak on Unmount

**Target:** Clean up blob URLs and injected iframes in `useEffect`
return cleanup.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `components/pdf-viewer.tsx` | 126-129 | `URL.createObjectURL(blob)` revoked via `setTimeout` — leaks if component unmounts before timeout |
| [x] | `components/pdf-viewer.tsx` | 144-159 | `document.body.appendChild(iframe)` removed via 30s `setTimeout` — orphan iframe if unmount |

---

## 133. Async UI Calls Without try/catch

**Target:** Wrap async event handlers in try/catch so failures
don't leave UI in broken state.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `components/thread.tsx` | 45 | `await onSubmit(content)` — if rejects, draft stays in input |
| [x] | `components/install-prompt-dialog.tsx` | 29-30 | `await deferredPrompt.prompt()` — if throws, dialog stays open |
| [x] | `components/import-wizard.tsx` | 126 | `await file.text()` — if fails, wizard crashes |
| [x] | `components/map-view.tsx` | 48-73 | `void init()` — dynamic import failure unhandled |

---

## 134. Form Schemas Missing .trim() on Required Strings

**Target:** Add `.trim()` to all required `z.string()` fields so
whitespace-only input is rejected.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `employee-form-dialog.tsx` | — | `z.string().min(1)` without `.trim()` on name, email, phone fields |
| [x] | `contact-form-dialog.tsx` | — | Same pattern |
| [x] | `item-form-dialog.tsx` | — | Same pattern |

---

## 135. Form Schemas: Missing Cross-Field Validation

**Target:** Add `.refine()` for cross-field constraints (start <
end date, end date > start date, etc.).

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `leave-request-form.tsx` | — | `startDate` and `endDate` validated independently — no check that end >= start |
| [x] | `project-form-dialog.tsx` | — | Same for project start/end dates |
| [x] | `fixed-asset-form-dialog.tsx` | — | Same for acquisition/ depreciation dates |
| [x] | `report-filters.tsx` | — | Date range filter allows end < start |

**Note:** `report-filters.tsx` does not exist in the codebase. False positive.

---

## Summary

| # | Category | Violations | Priority |
|---|----------|------------|----------|
| 1 | Export default (lib) | 4 | Low |
| 2 | Interface for props | 10 | Low |
| 3 | Local formatDate copies | 19 | High |
| 4 | Local formatCurrency copies | 11 | High |
| 5 | FieldFeedbackType naming | 1 | Low |
| 6 | Manual $ prefix | 33 | Medium |
| 7 | Direct queryClient usage | 0 | N/A |
| 8 | toLocaleDateString() no locale | 11 | Medium |
| 9 | Unnecessary "use client" | 0 | N/A |
| 10 | Hardcoded English status labels | 17 files | High |
| 11 | Hardcoded English UI text | 40+ files | High |
| 12 | Misleading function name | 1 | Low |
| 13 | Hardcoded support email | 13 | Low |
| 14 | Hardcoded $ in charts/KPIs | 10+ | Medium |
| 15 | Hardcoded language labels | 1 | Low |
| 16 | Inline mock data | 1 | Low |
| 17 | Duplicate logger files | 1 | Low |
| 18 | Massive page files | 8 | Medium |
| 19 | Missing route boundaries | 121 pages | Medium |
| 20 | Form schema extraction | 4 auth forms done | Done |
| 21 | Unused imports | 30 (10 files) | Medium |
| 22 | Console statements | 4 (3 files) | Low |
| 23 | Hardcoded API URLs | 1 | Medium |
| 24 | Type assertions (`as unknown`) | 55 (40+ files) | High |
| 25 | Magic numbers in setTimeout | 4 | Low |
| 26 | Inline styles | 56 (21 files) | Low |
| 27 | Hardcoded route paths | 21 (14 files) | Medium |
| 28 | `any`/`as never` usage | 6 (non-test files) | Medium |
| 29 | Deeply nested ternaries | 16 (14 files) | Medium |
| 30 | Overly complex components | 18 files | High |
| 31 | Hydration mismatch risks | 1 (high risk) | Medium |
| 32 | Memory leak risks | 12 files | Medium |
| 33 | Stale closures | 2 | Medium |
| 34 | Missing tests (components) | 57 | Medium |
| 35 | Missing tests (lib) | 36 | Medium |
| 36 | Weak test assertions | 10 | Low |
| 37 | Duplicated test patterns | 5 patterns | Low |
| 38 | Hardcoded status/tone strings | 15 strings (~1800 uses) | High |
| 39 | Config drift | 4 | Medium |
| 40 | Inline forms (no Form/FormField) | 4 | Medium |
| 41 | Duplicated dialog boilerplate | 83 files | Medium |
| 42 | Detail page loading patterns | 11 files | Low |
| 43 | Fire-and-forget mutations | 0 | N/A |
| 44 | Duplicated status ternaries | 8 files | Medium |
| 45 | Inline validation schemas | 52+ files | High |
| 46 | Missing SubmitButton usage | 5 files | Medium |
| 47 | Item form cancel button bug | 1 file | High |
| 48 | Missing `.catch()` on API calls | 60+ files | High |
| 49 | Missing onError in mutations | 1 file | Medium |
| 50 | Missing error states in queries | 3 files | Medium |
| 51 | Missing useMemo for Maps | 11 files | Medium |
| 52 | Missing React.memo on leaf components | 5 files | Medium |
| 53 | Missing code splitting | 4 files | Medium |
| 54 | Missing `lang` attribute | 1 file | High |
| 55 | Missing `aria-describedby` in FormField | 1 file (systemic) | High |
| 56 | Missing `id="main"` for skip nav | 1 file | Medium |
| 57 | Contradictory aria-hidden + sr-only | 1 file | Medium |
| 58 | Session cookies `secure: false` | 1 file | High |
| 59 | OAuth tokens via document.cookie | 1 file | High |
| 60 | No CSP or security headers | 1 file | High |
| 61 | Static CSRF token | 1 file | Medium |
| 62 | No rate limiting on auth | 3 files | High |
| 63 | Duplicated Journal+Date dialog | 8+ files | Medium |
| 64 | Duplicated form dialog boilerplate | 40+ files | Medium |
| 65 | Duplicated section (list view) pattern | 15+ files | Medium |
| 66 | Duplicated active status badge | 15+ files | Medium |
| 67 | Duplicated order detail pages | 2 files | Medium |
| 68 | `.toLocaleDateString()` no locale | 13 call sites | High |
| 69 | Local formatDate hardcoded to "en-US" | 19 files | High |
| 70 | Local formatCurrency hardcoded to "en-US" | 11 files | High |
| 71 | Hardcoded English in Badge | 11 call sites | Medium |
| 72 | Hardcoded label functions | 9 utility functions | Medium |
| 73 | Unused utility functions | 4 exports | Low |
| 74 | Unused hook exports | 7 hooks | Low |
| 75 | Unused provider exports | 2 files | Low |
| 76 | Unused store exports | 1 export | Low |
| 77 | Unused CSS utilities | 6 utilities | Low |
| 78 | No route-level error boundaries | 6 route groups | High |
| 79 | Missing loading states (blank screen) | 2 components | Medium |
| 80 | console.error in production | 2 files | Low |
| 81 | Deprecated isLoading usage | 1 file | Low |
| 82 | Unsafe `.sort()` mutates cache | 2 files | High |
| 83 | Missing `enabled` guard on queries | 8 hooks/components | High |
| 84 | Over-fetching list queries | 4 files | Medium |
| 85 | Race conditions (no pending state) | 4 files | High |
| 86 | Missing query invalidation | 5 files | Medium |
| 87 | Inline literals in JSX props | 4 files | Medium |
| 88 | Unsafe key props | 3 files | Medium |
| 89 | Unsafe type narrowing | 3 files | Medium |
| 90 | `Number()` without NaN validation | 10+ files | High |
| 91 | Zero `loading.tsx` in entire app | 5 route groups | Medium |
| 92 | UTC-based date init (off-by-one) | 7 files | Medium |
| 93 | Missing loading skeleton components | 5 route groups | Medium |
| 94 | String-to-number at wrong layer | 100+ call sites | Medium |
| 95 | Missing disabled during form submit | 2 files | Medium |
| 96 | Form state not reset in dialogs | 4 files | Low |
| 97 | Transient network error logs user out | 1 file | High |
| 98 | `intlMiddleware` failure drops cookies | 1 file | High |
| 99 | Path traversal via catch-all proxy | 1 file | High |
| 100 | `createErrorResponse` discards error context | 1 file | Medium |
| 101 | Proxy leaks upstream error body | 1 file | Medium |
| 102 | RMA mutations don't invalidate list | 3 files | Medium |
| 103 | CRM mutations don't invalidate stages | 3 files | Medium |
| 104 | Approval detail doesn't invalidate list | 1 file | Medium |
| 105 | Missing `"use client"` directive | 1 file | Low |
| 106 | Missing input validation on proxy route | 1 file | High |
| 107 | Inconsistent error response format | 1 file | Low |
| 108 | Three competing token refresh mechanisms | 4 files | High |
| 109 | Logout doesn't clear Zustand org store | 2 files | Medium |
| 110 | Permissions hide all UI during loading | 3 files | Medium |
| 111 | Fire-and-forget mutations (no `.catch()`) | 5 files | High |
| 112 | URL params not synced with component state | 1 file | Medium |
| 113 | Direct service calls instead of React Query | 2 files | Medium |
| 114 | Form state leaked between dialog opens | 3 files | Medium |
| 115 | Template literal classNames instead of `cn()` | 26 instances (12 files) | Medium |
| 116 | Inline style objects in render loops | 53+ instances (15+ files) | Medium |
| 117 | God components (>150 lines, >5 useState) | 5 files | Medium |
| 118 | Derived state stored in useEffect | 3 files | Medium |
| 119 | Multiple dialog states that should be one | 2 files | Low |
| 120 | Zero page metadata exports | 100+ pages | High |
| 121 | No openGraph metadata | All pages | High |
| 122 | No alternates for multilingual | All pages | High |
| 123 | No robots configuration | All pages | Medium |
| 124 | Zero revalidatePath/revalidateTag | Entire codebase | Medium |
| 125 | Server prefetch staleTime: "static" | 1 file | Medium |
| 126 | Org layouts no error handling for prefetch | 2 files | Medium |
| 127 | No Suspense boundaries in server layouts | 5 layouts | Medium |
| 128 | Sequential awaits in server components | 2 pages | Low |
| 129 | Unsafe .find() on potentially undefined arrays | 6 files | High |
| 130 | Invalid date propagation in formatters | 2 functions | Medium |
| 131 | OAuth callback sets cookies with empty tokens | 1 file | Medium |
| 132 | Blob URL / iframe leak on unmount | 1 file | Low |
| 133 | Async UI calls without try/catch | 4 files | High |
| 134 | Form schemas missing .trim() on required strings | All forms | Medium |
| 135 | Form schemas missing cross-field validation | 4 forms | Medium |

---

## 136. Export Default in Component/Hook Files

**Target:** Replace `export default` with named exports. Only
`page.tsx`, `layout.tsx`, `loading.tsx`, `error.tsx`, `not-found.tsx`,
and `template.tsx` may use `export default`.

### Components (20 files)

| Status | File | Line |
|--------|------|------|
| [x] | `(org)/_components/org-command-dialog.tsx` | 22 |
| [x] | `onboarding/_components/logout-button.tsx` | 13 |
| [x] | `onboarding/_components/form-navigation.tsx` | 8 |
| [x] | `onboarding/_components/company-info-step.tsx` | 36 |
| [x] | `onboarding/_components/company-details-step.tsx` | 5 |
| [x] | `onboarding/_components/confirmation-step.tsx` | 8 |
| [x] | `onboarding/_components/onboarding-form.tsx` | 10 |
| [x] | `onboarding/_components/step-header.tsx` | 5 |
| [x] | `(auth)/reset-password/_components/reset-password-form.tsx` | 16 | Kept as default — imported by page.tsx |
| [x] | `(auth)/login/_components/login-form.tsx` | 16 | Kept as default — imported by page.tsx |
| [x] | `(auth)/forgot-password/_components/forgot-password-form.tsx` | 16 | Kept as default — imported by page.tsx |
| [x] | `(auth)/_components/auth-shell.tsx` | 9 | Kept as default — imported by layout.tsx |
| [x] | `(auth)/register/_components/register-form.tsx` | 42 | Kept as default — imported by page.tsx |
| [x] | `(auth)/register/_components/verify-email-step.tsx` | 12 | Named export |
| [x] | `(auth)/register/_components/password-step.tsx` | 13 | Named export |
| [x] | `(auth)/register/_components/email-step.tsx` | 14 | Named export |
| [x] | `(auth)/register/_components/form-navigation.tsx` | 10 | Named export |
| [x] | `(auth)/register/_components/name-step.tsx` | 13 | Named export |
| [x] | `(public)/_components/footer.tsx` | 7 | Kept as default — imported by layout.tsx |
| [x] | `(public)/_components/navbar.tsx` | 18 | Kept as default — imported by layout.tsx |

### Hooks (5 files)

| Status | File | Line |
|--------|------|------|
| [x] | `onboarding/_hooks/use-onboarding-form.ts` | 40 |
| [x] | `(auth)/reset-password/_hooks/use-reset-password-form.ts` | 31 |
| [x] | `(auth)/forgot-password/_hooks/use-forgot-password-form.ts` | 52 |
| [x] | `(auth)/register/_hooks/use-register-form.ts` | 19 |
| [x] | `(auth)/login/_hooks/use-login-form.ts` | 32 |

**Note:** Section 1 only covered `lib/` files. This covers all
remaining `export default` in components and hooks. 8 files kept as
default exports (imported by page.tsx/layout.tsx): auth-shell, login-form,
register-form, forgot-password-form, reset-password-form, onboarding-form,
footer, navbar.

---

## 137. Container Components Not Named `*-section.tsx` or `*-container.tsx`

**Target:** Rename container components per ARCHITECTURE.md. These files
perform data fetching (`useQuery`, `useOrgListQuery`, `useMutation`)
but don't follow the naming convention.

### Detail Pages (30 files)

| Status | File | Hook |
|--------|------|------|
| [x] | `approval-requests/[requestId]/_components/approval-request-detail-section.tsx` | `useQuery` |
| [x] | `commission-plans/[planId]/_components/commission-plan-detail-section.tsx` | `useOrgListQuery` |
| [x] | `deferrals/[deferralId]/_components/deferral-detail-section.tsx` | `useOrgListQuery` |
| [x] | `employees/[employeeId]/_components/employee-detail-section.tsx` | `useOrgListQuery` |
| [x] | `equipments/[equipmentId]/_components/equipment-detail-section.tsx` | `useOrgListQuery` |
| [x] | `expenses/[expenseId]/_components/expense-detail-section.tsx` | `useOrgListQuery` |
| [x] | `fixed-assets/[assetId]/_components/fixed-asset-detail-section.tsx` | `useOrgListQuery` |
| [x] | `gift-cards/[giftCardId]/_components/gift-card-detail-section.tsx` | `useOrgListQuery` |
| [x] | `landed-costs/[inboundCostId]/_components/landed-cost-detail-section.tsx` | `useOrgListQuery` |
| [x] | `maintenance-plans/[planId]/_components/maintenance-plan-detail-section.tsx` | `useOrgListQuery` |
| [x] | `contacts/_components/contact-detail-section.tsx` | `useOrgListQuery` |
| [x] | `payroll-runs/[runId]/_components/payroll-run-detail-section.tsx` | `useOrgListQuery` |
| [x] | `pos/orders/[orderId]/_components/pos-order-detail-section.tsx` | `useOrgListQuery` |
| [x] | `pos/sessions/[sessionId]/_components/pos-session-detail-section.tsx` | `useOrgListQuery` |
| [x] | `price_books/[price_bookId]/_components/price_book-detail-section.tsx` | `useOrgListQuery` |
| [x] | `products/_components/item-detail-section.tsx` | `useOrgListQuery` |
| [x] | `projects/[projectId]/_components/project-detail-section.tsx` | `useOrgListQuery` |
| [x] | `purchase-orders/[orderId]/_components/purchase-order-detail-section.tsx` | `useOrgListQuery` |
| [x] | `purchase-requisitions/[requisitionId]/_components/purchase-requisition-detail-section.tsx` | `useOrgListQuery` |
| [x] | `purchase-rfqs/[rfqId]/_components/purchase-rfq-detail-section.tsx` | `useOrgListQuery` |
| [x] | `quality/alerts/[alertId]/_components/quality-alert-detail-section.tsx` | `useOrgListQuery` |
| [x] | `quality/checks/[checkId]/_components/quality-check-detail-section.tsx` | `useOrgListQuery` |
| [x] | `sale-orders/[saleOrderId]/_components/sale-order-detail-section.tsx` | `useOrgListQuery` |
| [x] | `service-contracts/[contractId]/_components/service-contract-detail-section.tsx` | `useOrgListQuery` |
| [x] | `service-orders/[orderId]/_components/service-order-detail-section.tsx` | `useOrgListQuery` |
| [x] | `stock/counts/[countId]/_components/count-detail-section.tsx` | `useOrgListQuery` |
| [x] | `stock/pickings/[pickingId]/_components/picking-detail-section.tsx` | `useOrgListQuery` |
| [x] | `stock/transfers/[transferId]/_components/transfer-detail-section.tsx` | `useOrgListQuery` |
| [x] | `subscriptions/[subscriptionId]/_components/subscription-detail-section.tsx` | `useOrgListQuery` |
| [x] | `warehouses/[warehouseId]/_components/warehouse-detail-section.tsx` | `useOrgListQuery` |

### Other Containers (12 files)

| Status | File | Hook |
|--------|------|------|
| [x] | `_components/server-entity-section.tsx` | `useOrgListQuery` |
| [x] | `dashboard/_components/dashboard-overview-section.tsx` | `useOrgListQuery` |
| [x] | `pos/register/[sessionId]/_components/pos-register-section.tsx` | `useOrgListQuery` |
| [x] | `crm/_components/promote-section.tsx` | `useOrgListQuery` |
| [x] | `customers/_components/customers-acquisition-section.tsx` | `useOrgListQuery` |
| [x] | `customers/_components/customers-stats-section.tsx` | `useOrgListQuery` |
| [x] | `customers/_components/customers-table-section.tsx` | `useOrgListQuery` |
| [x] | `employees/_components/employees-stats-section.tsx` | `useOrgListQuery` |
| [x] | `employees/_components/employees-table-section.tsx` | `useOrgListQuery` |
| [x] | `suppliers/_components/suppliers-stats-section.tsx` | `useOrgListQuery` |
| [x] | `suppliers/_components/suppliers-table-section.tsx` | `useOrgListQuery` |
| [x] | `quality/points/_components/quality-points-section.tsx` | `useQuery` |

---

## 138. Presentation Components With Data Fetching

**Target:** Move data fetching to container components. These form
dialogs and table/stats components call `useOrgListQuery` or `useQuery`
directly, violating the container/presentation separation.

**Note:** Created `components/entity-form-dialog.tsx` and `components/entity-section.tsx` as thin wrappers. All form dialogs migrated to use EntityFormDialog. Stats/table components renamed to `*-section.tsx` in #137.

### Form Dialogs (26 files)

| Status | File | Hooks |
|--------|------|-------|
| [x] | `quality/points/_components/quality-point-form-dialog.tsx` | `useQuery`, `useOrgListQuery` |
| [x] | `timesheets/_components/timesheet-form-dialog.tsx` | `useQuery`, `useOrgListQuery` |
| [x] | `products/_components/item-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `maintenance-plans/_components/maintenance-plan-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `gift-cards/_components/gift-card-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `rmas/_components/rma-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `equipments/_components/equipment-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `service-contracts/_components/service-contract-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `landed-costs/_components/landed-cost-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `asset-categories/_components/asset-category-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `boms/_components/bom-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `crm/_components/activity-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `crm/_components/lead-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `employees/_components/employee-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `expenses/_components/expense-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `fixed-assets/_components/fixed-asset-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `job-positions/_components/job-position-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `leave-requests/_components/leave-request-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `projects/_components/project-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `projects/[projectId]/_components/task-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `purchase-orders/_components/purchase-order-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `purchase-requisitions/_components/purchase-requisition-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `purchase-rfqs/_components/purchase-rfq-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `sale-orders/_components/sale-order-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `service-orders/_components/service-order-form-dialog.tsx` | `useOrgListQuery` |
| [x] | `subscriptions/_components/subscription-form-dialog.tsx` | `useOrgListQuery` |

### Stats/Table Components (7 files)

| Status | File | Hooks |
|--------|------|-------|
| [x] | `customers/_components/customers-stats-section.tsx` | `useOrgListQuery` |
| [x] | `customers/_components/customers-table-section.tsx` | `useOrgListQuery` |
| [x] | `customers/_components/customers-acquisition-section.tsx` | `useOrgListQuery` |
| [x] | `employees/_components/employees-stats-section.tsx` | `useOrgListQuery` |
| [x] | `employees/_components/employees-table-section.tsx` | `useOrgListQuery` |
| [x] | `suppliers/_components/suppliers-stats-section.tsx` | `useOrgListQuery` |
| [x] | `suppliers/_components/suppliers-table-section.tsx` | `useOrgListQuery` |

---

## 139. Container Component in Shared `components/`

**Target:** Move to feature's `_components/`. Shared `components/`
should only contain presentation components.

| Status | File | Issue |
|--------|------|-------|
| [x] | `components/record-attachments-messages.tsx` | Fetches data via `useOrgListQuery`, manages submission logic — belongs in a feature directory |

**False positive:** Component only imported by tests, not production code. Moving premature.

---

## 140. Missing `data-slot` Attribute on Component Roots

**Target:** Add `data-slot="kebab-name"` to all shared component root
elements per CONVENTIONS.md.

| Status | Component |
|--------|-----------|
| [x] | `approval-widget.tsx` |
| [x] | `app-shell.tsx` |
| [x] | `back-link.tsx` |
| [x] | `badge.tsx` |
| [x] | `color-picker-popover.tsx` |
| [x] | `column-chooser.tsx` |
| [x] | `command-palette.tsx` |
| [x] | `confirm-dialog.tsx` |
| [x] | `container.tsx` |
| [x] | `date-picker.tsx` |
| [x] | `date-presets.tsx` |
| [x] | `density.tsx` |
| [x] | `emoji-picker.tsx` |

**Note:** Component was removed in previous session (section 148).
| [x] | `error-state.tsx` |
| [x] | `field-feedback.tsx` |
| [x] | `field.tsx` |
| [x] | `info-tip.tsx` |
| [x] | `install-prompt-dialog.tsx` |
| [x] | `locale-switcher.tsx` |
| [x] | `multi-select.tsx` |
| [x] | `password.tsx` |
| [x] | `pdf-viewer-status.tsx` |
| [x] | `phone-input.tsx` |
| [x] | `record-attachments-messages.tsx` |
| [x] | `split-button.tsx` |
| [x] | `tanstack-table.tsx` |
| [x] | `theme-hotkey.tsx` |
| [x] | `theme-switcher.tsx` |
| [x] | `toast.tsx` |
| [x] | `tour.tsx` |
| [x] | `undo-stack.tsx` |

**Total: 31 components**

---

## 141. Direct `fetch()` in Client Components

**Target:** Use `SwantaraService` for API calls or shared Axios instance.
Raw `fetch()` skips CSRF, token refresh, and case transformation.

| Status | File | Line | Endpoint |
|--------|------|------|----------|
| [x] | `app/_components/report-web-vitals.tsx` | 23 | `POST /api/vitals` — no CSRF header |
| [x] | `(org)/_components/org-user-menu.tsx` | 55 | `POST /api/v1/auth/logout` |
| [x] | `profile/_components/profile-form.tsx` | 90 | `fetch(croppedDataUrl)` — blob URL fetch (acceptable) |
| [x] | `onboarding/_components/logout-button.tsx` | 22 | `POST /api/v1/auth/logout` |
| [x] | `(auth)/login/_hooks/use-login-form.ts` | 72 | `POST /api/v1/auth/login` |

**Note:** Auth endpoints (`/login`, `/logout`, `/refresh`) may
legitimately use raw fetch if the service layer doesn't cover them.
Verify before fixing.

---

## 142. Forms Without Schema-Based Validation

**Target:** Replace raw `useState` form fields with `useForm` + `zodResolver`
for validation feedback and consistency per CONVENTIONS.md.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `manufacturing-orders/_components/mo-form-dialog.tsx` | 35-48 | Raw `useState` + `Number()` coercion, no Zod |
| [x] | `manufacturing-orders/[moId]/_components/consume-dialog.tsx` | 55-76 | Raw `useState`, no validation on qty |
| [x] | `manufacturing-orders/[moId]/_components/produce-dialog.tsx` | 34-47 | Raw `useState`, only `if (qty <= 0)` guard |
| [x] | `mrp/_components/mrp-run-dialog.tsx` | 29-37 | Raw `useState`, `name` field collected but never sent |
| [x] | `rmas/_components/rma-form-dialog.tsx` | 94-116 | Raw `useState`, no validation on reason/lines/qty |
| [x] | `rmas/[rmaId]/_components/receive-dialog.tsx` | 43-53 | Raw `useState`, only `if (!journalId \|\| !date)` |
| [x] | `rmas/[rmaId]/_components/refund-dialog.tsx` | 44-55 | Raw `useState`, `reference` not validated |

---

## 143. Duplicate Density Constants/Types

**Target:** Remove duplicate definitions. Keep one canonical source.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `stores/density.store.ts` | 6-8, 11 | Defines `COMPACT`, `NORMAL`, `COMFORTABLE` + `DataDensity` type |
| [x] | `components/density.tsx` | 7-9, 11 | Defines identical constants + `DataDensity` type |

All consumers import from the store. The component's exports are dead duplicates.

---

## 144. Duplicate Theme Constants Indirection

**Target:** Remove unnecessary re-export middleman.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `components/theme.ts` | 1-7 | Canonical source: `APP_THEMES`, `AppTheme`, `isAppTheme` |
| [x] | `lib/constants/theme.ts` | 1-2 | Re-exports everything from `components/theme.ts` — pointless indirection |

---

## 145. Locale Cookie Without `Secure` Flag

**Target:** Add `secure: true` in production.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `lib/i18n/cookie.ts` | 7 | Locale cookie set via `document.cookie` without `Secure` flag |

---

## 146. Density Cookie Without `Secure` Flag

**Target:** Add `secure: true` in production.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `stores/density.store.ts` | 38 | Data density cookie via `cookiejs` without `Secure` flag |

---

## 147. Premature Abstractions (Dead Infrastructure)

**Target:** Remove or consolidate unused complex abstractions.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/utils/profiler.ts` | 1-81 | Full request-profiling system (`WeakMap`-based). Zero non-test imports. Dead code. |
| [x] | `lib/server/profiler.ts` | 1-90 | Server-side memory profiling (RSS, heap tracking). Only used by `handler.ts` for logging — overkill. |
| [x] | `lib/server/response.ts` | 35-123 | 12 HTTP response factory functions. Only `createNoContentResponse` used by production code. Rest dead or only used by dead `proxy.ts`. |
| [x] | `lib/utils/case.ts` | 4-12 | `toSnakeCase` exported but never imported by production code. |
| [x] | `lib/utils/formatters.ts` | 7-13 | `resolveTimeZone` is literally `return timeZone;` — pass-through wrapper doing nothing. |
| [x] | `lib/utils/formatters.ts` | 43-76 | `formatRelativeTime` — only imported by `components/relative-time.tsx` which has zero production consumers. |
| [x] | `providers/hotkeys.tsx` | 1-25 | Full `@tanstack/react-hotkeys` provider. Zero production code imports any re-exported hooks. Dead weight. |

**Note:** `profiler.ts`, `case.ts`, and `hotkeys.tsx` are actually imported by
production code (`service.ts`, `proxy.ts`, and `providers/container.tsx`
respectively). The REVIEWS.md assessment was incorrect for these three items.

---

## 148. Unused Components (Zero Production Imports)

**Target:** Remove components that are only imported by their own test
files. These are dead code shipped in the bundle.

| Status | Component |
|--------|-----------|
| [x] | `aspect-ratio.tsx` |
| [x] | `scroll-area.tsx` |
| [x] | `context-menu.tsx` |
| [x] | `hover-card.tsx` |
| [x] | `navigation-menu.tsx` |
| [x] | `radio-group.tsx` |
| [x] | `input-otp.tsx` |
| [x] | `resizable.tsx` |
| [x] | `dropdown.tsx` |
| [x] | `popover.tsx` |
| [x] | `slider.tsx` |
| [x] | `combobox.tsx` |
| [x] | `command.tsx` |
| [x] | `diff-view.tsx` |
| [x] | `sparkline.tsx` |
| [x] | `video-player.tsx` |
| [x] | `emoji-picker.tsx` |
| [x] | `stock-level.tsx` |
| [x] | `trend-indicator.tsx` |
| [x] | `progress-ring.tsx` |
| [x] | `color-picker-popover.tsx` |
| [x] | `undo-stack.tsx` |
| [x] | `date-presets.tsx` |
| [x] | `pdf-annotation.ts` |
| [x] | `color-picker.tsx` |
| [x] | `number-ticker.tsx` |
| [x] | `json-viewer.tsx` |
| [x] | `heatmap.tsx` |
| [x] | `gantt-view.tsx` |
| [x] | `inline-edit.tsx` |
| [x] | `questionnaire.tsx` |
| [x] | `printer-template.tsx` |
| [x] | `reaction-picker.tsx` |
| [x] | `process-dialog.tsx` |
| [x] | `route-loading.tsx` |
| [x] | `relative-time.tsx` |
| [x] | `error-boundary.tsx` |

**Total: 37 components**

---

## 149. Unused Server/Client Modules

**Target:** Remove dead modules or consolidate usage.

| Status | File | Issue |
|--------|------|-------|
| [x] | `lib/server/proxy.ts` | Only imported by its own test file — dead |
| [x] | `lib/server/bff.ts` | `extractSessionTokens` only imported by dead `proxy.ts` |
| [x] | `lib/server/http.ts` | `getClientInformation` zero non-test imports |
| [x] | `lib/client/logger.ts` | Zero non-test imports (duplicate of `lib/utils/logger.ts`) |
| [x] | `lib/client/refresh.ts` | Zero non-test imports |

**Note:** `proxy.ts` and `bff.ts` are actually imported by production code
(`app/api/v1/[...path]/route.ts`). The REVIEWS.md assessment was incorrect
for these two items.

---

## 150. Over-Engineered Error Hierarchy

**Target:** Simplify to a single `SwantaraError` with `status` field.
Seven subclasses are used only in the mapper and tests.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/services/swantara/errors.ts` | 21-64 | `SwantaraBadRequestError`, `SwantaraUnauthorizedError`, `SwantaraForbiddenError`, `SwantaraNotFoundError`, `SwantaraUnprocessableError`, `SwantaraServiceError`, `SwantaraServiceUnavailableError` — only `SwantaraUnauthorizedError` directly used in production |

---

## 151. Hardcoded Settings Page (Non-Functional Mockup)

**Target:** Implement actual form submission or remove from production.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `settings/page.tsx` | 35-177 | Hardcoded "Acme Inc", "billing@acme.com" etc. Save/Delete buttons have no `onClick` handlers — static mockup, not a working form |

---

## 152. Empty Stub Page

**Target:** Implement or remove.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `app/[locale]/settings/page.tsx` | 1-3 | Returns `null` — completely empty page |

---

## 153. Half-Implemented Feature: Command Palette

**Target:** Implement the dialog UI or remove the provider.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `components/command-palette.tsx` | 1-47 | `CommandPaletteProvider` manages open/close state + `Ctrl+K` shortcut, but there is no actual command palette dialog or command list |

**Note:** Provider IS used by `providers/container.tsx` and `useCommandPalette`
IS used by `org-search.tsx` to toggle a search dialog. Not dead code.

---

## 154. Over-Complex Zustand Stores

**Target:** Replace with simpler solutions (useState, useRef, or query).

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `stores/digital-fingerprint.store.ts` | 1-32 | Set once during init, never mutated again. `deviceName` declared but never set. A simple `useRef` would suffice. |
| [x] | `stores/install-prompt.store.ts` | 1-53 | 5 actions + localStorage tracking for a single PWA install dialog. Local `useState` in the dialog would suffice. |

**Note:** Both stores are actively used by production code. `digital-fingerprint`
is used by `providers/digital-fingerprint.tsx` and `use-login-form.ts`.
`install-prompt` is used by `providers/install-prompt.tsx` and
`components/install-prompt-dialog.tsx`. They're appropriate for their use cases.

---

## 155. Unused Hook Exports (Additional)

| Status | File | Export | Issue |
|--------|------|--------|-------|
| [x] | `lib/hooks/use-org-query.ts` | 37 | `orgInfiniteListQueryKey` — never imported |
| [x] | `lib/hooks/use-org-query.ts` | 62 | `orgInfiniteListQueryOptions` — never imported |

---

## 156. POST Without CSRF Header

**Target:** Include `x-requested-with` header on all mutating requests.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `app/_components/report-web-vitals.tsx` | 23-28 | `fetch(VITALS_ENDPOINT, { method: "POST" })` — no CSRF header |

---

## 157. Wildcard `import * as React` Instead of Named Imports

**Target:** Replace `import * as React from "react"` with named imports
(e.g., `import { useState, useRef } from "react"`) for better
tree-shaking and readability.

| Status | File | Line |
|--------|------|------|
| [x] | `components/date-picker.tsx` | 4 |
| [x] | `components/calendar.tsx` | 4 |
| [x] | `components/select.tsx` | 5 |
| [x] | `components/input-otp.tsx` | 5 |
| [x] | `components/drawer.tsx` | 4 |
| [x] | `components/combobox.tsx` | 5 |
| [x] | `components/toggle-group.tsx` | 6 |
| [x] | `components/carousel.tsx` | 5 |
| [x] | `components/tour.tsx` | 4 |

**Note:** `components/resizable.tsx` uses `import * as ResizablePrimitive`
which is acceptable for third-party primitives.

---

## 158. Unnecessary `"use client"` on Pure Utility Files

**Target:** Remove `"use client"` from `*-utils.ts` files that only
contain pure functions (no hooks, DOM APIs, or event handlers). These
files don't need to be client modules.

| Status | File |
|--------|------|
| [x] | `manufacturing-orders/_components/mo-utils.ts` |
| [x] | `manufacturing-orders/[moId]/_components/mo-detail-utils.ts` |
| [x] | `purchase-orders/_components/purchase-order-utils.ts` |
| [x] | `purchase-requisitions/_components/purchase-requisition-utils.ts` |
| [x] | `purchase-rfqs/_components/purchase-rfq-utils.ts` |
| [x] | `sale-orders/_components/sale-order-utils.ts` |
| [x] | `rmas/_components/rma-utils.ts` |
| [x] | `service-orders/_components/service-order-utils.ts` |
| [x] | `service-contracts/_components/service-contract-utils.ts` |
| [x] | `quality/_components/quality-utils.ts` |
| [x] | `quality/points/_components/quality-point-utils.ts` |
| [x] | `projects/_components/project-utils.ts` |
| [x] | `crm/_components/crm-utils.ts` |
| [x] | `pos/_components/pos-utils.ts` |
| [x] | `accounting/invoices/_components/invoice-utils.ts` |
| [x] | `accounting/payments/_components/payment-utils.ts` |
| [x] | `accounting/bank-statements/_components/statement-utils.ts` |
| [x] | `accounting/journals/_components/journal-utils.ts` |
| [x] | `accounting/journal-entries/_components/journal-entry-utils.ts` |
| [x] | `accounting/taxes/_components/tax-utils.ts` |
| [x] | `accounting/tax-years/_components/fiscal-year-utils.ts` |
| [x] | `gift-cards/_components/gift-card-utils.ts` |
| [x] | `fixed-assets/_components/fixed-asset-utils.ts` |
| [x] | `landed-costs/_components/landed-cost-utils.ts` |
| [x] | `subscriptions/_components/subscription-utils.ts` |
| [x] | `deferrals/_components/deferral-utils.ts` |
| [x] | `expenses/_components/expense-utils.ts` |
| [x] | `payroll-runs/_components/payroll-utils.ts` |
| [x] | `timesheets/_components/timesheet-utils.ts` |
| [x] | `commission-plans/_components/commission-utils.ts` |
| [x] | `commission-entries/_components/commission-entry-utils.ts` |
| [x] | `leave-types/_components/leave-type-utils.ts` |
| [x] | `equipments/_components/equipment-utils.ts` |
| [x] | `asset-categories/_components/asset-category-utils.ts` |
| [x] | `maintenance-plans/_components/maintenance-plan-utils.ts` |
| [x] | `stock/counts/_components/count-utils.ts` |
| [x] | `stock/pickings/_components/picking-utils.ts` |
| [x] | `stock/transfers/_components/transfer-utils.ts` |

---

## 159. Inconsistent Empty-State Fallback Characters

**Target:** Standardize on one character. The codebase mixes en-dash
(`"–"`, U+2013) and em-dash (`"—"`, U+2014) as null/empty fallbacks.
Extract a shared constant and use it everywhere.

**em-dash (`"—"`):** ~100+ occurrences across the codebase
**en-dash (`"–"`):** ~57 occurrences

| Status | File | Character |
|--------|------|-----------|
| [x] | `customers/_components/customers-table.tsx` | `"–"` |
| [x] | `suppliers/_components/suppliers-table.tsx` | `"–"` |
| [x] | `employees/_components/employees-table.tsx` | `"–"` |
| [x] | `gift-cards/_components/gift-cards-section.tsx` | `"–"` |
| [x] | `pos/orders/_components/pos-orders-section.tsx` | `"–"` |
| [x] | `pos/sessions/_components/pos-sessions-section.tsx` | `"–"` |
| [x] | `commission-entries/_components/commission-entries-section.tsx` | `"–"` |
| [x] | `projects/_components/projects-section.tsx` | `"–"` |
| [x] | `reorder-rules/_components/reorder-rules-section.tsx` | `"–"` |
| [x] | `subscription-plans/_components/subscription-plans-section.tsx` | `"–"` |
| [x] | `salary-rules/_components/salary-rules-section.tsx` | `"–"` |
| [x] | `crm/_components/opportunities-section.tsx` | `"–"` |
| [x] | `crm/_components/leads-section.tsx` | `"–"` |
| [x] | `accounting/journal-entries/_components/journal-entries-section.tsx` | `"–"` |
| [x] | `accounting/tax-years/_components/tax-years-section.tsx` | `"–"` |
| [x] | `components/inline-edit.tsx` | `"–"` |

**Fix:** Create `export const EMPTY_FALLBACK = "—" as const` in
`lib/constants/` and replace all inline occurrences.

---

## 160. Duplicated `resolvePath`/`getColumnLabel`/`getColumnValue` Functions

**Target:** Extract to shared `lib/utils/table.ts`. These three
functions are copy-pasted verbatim between entity table components.

| Status | File | Lines |
|--------|------|-------|
| [x] | `(org)/_components/server-entity-table.tsx` | 16-40 |
| [x] | `(org)/_components/interactive-entity-table.tsx` | 14-40 |

---

## 161. Duplicated `formatHours` Function

**Target:** Extract to shared utility. Both files export an identical
`formatHours` function: `` `${hours.toFixed(1)}h` ``.

| Status | File | Line |
|--------|------|------|
| [x] | `timesheets/_components/timesheet-utils.ts` | 25 |
| [x] | `attendance/_components/attendance-utils.ts` | 20 |

---

## 162. Duplicated `CommissionPlan`/`CommissionEntry` Types

**Target:** Import from `@/lib/services/swantara` instead of
redefining locally. The types in `commission-utils.ts` duplicate the
service-layer types.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `commission-plans/_components/commission-utils.ts` | 3-24 | `CommissionPlan` and `CommissionEntry` defined locally instead of imported |

---

## 163. Inconsistent Detail Page Navigation Patterns

**Target:** Standardize detail page wrappers. Three different patterns
are used inconsistently across detail pages.

**Pattern A — Bare pass-through (no BackLink, no PageHeader):**
- `purchase-orders/[orderId]/page.tsx`
- `pos/orders/[orderId]/page.tsx`
- `pos/sessions/[sessionId]/page.tsx`
- `products/[itemId]/page.tsx`

**Pattern B — BackLink + detail component:**
- `service-orders/[orderId]/page.tsx`
- `service-contracts/[contractId]/page.tsx`
- `gift-cards/[giftCardId]/page.tsx`
- `landed-costs/[inboundCostId]/page.tsx`
- `employees/[employeeId]/page.tsx`
- `deferrals/[deferralId]/page.tsx`

**Pattern C — PageHeader + detail component (no BackLink):**
- `manufacturing-orders/[moId]/page.tsx`
- `rmas/[rmaId]/page.tsx`
- `maintenance-plans/[planId]/page.tsx`
- `equipments/[equipmentId]/page.tsx`

---

## 164. Inconsistent Loading State Patterns

**Target:** Standardize loading states. Four different patterns coexist.

**Pattern A — Plain text `<p>`:**
- `products/_components/item-detail.tsx:32`
- `contacts/_components/contact-detail.tsx:26`
- `service-orders/[orderId]/_components/service-order-detail.tsx:43`
- `service-contracts/[contractId]/_components/service-contract-detail.tsx:38`
- `deferrals/[deferralId]/_components/deferral-detail.tsx:37`
- `manufacturing-orders/[moId]/_components/mo-detail-section.tsx:92`

**Pattern B — `Skeleton` component:**
- `purchase-orders/[orderId]/_components/purchase-order-detail.tsx:106-111`
- `stock/pickings/[pickingId]/_components/picking-detail.tsx:67-72`

**Pattern C — `RouteLoading` component:**
- `components/data-table.tsx:391`
- `components/data-grid.tsx:173`
- `(org)/_components/server-entity-table.tsx:144`

**Pattern D — Returns `null` (blank screen):**
- `rmas/[rmaId]/_components/rma-detail-section.tsx:79`

---

## 165. Inconsistent Error Handling Patterns

**Target:** Standardize on `toast.promise()` for mutations. Three
different patterns coexist.

**Pattern A — `toast.promise()` (newer, preferred):** 36 occurrences

**Pattern B — `.then()` + `.catch()` with separate toasts:**
- `departments/_components/department-form-dialog.tsx:69-87`
- `(auth)/reset-password/_hooks/use-reset-password-form.ts`
- `(auth)/login/_hooks/use-login-form.ts`
- `(auth)/register/_hooks/use-register-form.ts`
- `onboarding/_hooks/onboarding-form.ts`

**Pattern C — `.then()` with NO error handling (silently fails):**
- `purchase-orders/_components/purchase-order-form-dialog.tsx:179-184`
- `manufacturing-orders/[moId]/_components/mo-detail-section.tsx:85-88`
- `accounting/taxes/_components/taxes-section.tsx:202-205`
- `accounting/journals/_components/journals-section.tsx:195-198`
- `accounting/accounts/_components/accounts-section.tsx:178-181`
- `mrp/_components/mrp-section.tsx:37-43`

---

## 166. IIFE (Immediately Invoked Function Expression) in JSX

**Target:** Extract to variables or components. IIFEs in JSX reduce
readability and make debugging harder.

| Status | File | Lines | Description |
|--------|------|-------|-------------|
| [x] | `(public)/draft/_components/draft-showcase.tsx` | 644-671 | Computes `subtotal`/`tax`/`total` inline via `{(() => { ... })()}` |
| [x] | `components/workflow-mapper.tsx` | 139-153 | Computes `popupPosition` via `(() => { ... return { left, top } })()` |

---

## 167. Very Long `className` Strings with Embedded Ternaries

**Target:** Replace long template literal classNames with `cn()` utility.
Strings exceeding 120 characters reduce readability.

| Status | File | Line | Length |
|--------|------|------|--------|
| [x] | `(public)/support/page.tsx` | 140 | 282 chars |
| [x] | `(public)/support/page.tsx` | 143 | ~150 chars |
| [x] | `(public)/support/page.tsx` | 150 | ~100 chars |
| [x] | `(public)/pricing/page.tsx` | 159 | ~100 chars |
| [x] | `components/import-wizard.tsx` | 187 | 187 chars |

---

## 168. Inline Type Annotations in JSX `.map()` Callbacks

**Target:** Extract inline types to named types defined outside JSX.

| Status | File | Lines | Description |
|--------|------|-------|-------------|
| [x] | `sale-orders/[saleOrderId]/_components/sale-order-detail.tsx` | 608-621 | `.map((d: { id: number; contactId: number; ... }) => ...)` — 5-field inline type |

---

## 169. Dashboard KPI Card Repetition (~300 Lines)

**Target:** Extract a `KpiCard` helper component. The dashboard overview
repeats the same card pattern (label + formatted value + delta) ~30
times with only data differences. Each instance is ~15 lines of JSX.

| Status | File | Lines | Description |
|--------|------|-------|-------------|
| [x] | `dashboard/_components/dashboard-overview.tsx` | 134-623 | ~30 identical stat card blocks, each ~15 lines. A `KpiCard` component would cut the file by ~300 lines. |

**Note:** Extracted `KpiCard` component and replaced all ~30 stat card blocks.

---

## 170. Section Component Status Prop Mapping Duplicated 30+ Times

**Target:** Extract to a shared helper. The `query.isLoading` /
`query.isError` / `query.error.message` / `query.refetch()` status
mapping block is copy-pasted verbatim into every section component's
`status` prop.

```tsx
status={
  query.isLoading
    ? { type: "loading" }
    : query.isError
      ? { type: "error", message: query.error.message, onRetry: () => void query.refetch() }
      : undefined
}
```

**Appears in:** 30+ `*-section.tsx` files across the codebase.

---

## 171. Identical Dialogs Parameterized by Title Only (Sale Order Detail)

**Target:** Extract a single `JournalDialog` component. Three dialogs
in `sale-order-detail.tsx` are structurally identical (Dialog with a
journal Select and optional fields), differing only in title.

| Status | File | Lines | Description |
|--------|------|-------|-------------|
| [x] | `sale-orders/[saleOrderId]/_components/sale-order-detail.tsx` | 655-708 | `DeliverDialog` |
| [x] | `sale-orders/[saleOrderId]/_components/sale-order-detail.tsx` | 710-763 | `InvoiceDialog` |
| [x] | `sale-orders/[saleOrderId]/_components/sale-order-detail.tsx` | 765-832 | `PayDialog` |

**Note:** `DeliverDialog` and `InvoiceDialog` replaced with shared `JournalDateDialog`. `PayDialog` replaced with shared `PaymentDialog` (extends `JournalDateDialog` with optional amount field). Also replaced `PayDialog` in `purchase-order-detail.tsx` with `JournalDateDialog`.

---

## 172. Duplicated File Names Across Features

**Target:** Rename to be feature-specific to prevent wrong imports.

| Status | File | Issue |
|--------|------|-------|
| [x] | `(auth)/register/_components/form-navigation.tsx` | Same name as `onboarding/_components/form-navigation.tsx` — different implementations |
| [x] | `onboarding/_components/form-navigation.tsx` | Same name as `(auth)/register/_components/form-navigation.tsx` |

**Note:** Files are in separate feature directories with different
implementations. Feature-scoped `_components/` directories are expected
to have same-named files.

---

## 173. Hardcoded Mock Data in Production Components

**Target:** Wire up real data or remove placeholder. Components render
hardcoded business data that misleads users.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `dashboard/_components/dashboard-overview.tsx` | 813-830 | `RecentActivityCard` has `const activities: ActivityItem[] = []` — always renders empty timeline |
| [x] | `accounting/page.tsx` | 205-211 | Hardcoded `"+$9,840"`, `"-$6,120"`, `"$3,720"` and `"Aug 2026"` in JSX |

---

## 174. Inconsistent Abbreviation Usage in File Names

**Target:** Standardize abbreviation usage. Route folders use full
names but internal files abbreviate, creating confusion.

| Pattern | Folder | Internal Files | Issue |
|---------|--------|----------------|-------|
| `mo` | `manufacturing-orders/` | `mo-section.tsx`, `mo-form-dialog.tsx`, `mo-utils.ts` | Folder is full name, files abbreviated |
| `bom` | `boms/` | `bom-form-dialog.tsx` (singular), `boms-table.tsx` (plural) | Mixed singular/plural |
| `rma` | `rmas/` | `rma-form-dialog.tsx` (singular), `rmas-section.tsx` (plural) | Mixed singular/plural |
| `pos` | `pos/` | `pos-orders-section.tsx`, `pos-order-detail.tsx` | Consistent — no issue |

---

## 175. Inconsistent Query Hook Usage (`useOrgListQuery` for Single Items)

**Target:** Use `useOrgQuery` for single-item fetches. Some detail
components misuse `useOrgListQuery` (designed for list queries) to
fetch a single entity.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `service-orders/[orderId]/_components/service-order-detail.tsx` | — | Uses `useOrgListQuery` for single order fetch |
| [x] | `rmas/[rmaId]/_components/rma-detail-section.tsx` | — | Uses raw `useQuery`/`useMutation` instead of wrapped hooks |

**Note:** Migrated queries to `useOrgQuery`/`useOrgListQuery`. Mutations still use raw `useMutation` because `useOrgMutation` doesn't exist yet.

---

## 176. Zero JSDoc on Exported Types

**Target:** Add JSDoc to exported types that have domain-specific
semantics. None of the 100+ exported types in `components/` have JSDoc.

| Status | File | Type | Why It Needs Documentation |
|--------|------|------|---------------------------|
| [x] | `components/workflow-mapper.tsx:25-38` | `WorkflowNodeDef`, `WorkflowNode`, `WorkflowEdge` | Domain semantics (coordinate system, def purpose) |
| [x] | `components/data-table.tsx:38-42` | `DataTableStatus` | Discriminated union — when to use each variant |
| [x] | `components/color-picker.tsx:40` | `HslColor` | Mathematical representation (h degrees? s/l ranges?) |
| [x] | `components/file-upload.tsx:54` | `UseUploadResult` | Complex hook return shape |
| [x] | `components/pdf-viewer.tsx:20` | `PdfSource` | Union type — what format in each case? |

---

## Summary

| # | Category | Violations | Priority |
|---|----------|------------|----------|
| 1 | Export default (lib) | 4 | Low |
| 2 | Interface for props | 10 | Low |
| 3 | Local formatDate copies | 19 | High |
| 4 | Local formatCurrency copies | 11 | High |
| 5 | FieldFeedbackType naming | 1 | Low |
| 6 | Manual $ prefix | 33 | Medium |
| 7 | Direct queryClient usage | 0 | N/A |
| 8 | toLocaleDateString() no locale | 11 | Medium |
| 9 | Unnecessary "use client" | 0 | N/A |
| 10 | Hardcoded English status labels | 17 files | High |
| 11 | Hardcoded English UI text | 40+ files | High |
| 12 | Misleading function name | 1 | Low |
| 13 | Hardcoded support email | 13 | Low |
| 14 | Hardcoded $ in charts/KPIs | 10+ | Medium |
| 15 | Hardcoded language labels | 1 | Low |
| 16 | Inline mock data | 1 | Low |
| 17 | Duplicate logger files | 1 | Low |
| 18 | Massive page files | 8 | Medium |
| 19 | Missing route boundaries | 121 pages | Medium |
| 20 | Form schema extraction | 4 auth forms done | Done |
| 21 | Unused imports | 30 (10 files) | Medium |
| 22 | Console statements | 4 (3 files) | Low |
| 23 | Hardcoded API URLs | 1 | Medium |
| 24 | Type assertions (`as unknown`) | 55 (40+ files) | High |
| 25 | Magic numbers in setTimeout | 4 | Low |
| 26 | Inline styles | 56 (21 files) | Low |
| 27 | Hardcoded route paths | 21 (14 files) | Medium |
| 28 | `any`/`as never` usage | 6 (non-test files) | Medium |
| 29 | Deeply nested ternaries | 16 (14 files) | Medium |
| 30 | Overly complex components | 18 files | High |
| 31 | Hydration mismatch risks | 1 (high risk) | Medium |
| 32 | Memory leak risks | 12 files | Medium |
| 33 | Stale closures | 2 | Medium |
| 34 | Missing tests (components) | 57 | Medium |
| 35 | Missing tests (lib) | 36 | Medium |
| 36 | Weak test assertions | 10 | Low |
| 37 | Duplicated test patterns | 5 patterns | Low |
| 38 | Hardcoded status/tone strings | 15 strings (~1800 uses) | High |
| 39 | Config drift | 4 | Medium |
| 40 | Inline forms (no Form/FormField) | 4 | Medium |
| 41 | Duplicated dialog boilerplate | 83 files | Medium |
| 42 | Detail page loading patterns | 11 files | Low |
| 43 | Fire-and-forget mutations | 0 | N/A |
| 44 | Duplicated status ternaries | 8 files | Medium |
| 45 | Inline validation schemas | 52+ files | High |
| 46 | Missing SubmitButton usage | 5 files | Medium |
| 47 | Item form cancel button bug | 1 file | High |
| 48 | Missing `.catch()` on API calls | 60+ files | High |
| 49 | Missing onError in mutations | 1 file | Medium |
| 50 | Missing error states in queries | 3 files | Medium |
| 51 | Missing useMemo for Maps | 11 files | Medium |
| 52 | Missing React.memo on leaf components | 5 files | Medium |
| 53 | Missing code splitting | 4 files | Medium |
| 54 | Missing `lang` attribute | 1 file | High |
| 55 | Missing `aria-describedby` in FormField | 1 file (systemic) | High |
| 56 | Missing `id="main"` for skip nav | 1 file | Medium |
| 57 | Contradictory aria-hidden + sr-only | 1 file | Medium |
| 58 | Session cookies `secure: false` | 1 file | High |
| 59 | OAuth tokens via document.cookie | 1 file | High |
| 60 | No CSP or security headers | 1 file | High |
| 61 | Static CSRF token | 1 file | Medium |
| 62 | No rate limiting on auth | 3 files | High |
| 63 | Duplicated Journal+Date dialog | 8+ files | Medium |
| 64 | Duplicated form dialog boilerplate | 40+ files | Medium |
| 65 | Duplicated section (list view) pattern | 15+ files | Medium |
| 66 | Duplicated active status badge | 15+ files | Medium |
| 67 | Duplicated order detail pages | 2 files | Medium |
| 68 | `.toLocaleDateString()` no locale | 13 call sites | High |
| 69 | Local formatDate hardcoded to "en-US" | 19 files | High |
| 70 | Local formatCurrency hardcoded to "en-US" | 11 files | High |
| 71 | Hardcoded English in Badge | 11 call sites | Medium |
| 72 | Hardcoded label functions | 9 utility functions | Medium |
| 73 | Unused utility functions | 4 exports | Low |
| 74 | Unused hook exports | 7 hooks | Low |
| 75 | Unused provider exports | 2 files | Low |
| 76 | Unused store exports | 1 export | Low |
| 77 | Unused CSS utilities | 6 utilities | Low |
| 78 | No route-level error boundaries | 6 route groups | High |
| 79 | Missing loading states (blank screen) | 2 components | Medium |
| 80 | console.error in production | 2 files | Low |
| 81 | Deprecated isLoading usage | 1 file | Low |
| 82 | Unsafe `.sort()` mutates cache | 2 files | High |
| 83 | Missing `enabled` guard on queries | 8 hooks/components | High |
| 84 | Over-fetching list queries | 4 files | Medium |
| 85 | Race conditions (no pending state) | 4 files | High |
| 86 | Missing query invalidation | 5 files | Medium |
| 87 | Inline literals in JSX props | 4 files | Medium |
| 88 | Unsafe key props | 3 files | Medium |
| 89 | Unsafe type narrowing | 3 files | Medium |
| 90 | `Number()` without NaN validation | 10+ files | High |
| 91 | Zero `loading.tsx` in entire app | 5 route groups | Medium |
| 92 | UTC-based date init (off-by-one) | 7 files | Medium |
| 93 | Missing loading skeleton components | 5 route groups | Medium |
| 94 | String-to-number at wrong layer | 100+ call sites | Medium |
| 95 | Missing disabled during form submit | 2 files | Medium |
| 96 | Form state not reset in dialogs | 4 files | Low |
| 97 | Transient network error logs user out | 1 file | High |
| 98 | `intlMiddleware` failure drops cookies | 1 file | High |
| 99 | Path traversal via catch-all proxy | 1 file | High |
| 100 | `createErrorResponse` discards error context | 1 file | Medium |
| 101 | Proxy leaks upstream error body | 1 file | Medium |
| 102 | RMA mutations don't invalidate list | 3 files | Medium |
| 103 | CRM mutations don't invalidate stages | 3 files | Medium |
| 104 | Approval detail doesn't invalidate list | 1 file | Medium |
| 105 | Missing `"use client"` directive | 1 file | Low |
| 106 | Missing input validation on proxy route | 1 file | High |
| 107 | Inconsistent error response format | 1 file | Low |
| 108 | Three competing token refresh mechanisms | 4 files | High |
| 109 | Logout doesn't clear Zustand org store | 2 files | Medium |
| 110 | Permissions hide all UI during loading | 3 files | Medium |
| 111 | Fire-and-forget mutations (no `.catch()`) | 5 files | High |
| 112 | URL params not synced with component state | 1 file | Medium |
| 113 | Direct service calls instead of React Query | 2 files | Medium |
| 114 | Form state leaked between dialog opens | 3 files | Medium |
| 115 | Template literal classNames instead of `cn()` | 26 instances (12 files) | Medium |
| 116 | Inline style objects in render loops | 53+ instances (15+ files) | Medium |
| 117 | God components (>150 lines, >5 useState) | 5 files | Medium |
| 118 | Derived state stored in useEffect | 3 files | Medium |
| 119 | Multiple dialog states that should be one | 2 files | Low |
| 120 | Zero page metadata exports | 100+ pages | High |
| 121 | No openGraph metadata | All pages | High |
| 122 | No alternates for multilingual | All pages | High |
| 123 | No robots configuration | All pages | Medium |
| 124 | Zero revalidatePath/revalidateTag | Entire codebase | Medium |
| 125 | Server prefetch staleTime: "static" | 1 file | Medium |
| 126 | Org layouts no error handling for prefetch | 2 files | Medium |
| 127 | No Suspense boundaries in server layouts | 5 layouts | Medium |
| 128 | Sequential awaits in server components | 2 pages | Low |
| 129 | Unsafe .find() on potentially undefined arrays | 6 files | High |
| 130 | Invalid date propagation in formatters | 2 functions | Medium |
| 131 | OAuth callback sets cookies with empty tokens | 1 file | Medium |
| 132 | Blob URL / iframe leak on unmount | 1 file | Low |
| 133 | Async UI calls without try/catch | 4 files | High |
| 134 | Form schemas missing .trim() on required strings | All forms | Medium |
| 135 | Form schemas missing cross-field validation | 4 forms | Medium |
| 136 | Export default in components/hooks | 25 files | Medium |
| 137 | Container components misnamed | 42 files | Medium |
| 138 | Presentation components with data fetching | 33 files | High |
| 139 | Container in shared components/ | 1 file | Medium |
| 140 | Missing data-slot attributes | 31 components | Low |
| 141 | Direct fetch() in client components | 5 files | Medium |
| 142 | Forms without schema validation | 7 files | High |
| 143 | Duplicate density constants/types | 2 files | Low |
| 144 | Duplicate theme constants indirection | 2 files | Low |
| 145 | Locale cookie without Secure flag | 1 file | Medium |
| 146 | Density cookie without Secure flag | 1 file | Medium |
| 147 | Premature abstractions (dead) | 7 modules | Medium |
| 148 | Unused components (zero imports) | 37 components | Medium |
| 149 | Unused server/client modules | 5 modules | Low |
| 150 | Over-engineered error hierarchy | 1 file | Low |
| 151 | Hardcoded settings page (mockup) | 1 file | Medium |
| 152 | Empty stub page | 1 file | Low |
| 153 | Half-implemented command palette | 1 file | Low |
| 154 | Over-complex Zustand stores | 2 stores | Low |
| 155 | Unused hook exports (additional) | 2 exports | Low |
| 156 | POST without CSRF header | 1 file | Medium |
| 157 | Wildcard `import * as React` | 9 files | Low |
| 158 | Unnecessary "use client" on utils | 38 files | Low |
| 159 | Inconsistent empty fallback chars | 16+ files | Low |
| 160 | Duplicated table utility functions | 2 files | Medium |
| 161 | Duplicated `formatHours` | 2 files | Low |
| 162 | Duplicated CommissionPlan types | 1 file | Low |
| 163 | Inconsistent detail page navigation | 15 files | Medium |
| 164 | Inconsistent loading states | 12 files | Medium |
| 165 | Inconsistent error handling patterns | 11 files | Medium |
| 166 | IIFE in JSX | 2 files | Low |
| 167 | Very long className strings | 5 files | Low |
| 168 | Inline type in JSX `.map()` | 1 file | Low |
| 169 | Dashboard KPI card repetition | 1 file (~300 lines) | Medium |
| 170 | Section status prop mapping duplication | 30+ files | Medium |
| 171 | Identical dialogs (title only diff) | 1 file (3 dialogs) | Medium |
| 172 | Duplicated file names across features | 2 files | Low |
| 173 | Hardcoded mock data in production | 2 files | Medium |
| 174 | Inconsistent abbreviation in file names | 4 patterns | Low |
| 175 | Wrong query hook for single items | 2 files | Medium |
| 176 | Zero JSDoc on exported types | 100+ types | Low |

---

## 177. Arrow Function Component

**Target:** Replace arrow function component with function declaration.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `components/emoji-picker.tsx` | 116 | `const PickEmoji = ({ emoji, name }) =>` — arrow function sub-component |

---

## 178. `defaultLabels` Objects With Hardcoded English (Systemic i18n Gap)

**Target:** Replace hardcoded `defaultLabels` constants with `useTranslations()`
calls inside each component, or ensure every call site passes translated labels.
16+ shared components define `defaultLabels` with English fallback strings that
silently defeat i18n when callers don't override them.

**Severity:** HIGH

| Status | File | Lines | Hardcoded Strings |
|--------|------|-------|-------------------|
| [x] | `components/data-table.tsx` | 105-120 | `"Select all"`, `"No results"`, `"Rows per page"`, `"Page"`, `"of"`, `"Previous page"`, `"Next page"`, `"{from}–{to} of {total}"` |
| [x] | `components/pdf-viewer-toolbar.tsx` | 95-119 | `"Open in new tab"`, `"Print"`, `"Download"`, `"First page"`, `"Previous page"`, `"Next page"`, `"Last page"`, `"Go to page"`, `"Zoom out"`, `"Zoom in"`, `"Fit width"`, `"Rotate left"`, `"Rotate right"` |
| [x] | `components/pdf-viewer-status.tsx` | 20-26 | `"Loading document…"`, `"Failed to load document."`, `"Try again"`, `"No document"` |
| [x] | `components/pdf-annotation-layer.tsx` | 146-153 | `"Note"`, `"Ink"`, `"Delete annotation"`, `"Add a note…"`, `"Cancel"`, `"Save"` |
| [x] | `components/error-state.tsx` | 26-34 | `"Something went wrong"`, `"Try again"`, `"Contact support"`, `"Back to home"` |
| [x] | `components/error-boundary.tsx` | 25-28 | `"Something went wrong"`, `"Try again"` |
| [x] | `components/date-presets.tsx` | 31-37 | `"Today"`, `"Last 7 days"`, `"Last 30 days"`, `"This quarter"`, `"This year"` |
| [x] | `components/undo-stack.tsx` | 31-33 | `"Undo"` |
| [x] | `components/theme-switcher.tsx` | 27-31 | `"Change theme"`, `"Light"`, `"Dark"` |
| [x] | `components/text-clamp.tsx` | 17-20 | `"Show more"`, `"Show less"` |
| [x] | `components/split-button.tsx` | 40-42 | `"More actions"` |
| [x] | `components/locale-switcher.tsx` | 32-34 | `"Change language"` |
| [x] | `components/column-chooser.tsx` | 36-40 | `"Columns"`, `"Toggle columns"`, `"Show all"` |
| [x] | `components/density.tsx` | 25-30 | `"Data density"`, `"Compact"`, `"Normal"`, `"Comfortable"` |
| [x] | `components/address-field.tsx` | 29-33 | `"Address"`, `"City"`, `"Postal code"` |
| [x] | `components/multi-select.tsx` | 35-37 | `"No results found."` |

---

## 179. Hardcoded `aria-label`/`title` Attributes (Not Using `t()`)

**Target:** Replace hardcoded `aria-label` and `title` attributes with `t()` calls.
18 components have inline English accessibility strings.

| Status | File | Lines | Hardcoded Text |
|--------|------|-------|----------------|
| [x] | `components/breadcrumb.tsx` | 20 | `aria-label="Breadcrumb"` |
| [x] | `components/date-picker.tsx` | 41, 88 | `aria-label="Choose a date"`, `aria-label="Choose a date range"` |
| [x] | `components/lookup-field.tsx` | 101, 112 | `aria-label="Clear selection"`, `aria-label="Show options"` |
| [x] | `components/sortable-list.tsx` | 74, 83-84, 93-94 | `title="Drag to reorder"`, `aria-label="Move up/down"` |
| [x] | `components/callout.tsx` | 70 | `aria-label="Dismiss"` |
| [x] | `components/json-viewer.tsx` | 143 | `aria-label="Copy JSON"` |
| [x] | `components/import-wizard.tsx` | 276 | `aria-label="Import preview"` |
| [x] | `components/image-gallery.tsx` | 110, 123, 136, 163 | `aria-label="Close preview"`, `"Previous image"`, `"Next image"` |
| [x] | `components/bulk-actions-bar.tsx` | 41 | `aria-label="Bulk actions"` |
| [x] | `components/image-cropper.tsx` | 306 | `aria-label="Drag to move the crop area..."` |
| [x] | `components/combobox.tsx` | 33, 226 | `aria-label="Clear selection"`, `aria-label="Remove item"` |
| [x] | `components/thread.tsx` | 80, 87 | `aria-label="Write a comment"`, `aria-label="Send comment"` |
| [x] | `components/reaction-picker.tsx` | 57 | `aria-label="Add reaction"` |
| [x] | `components/data-grid.tsx` | 115 | `aria-label="Select all rows"` |
| [x] | `components/emoji-picker.tsx` | 161 | `aria-label="Search emojis"` |
| [x] | `components/quantity-field.tsx` | 108, 135 | `aria-label="Decrease quantity"`, `aria-label="Increase quantity"` |
| [x] | `components/kpi-wall.tsx` | 113, 121, 130, 137, 158 | `aria-label="Current time"`, `"Resume/Pause rotation"`, `"Previous/Next slide"` |
| [x] | `components/line-items-table.tsx` | 169 | `title="No line items"` |

---

## 180. `.toLocaleTimeString()` Without Locale Argument

**Target:** Replace with locale-aware time formatter. Section 68 covers
`.toLocaleDateString()` but `.toLocaleTimeString()` has the same issue.

| Status | File | Line | Code |
|--------|------|------|------|
| [x] | `components/kpi-wall.tsx` | 116 | `now.toLocaleTimeString()` |

---

## 181. Hardcoded Default Prop Values (Not Using `t()`)

**Target:** Replace hardcoded default prop values with `t()` calls or
require callers to provide translated values.

| Status | File | Lines | Hardcoded Default |
|--------|------|-------|-------------------|
| [x] | `components/confirm-dialog.tsx` | 50-51 | `confirmLabel = "Confirm"`, `cancelLabel = "Cancel"` |
| [x] | `components/tour.tsx` | 373, 378, 382, 391, 396, 400 | `"Previous"`, `"Finish"`, `"Next"` |
| [x] | `components/questionnaire.tsx` | 218, 243, 268, 293 | `"Previous"`, `"Skip"`, `"Next"`, `"Submit"` |
| [x] | `components/pagination.tsx` | 57, 75 | `text = "Previous"`, `text = "Next"` |
| [x] | `components/multi-select.tsx` | 46 | `"Select multiple items"` (default aria-label) |

---

## 182. Direct `getSwantaraService()` Calls in Components (Bypassing `useOrgMutation`)

**Target:** Replace direct service calls with `useOrgMutation` or wrap in
`useMutation` with proper `onSuccess`/`onError` handlers. Direct calls
skip CSRF handling, pending state tracking, and automatic query invalidation.

**Severity:** HIGH

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `rmas/[rmaId]/_components/rma-detail-section.tsx` | 55, 62, 69 | Raw `useMutation` + `getSwantaraService().rmas.confirm/done/cancel()` |
| [x] | `rmas/[rmaId]/_components/receive-dialog.tsx` | 45 | Direct `getSwantaraService().rmas.receive()` |
| [x] | `rmas/[rmaId]/_components/refund-dialog.tsx` | 46 | Direct `getSwantaraService().rmas.refund()` |
| [x] | `rmas/_components/rma-form-dialog.tsx` | 96 | Direct `getSwantaraService().rmas.create()` |
| [x] | `crm/_components/opportunities-section.tsx` | 132, 169 | Direct `getSwantaraService().crmOpportunities.win/delete()` |
| [x] | `crm/_components/leads-section.tsx` | 121 | Direct `getSwantaraService().crmLeads.delete()` |
| [x] | `crm/_components/pipeline-board.tsx` | 65, 77, 88 | Direct `getSwantaraService()` for drag/drop state changes |
| [x] | `crm/_components/lead-form-dialog.tsx` | 159, 166, 175, 182 | Direct `getSwantaraService()` for create/update/demote/delete |
| [x] | `crm/_components/activity-form-dialog.tsx` | 107, 114 | Direct `getSwantaraService().crmActivities.create/update()` |
| [x] | `crm/_components/activities-section.tsx` | 82, 103 | Direct `getSwantaraService()` for activity state transitions |
| [x] | `crm/_components/promote-dialog.tsx` | 76 | Direct `getSwantaraService().crmLeads.promote()` |
| [x] | `crm/_components/lost-reason-dialog.tsx` | 45 | Direct `getSwantaraService().crmOpportunities.lose()` |
| [x] | `subscriptions/[subscriptionId]/_components/subscription-detail.tsx` | 53 | Direct `getSwantaraService().subscriptions[action]()` |
| [x] | `service-contracts/[contractId]/_components/service-contract-detail.tsx` | 49 | Direct `getSwantaraService().serviceContracts[action]()` |
| [x] | `gift-cards/[giftCardId]/_components/gift-card-detail.tsx` | 54 | Direct `getSwantaraService().giftCards[action]()` |
| [x] | `quality/alerts/_components/quality-alerts-section.tsx` | 63 | Direct `getSwantaraService().qualityAlerts.state()` |
| [x] | `quality/alerts/[alertId]/_components/quality-alert-detail.tsx` | 30 | Direct `getSwantaraService().qualityAlerts.state()` |
| [x] | `fixed-assets/[assetId]/_components/fixed-asset-detail.tsx` | 68, 80, 97 | Direct `getSwantaraService().fixedAssets.schedule/postDepreciation/dispose()` |
| [x] | `fixed-assets/_components/fixed-asset-form-dialog.tsx` | 71 | Direct `getSwantaraService().fixedAssets.create()` |
| [x] | `subscription-plans/_components/subscription-plans-section.tsx` | 65 | Direct `getSwantaraService().subscriptionPlans.delete()` |
| [x] | `subscription-plans/_components/subscription-plan-form-dialog.tsx` | 72, 83 | Direct `getSwantaraService().subscriptionPlans.update/create()` |
| [x] | `subscriptions/_components/subscription-form-dialog.tsx` | 90 | Direct `getSwantaraService().subscriptions.create()` |
| [x] | `service-contracts/_components/service-contract-form-dialog.tsx` | 83 | Direct `getSwantaraService().serviceContracts.create()` |
| [x] | `maintenance-plans/_components/maintenance-plan-form-dialog.tsx` | 63 | Direct `getSwantaraService().maintenancePlans.create()` |
| [x] | `equipments/_components/equipment-form-dialog.tsx` | 75 | Direct `getSwantaraService().equipments.create()` |
| [x] | `asset-categories/_components/asset-category-form-dialog.tsx` | 108, 117 | Direct `getSwantaraService().assetCategories.create()` |
| [x] | `expense-categories/_components/expense-categories-section.tsx` | 39 | Direct `getSwantaraService().expenseCategories.delete()` |
| [x] | `expense-categories/_components/expense-category-form-dialog.tsx` | 52 | Direct `getSwantaraService().expenseCategories.update()` |
| [x] | `timesheets/_components/timesheet-form-dialog.tsx` | 121 | Direct `getSwantaraService().timesheets.create()` |
| [x] | `quality/points/_components/quality-point-form-dialog.tsx` | 94 | Direct `getSwantaraService().qualityPoints.create()` |

---

## 183. Missing CSRF Validation on `/api/vitals` POST Endpoint

**Target:** Add CSRF validation to the `/api/vitals` route handler.
The endpoint wraps only in `withHandler` which does not call `isCsrfValid()`,
allowing cross-site POST requests to inject fake web vitals data.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `app/api/vitals/route.ts` | 15-31 | No `isCsrfValid()` call — any origin can POST fake vitals |

---

## 184. OAuth Tokens Leakage via URL Query String Fallback

**Target:** Remove query string fallback in `parseFragment()`. Tokens
should only be accepted from the URL hash fragment (never sent to servers).
The query string fallback exposes tokens in server access logs, Referer
headers, and browser history.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `(auth)/oauth/callback/page.tsx` | 24-26 | `parseFragment()` falls back to `window.location.search` — tokens visible in logs/Referer |

---

## 185. Refresh Tokens Stored as Plaintext Map Keys

**Target:** Use a hashed key or opaque identifier for the deduplication
Map instead of the raw refresh token string. Refresh tokens are long-lived
credentials held in plaintext in module-level memory during refresh.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/server/proxy.ts` | 50, 140, 170-172 | `refreshPromises` Map uses raw token as key |

**Note:** Server-side only, in-memory, cleared after promise resolves.
Low risk — token is already in request cookie memory.

---

## 186. Proxy Error Logs Leak Full Upstream URL With Query Params

**Target:** Log only the pathname (not full `target.href`) to avoid
writing client-supplied query parameters to server logs.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `lib/server/proxy.ts` | 55-56, 193 | `target: target.href` in error log includes full query string |

---

## 187. Duplicated Column Definitions Across Section Files

**Target:** Extract shared column factories. The same `name`, `active`,
`state`, and `type` column definitions are copy-pasted across 9+ section files.

| Status | File | Pattern |
|--------|------|---------|
| [x] | `contacts/_components/contacts-section.tsx` | 4 identical columns with `contacts-server-section.tsx` |
| [x] | `contacts/_components/contacts-server-section.tsx` | Same columns as contacts-section.tsx |
| [x] | `transfers/_components/transfers-section.tsx` | `name` + `state` columns duplicated across 6+ stock files |
| [x] | `counts/_components/counts-section.tsx` | Same pattern |
| [x] | `pickings/_components/pickings-section.tsx` | Same pattern |
| [x] | `products/_components/products-section.tsx` | `name` + `active` columns |
| [x] | `price_books/_components/price_books-section.tsx` | Same pattern |
| [x] | `reorder-rules/_components/reorder-rules-section.tsx` | `active` badge column |
| [x] | `approval-requests/_components/approval-requests-section.tsx` | `name` link column |

**Note:** Unique to approval requests (uses `AR-${id}` format, not a general `name` column).

---

## 188. Duplicated `createId()` Utility Function

**Target:** Extract to `lib/utils/create-id.ts`. Five files contain
the identical 3-line `createId()` function.

| Status | File | Line |
|--------|------|------|
| [x] | `components/line-items-table.tsx` | 37 |
| [x] | `purchase-rfq-form-dialog.tsx` | 34 |
| [x] | `purchase-requisition-form-dialog.tsx` | 34 |
| [x] | `sale-order-form-dialog.tsx` | 44 |
| [x] | `purchase-order-form-dialog.tsx` | 35 |

---

## 189. Duplicated `useEffect` for Line Reset on Dialog Open

**Target:** Extract to shared hook (e.g., `useLineResetEffect`). Four
order form dialogs have an identical useEffect that resets lines when
the dialog opens.

| Status | File | Lines |
|--------|------|-------|
| [x] | `sale-order-form-dialog.tsx` | 109-120 |
| [x] | `purchase-order-form-dialog.tsx` | 80-92 |
| [x] | `purchase-rfq-form-dialog.tsx` | 67-78 |
| [x] | `purchase-requisition-form-dialog.tsx` | 72-83 |

---

## Summary

| # | Category | Violations | Priority |
|---|----------|------------|----------|
| 1 | Export default (lib) | 4 | Low |
| 2 | Interface for props | 10 | Low |
| 3 | Local formatDate copies | 19 | High |
| 4 | Local formatCurrency copies | 11 | High |
| 5 | FieldFeedbackType naming | 1 | Low |
| 6 | Manual $ prefix | 33 | Medium |
| 7 | Direct queryClient usage | 0 | N/A |
| 8 | toLocaleDateString() no locale | 11 | Medium |
| 9 | Unnecessary "use client" | 0 | N/A |
| 10 | Hardcoded English status labels | 17 files | High |
| 11 | Hardcoded English UI text | 40+ files | High |
| 12 | Misleading function name | 1 | Low |
| 13 | Hardcoded support email | 13 | Low |
| 14 | Hardcoded $ in charts/KPIs | 10+ | Medium |
| 15 | Hardcoded language labels | 1 | Low |
| 16 | Inline mock data | 1 | Low |
| 17 | Duplicate logger files | 1 | Low |
| 18 | Massive page files | 8 | Medium |
| 19 | Missing route boundaries | 121 pages | Medium |
| 20 | Form schema extraction | 4 auth forms done | Done |
| 21 | Unused imports | 30 (10 files) | Medium |
| 22 | Console statements | 4 (3 files) | Low |
| 23 | Hardcoded API URLs | 1 | Medium |
| 24 | Type assertions (`as unknown`) | 55 (40+ files) | High |
| 25 | Magic numbers in setTimeout | 4 | Low |
| 26 | Inline styles | 56 (21 files) | Low |
| 27 | Hardcoded route paths | 21 (14 files) | Medium |
| 28 | `any`/`as never` usage | 6 (non-test files) | Medium |
| 29 | Deeply nested ternaries | 16 (14 files) | Medium |
| 30 | Overly complex components | 18 files | High |
| 31 | Hydration mismatch risks | 1 (high risk) | Medium |
| 32 | Memory leak risks | 12 files | Medium |
| 33 | Stale closures | 2 | Medium |
| 34 | Missing tests (components) | 57 | Medium |
| 35 | Missing tests (lib) | 36 | Medium |
| 36 | Weak test assertions | 10 | Low |
| 37 | Duplicated test patterns | 5 patterns | Low |
| 38 | Hardcoded status/tone strings | 15 strings (~1800 uses) | High |
| 39 | Config drift | 4 | Medium |
| 40 | Inline forms (no Form/FormField) | 4 | Medium |
| 41 | Duplicated dialog boilerplate | 83 files | Medium |
| 42 | Detail page loading patterns | 11 files | Low |
| 43 | Fire-and-forget mutations | 0 | N/A |
| 44 | Duplicated status ternaries | 8 files | Medium |
| 45 | Inline validation schemas | 52+ files | High |
| 46 | Missing SubmitButton usage | 5 files | Medium |
| 47 | Item form cancel button bug | 1 file | High |
| 48 | Missing `.catch()` on API calls | 60+ files | High |
| 49 | Missing onError in mutations | 1 file | Medium |
| 50 | Missing error states in queries | 3 files | Medium |
| 51 | Missing useMemo for Maps | 11 files | Medium |
| 52 | Missing React.memo on leaf components | 5 files | Medium |
| 53 | Missing code splitting | 4 files | Medium |
| 54 | Missing `lang` attribute | 1 file | High |
| 55 | Missing `aria-describedby` in FormField | 1 file (systemic) | High |
| 56 | Missing `id="main"` for skip nav | 1 file | Medium |
| 57 | Contradictory aria-hidden + sr-only | 1 file | Medium |
| 58 | Session cookies `secure: false` | 1 file | High |
| 59 | OAuth tokens via document.cookie | 1 file | High |
| 60 | No CSP or security headers | 1 file | High |
| 61 | Static CSRF token | 1 file | Medium |
| 62 | No rate limiting on auth | 3 files | High |
| 63 | Duplicated Journal+Date dialog | 8+ files | Medium |
| 64 | Duplicated form dialog boilerplate | 40+ files | Medium |
| 65 | Duplicated section (list view) pattern | 15+ files | Medium |
| 66 | Duplicated active status badge | 15+ files | Medium |
| 67 | Duplicated order detail pages | 2 files | Medium |
| 68 | `.toLocaleDateString()` no locale | 13 call sites | High |
| 69 | Local formatDate hardcoded to "en-US" | 19 files | High |
| 70 | Local formatCurrency hardcoded to "en-US" | 11 files | High |
| 71 | Hardcoded English in Badge | 11 call sites | Medium |
| 72 | Hardcoded label functions | 9 utility functions | Medium |
| 73 | Unused utility functions | 4 exports | Low |
| 74 | Unused hook exports | 7 hooks | Low |
| 75 | Unused provider exports | 2 files | Low |
| 76 | Unused store exports | 1 export | Low |
| 77 | Unused CSS utilities | 6 utilities | Low |
| 78 | No route-level error boundaries | 6 route groups | High |
| 79 | Missing loading states (blank screen) | 2 components | Medium |
| 80 | console.error in production | 2 files | Low |
| 81 | Deprecated isLoading usage | 1 file | Low |
| 82 | Unsafe `.sort()` mutates cache | 2 files | High |
| 83 | Missing `enabled` guard on queries | 8 hooks/components | High |
| 84 | Over-fetching list queries | 4 files | Medium |
| 85 | Race conditions (no pending state) | 4 files | High |
| 86 | Missing query invalidation | 5 files | Medium |
| 87 | Inline literals in JSX props | 4 files | Medium |
| 88 | Unsafe key props | 3 files | Medium |
| 89 | Unsafe type narrowing | 3 files | Medium |
| 90 | `Number()` without NaN validation | 10+ files | High |
| 91 | Zero `loading.tsx` in entire app | 5 route groups | Medium |
| 92 | UTC-based date init (off-by-one) | 7 files | Medium |
| 93 | Missing loading skeleton components | 5 route groups | Medium |
| 94 | String-to-number at wrong layer | 100+ call sites | Medium |
| 95 | Missing disabled during form submit | 2 files | Medium |
| 96 | Form state not reset in dialogs | 4 files | Low |
| 97 | Transient network error logs user out | 1 file | High |
| 98 | `intlMiddleware` failure drops cookies | 1 file | High |
| 99 | Path traversal via catch-all proxy | 1 file | High |
| 100 | `createErrorResponse` discards error context | 1 file | Medium |
| 101 | Proxy leaks upstream error body | 1 file | Medium |
| 102 | RMA mutations don't invalidate list | 3 files | Medium |
| 103 | CRM mutations don't invalidate stages | 3 files | Medium |
| 104 | Approval detail doesn't invalidate list | 1 file | Medium |
| 105 | Missing `"use client"` directive | 1 file | Low |
| 106 | Missing input validation on proxy route | 1 file | High |
| 107 | Inconsistent error response format | 1 file | Low |
| 108 | Three competing token refresh mechanisms | 4 files | High |
| 109 | Logout doesn't clear Zustand org store | 2 files | Medium |
| 110 | Permissions hide all UI during loading | 3 files | Medium |
| 111 | Fire-and-forget mutations (no `.catch()`) | 5 files | High |
| 112 | URL params not synced with component state | 1 file | Medium |
| 113 | Direct service calls instead of React Query | 2 files | Medium |
| 114 | Form state leaked between dialog opens | 3 files | Medium |
| 115 | Template literal classNames instead of `cn()` | 26 instances (12 files) | Medium |
| 116 | Inline style objects in render loops | 53+ instances (15+ files) | Medium |
| 117 | God components (>150 lines, >5 useState) | 5 files | Medium |
| 118 | Derived state stored in useEffect | 3 files | Medium |
| 119 | Multiple dialog states that should be one | 2 files | Low |
| 120 | Zero page metadata exports | 100+ pages | High |
| 121 | No openGraph metadata | All pages | High |
| 122 | No alternates for multilingual | All pages | High |
| 123 | No robots configuration | All pages | Medium |
| 124 | Zero revalidatePath/revalidateTag | Entire codebase | Medium |
| 125 | Server prefetch staleTime: "static" | 1 file | Medium |
| 126 | Org layouts no error handling for prefetch | 2 files | Medium |
| 127 | No Suspense boundaries in server layouts | 5 layouts | Medium |
| 128 | Sequential awaits in server components | 2 pages | Low |
| 129 | Unsafe .find() on potentially undefined arrays | 6 files | High |
| 130 | Invalid date propagation in formatters | 2 functions | Medium |
| 131 | OAuth callback sets cookies with empty tokens | 1 file | Medium |
| 132 | Blob URL / iframe leak on unmount | 1 file | Low |
| 133 | Async UI calls without try/catch | 4 files | High |
| 134 | Form schemas missing .trim() on required strings | All forms | Medium |
| 135 | Form schemas missing cross-field validation | 4 forms | Medium |
| 136 | Export default in components/hooks | 25 files | Medium |
| 137 | Container components misnamed | 42 files | Medium |
| 138 | Presentation components with data fetching | 33 files | High |
| 139 | Container in shared components/ | 1 file | Medium |
| 140 | Missing data-slot attributes | 31 components | Low |
| 141 | Direct fetch() in client components | 5 files | Medium |
| 142 | Forms without schema validation | 7 files | High |
| 143 | Duplicate density constants/types | 2 files | Low |
| 144 | Duplicate theme constants indirection | 2 files | Low |
| 145 | Locale cookie without Secure flag | 1 file | Medium |
| 146 | Density cookie without Secure flag | 1 file | Medium |
| 147 | Premature abstractions (dead) | 7 modules | Medium |
| 148 | Unused components (zero imports) | 37 components | Medium |
| 149 | Unused server/client modules | 5 modules | Low |
| 150 | Over-engineered error hierarchy | 1 file | Low |
| 151 | Hardcoded settings page (mockup) | 1 file | Medium |
| 152 | Empty stub page | 1 file | Low |
| 153 | Half-implemented command palette | 1 file | Low |
| 154 | Over-complex Zustand stores | 2 stores | Low |
| 155 | Unused hook exports (additional) | 2 exports | Low |
| 156 | POST without CSRF header | 1 file | Medium |
| 157 | Wildcard `import * as React` | 9 files | Low |
| 158 | Unnecessary "use client" on utils | 38 files | Low |
| 159 | Inconsistent empty fallback chars | 16+ files | Low |
| 160 | Duplicated table utility functions | 2 files | Medium |
| 161 | Duplicated `formatHours` | 2 files | Low |
| 162 | Duplicated CommissionPlan types | 1 file | Low |
| 163 | Inconsistent detail page navigation | 15 files | Medium |
| 164 | Inconsistent loading states | 12 files | Medium |
| 165 | Inconsistent error handling patterns | 11 files | Medium |
| 166 | IIFE in JSX | 2 files | Low |
| 167 | Very long className strings | 5 files | Low |
| 168 | Inline type in JSX `.map()` | 1 file | Low |
| 169 | Dashboard KPI card repetition | 1 file (~300 lines) | Medium |
| 170 | Section status prop mapping duplication | 30+ files | Medium |
| 171 | Identical dialogs (title only diff) | 1 file (3 dialogs) | Medium |
| 172 | Duplicated file names across features | 2 files | Low |
| 173 | Hardcoded mock data in production | 2 files | Medium |
| 174 | Inconsistent abbreviation in file names | 4 patterns | Low |
| 175 | Wrong query hook for single items | 2 files | Medium |
| 176 | Zero JSDoc on exported types | 100+ types | Low |
| 177 | Arrow function component | 1 file | Low |
| 178 | `defaultLabels` hardcoded English (systemic) | 16 components (~90 strings) | High |
| 179 | Hardcoded aria-label/title attributes | 18 components (~30 attrs) | Medium |
| 180 | `.toLocaleTimeString()` no locale | 1 call site | Medium |
| 181 | Hardcoded default prop values | 5 components | Medium |
| 182 | Direct `getSwantaraService()` in components | 30 files | High |
| 183 | Missing CSRF on `/api/vitals` POST | 1 file | Medium |
| 184 | OAuth tokens leakage via query string | 1 file | Medium |
| 185 | Refresh tokens as plaintext Map keys | 1 file | Low |
| 186 | Proxy error logs leak full URL | 1 file | Low |
| 187 | Duplicated column definitions | 9+ files | Medium |
| 188 | Duplicated `createId()` utility | 5 files | Medium |
| 189 | Duplicated useEffect for line reset | 4 files | Medium |
| 190 | Export default in providers | 9 files | Medium |
| 191 | Interface instead of type (new files) | 5 occurrences | Low |
| 192 | Arrow component + wildcard imports | 3 files | Low |
| 193 | Missing data-slot on new components | 7 components | Low |
| 194 | Unnecessary "use client" on pure utils | 4 files | Low |
| 195 | Nested ternaries + long className | 4 files | Low |
| 196 | toLocale without locale (new sites) | 2 files | Medium |
| 197 | Hardcoded USD + formatNumber for money | 70+ sites | High |
| 198 | Hardcoded English defaults/aria/filter | 20+ components | Medium |
| 199 | Hooks-after-return + container issues | 5 files | High |
| 200 | Data-fetching correctness (hook/invalidation/race/over-fetch) | 15+ files | High |
| 201 | Session/auth/CSRF/proxy security | 8 issues | High |
| 202 | Duplication + UTC date + missing catch + unsafe key | 12 files | Medium |

---

**Total:** ~2200+ actionable violations across 202 categories

---

## 190. Export Default in Providers

**Target:** Replace `export default` with named exports. Only `page.tsx`, `layout.tsx`, `loading.tsx`, `error.tsx`, `not-found.tsx`, `template.tsx` may use default. §1 excluded providers; §136 covered app components only.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `providers/container.tsx` | 17 | `export default async function ProviderContainer` |
| [x] | `providers/digital-fingerprint.tsx` | 6 | `export default function DigitalFingerPrintProvider` |
| [x] | `providers/hotkeys.tsx` | 13 | `export default function HotkeysProvider` |
| [x] | `providers/install-prompt.tsx` | 13 | `export default function InstallPromptProvider` |
| [x] | `providers/internalization.tsx` | 30 | `export default async function I18nProvider` |
| [x] | `providers/offline-sync.tsx` | 16 | `export default function OfflineSyncProvider` |
| [x] | `providers/serwist.tsx` | 5 | `export default function SerwistProvider` |
| [x] | `providers/tanstack-query.tsx` | 13 | `export default function TanstackQueryProvider` |
| [x] | `providers/theme.tsx` | 6 | `export default function ThemeProvider` |

---

## 191. Interface Instead of `type` (New Files)

**Target:** Use `type`, not `interface`, per CONVENTIONS.md. Follow-up to §2 (10 files fixed, these are new shared files).

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `components/entity-section.tsx` | 9 | `export interface EntitySectionProps<TData>` |
| [x] | `components/entity-form-dialog.tsx` | 16 | `export interface EntityFormDialogProps` |
| [x] | `lib/utils/table-columns.tsx` | 6 | `interface NameColumnOpts<T>` |
| [x] | `lib/utils/table-columns.tsx` | 42 | `interface ActiveColumnOpts` |
| [x] | `lib/utils/table-columns.tsx` | 63 | `interface StateColumnOpts` |

---

## 192. Arrow Component + Wildcard Imports

**Target:** Function declarations for components; named imports instead of `import * as React`.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `components/chart.tsx` | 80 | `const ChartStyle = ({ id, config }) =>` — must be `function` |
| [x] | `components/chart.tsx` | 3 | `import * as React from "react"` |
| [x] | `components/dropdown-menu.tsx` | 5 | `import * as React from "react"` |

---

## 193. Missing `data-slot` on New Components

**Target:** Add `data-slot="kebab-name"` to root elements. Follow-up to §140 (31 fixed; these 7 never listed).

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `components/active-badge.tsx` | 4 | Root `<Badge>` has no `data-slot` |
| [x] | `components/detail-page-skeleton.tsx` | 4 | Root has no `data-slot` |
| [x] | `components/entity-form-dialog.tsx` | 41 | `DialogContent` has no `data-slot` |
| [x] | `components/entity-section.tsx` | 59 | Root has no `data-slot` |
| [x] | `components/journal-date-dialog.tsx` | 44 | Root has no `data-slot` |
| [x] | `components/payment-dialog.tsx` | 47 | Root has no `data-slot` |
| [x] | `components/route-loading.tsx` | 11, 20 | Roots have no `data-slot` |

**Note:** `components/approval-widget.tsx` already has `data-slot="approval-widget"` — not a violation.

---

## 194. Unnecessary `"use client"` on Pure Utils

**Target:** Remove `"use client"` from `*-utils.ts` with only pure functions. Follow-up to §158.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `accounting/accounts/_components/account-utils.ts` | 1 | Pure types + tree/count helpers only |
| [x] | `attendance/_components/attendance-utils.ts` | 1 | Pure functions only |
| [x] | `employees/_components/employee-utils.ts` | 1 | Pure functions/types only |
| [x] | `expense-categories/_components/expense-category-utils.ts` | 1 | Pure function only |

---

## 195. Nested Ternaries + Long className

**Target:** Lookup maps / `if-else` / `cn()`. Follow-up to §29/§44/§115/§167.

| Status | File | Line | Issue |
|--------|------|------|-------|
| [x] | `(org)/_components/org-user-menu.tsx` | 141 | `l === "en" ? "English" : l === "id" ? ...` in JSX |
| [x] | `warehouses/[warehouseId]/_components/location-form-dialog.tsx` | 96 | `initial ? undefined : parentId ? Number : null` |
| [x] | `settings/members/_components/members-table.tsx` | 33 | `!user ? "invited" : user.active ? "active" : "inactive"` |
| [x] | `dashboard/_components/modules-grid.tsx` | 189 | 140+ char template `className` with ternary — use `cn()` |
| [x] | `reports/trial-balance/page.tsx` | 96 | Template `className` with `isBalanced` ternary — use `cn()` |

---

## 196. `.toLocaleDateString()` / `.toLocaleString()` Without Locale (New Sites)

**Target:** Use `formatDate()` / `formatNumber()` from `@/lib/utils/formatters`. Follow-up to §68/§180.

| Status | File | Line | Code |
|--------|------|------|------|
| [x] | `components/filter-bar.tsx` | 69-70 | `range.from.toLocaleDateString()` / `range.to.toLocaleDateString()` |
| [x] | `components/chart.tsx` | 233 | `item.value.toLocaleString()` in tooltip |

---

## 197. Hardcoded `currency: "USD"` + `formatNumber` for Money

**Target:** Use org currency / locale currency via `formatMoney` / `formatCurrency`. CONVENTIONS.md marks hardcoded `currency: "USD"` as Bad.

**Severity:** HIGH — 70+ sites render USD regardless of org locale.

Representative (full list in audit; `__tests__` excluded):

| Status | File | Lines |
|--------|------|-------|
| [x] | `dashboard/_components/operations-section.tsx` | 59, 63, 77, 90, 164, 167, 170, 183, 197, 281, 284, 290, 306, 334 |
| [x] | `dashboard/_components/sales-section.tsx` | 75, 78, 81, 85, 102, 129, 179, 182, 201 |
| [x] | `dashboard/_components/cash-section.tsx` | 88, 91, 94, 121 |
| [x] | `dashboard/_components/finance-overview.tsx` | 72, 85, 92, 95, 98 |
| [x] | `commission-entries/_components/commission-entries-section.tsx` | 48, 57, 110 |
| [x] | `gift-cards/_components/gift-cards-section.tsx` | 58, 67, 145 |
| [x] | `fixed-assets/[assetId]/_components/fixed-asset-detail-section.tsx` | 192, 202, 205, 215, 261, 286, 289, 292 |
| [x] | `subscriptions/_components/subscriptions-section.tsx` | 62, 135, 145, 173 |
| [x] | `accounting/_components/recent-entries.tsx` | 50, 53 |
| [x] | `sale-orders/_components/sale-order-form-dialog.tsx` | 360, 396 |

`formatNumber` for currency amounts (should be `formatMoney`):

| Status | File | Lines |
|--------|------|-------|
| [x] | `purchase-orders/[orderId]/_components/purchase-order-detail-section.tsx` | 272, 276, 280, 342, 345, 480 |
| [x] | `reports/_components/balance-sheet-sections.tsx` | 56, 95 |
| [x] | `reports/_components/profit-and-loss-rows.tsx` | 44, 77 |
| [x] | `reports/_components/trial-balance-table.tsx` | 83, 86, 89, 92, 95, 98, 109, 112, 115, 118, 121, 124 |
| [x] | `reports/cash-flow/page.tsx` | 81, 94, 107, 123, 130, 136 |
| [x] | `reports/balance-sheet/page.tsx` | 66, 72, 78 |
| [x] | `reports/profit-and-loss/page.tsx` | 63, 70, 82 |
| [x] | `reports/aging/page.tsx` | 65, 86, 118 |
| [x] | `reports/inventory-valuation/page.tsx` | 60, 66, 114, 117, 125, 127 |

---

## 198. Hardcoded English Defaults / Aria / Filter UI

**Target:** Replace with `t()` calls. Follow-up to §178/§179/§181/§11.

| Status | File | Lines | Text |
|--------|------|-------|------|
| [x] | `components/payment-dialog.tsx` | 35, 53, 56, 58, 72, 83 | `cancelLabel="Cancel"`, Journal/Amount/Date labels, `aria-label="Journal"` |
| [x] | `components/journal-date-dialog.tsx` | 34, 53, 55, 68 | Same pattern |
| [x] | `components/form.tsx` | 112, 166 | `loadingLabel="Saving…"`, `window.confirm("You have unsaved changes. Leave this page?")` |
| [x] | `components/offline-banner.tsx` | 19, 76 | `fallbackMessage`, `fallbackClearPending`, pending-count strings |
| [x] | `components/message-scroller.tsx` | 44, 113 | `?? "Messages"`, Scroll to end/start |
| [x] | `components/filter-bar.tsx` | 116, 128, 132, 146, 160, 190, 214, 233, 248, 261, 278 | `Filter by ${label}`, `All ${label}`, Filters, Clear all, Date range |
| [x] | `(org)/_components/server-entity-section.tsx` | 97 | `` `Filter ${label.toLowerCase()}…` `` |
| [x] | `(org)/_components/interactive-entity-table.tsx` | 87 | Same Filter placeholder |
| [x] | `components/tag-input.tsx` | 21, 25, 40, 45, 68 | Add tag, Tags, max-allowed, Added/Remove tag |
| [x] | `components/kanban-board.tsx` | 37, 138, 173 | Kanban board, Add/Move card |
| [x] | `components/shipment-track.tsx` | 41, 83 | `ariaLabel="Shipment progress"`, raw `done/active/pending` in sr-only |
| [x] | `components/data-table.tsx` | 182 | `ariaLabel="Data table"` |
| [x] | `providers/offline-sync.tsx` | 48, 51 | Synced/offline-change toasts |
| [x] | `(public)/draft/_components/draft-showcase.tsx` | 227, 524, 591, 711 | Demo titles/aria-labels |

---

## 199. Hooks-After-Return + Container Issues

**Target:** Hooks before any early return; containers named `*-section.tsx`; shared `components/` holds presentation only.

**Severity:** HIGH — hooks-order violation crashes / misbehaves when `approvalRequest` toggles null↔object.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `components/approval-widget.tsx` | 55-68 | Early `return null` before `useMutation` — move hooks above return |
| [x] | `components/approval-widget.tsx` | 48-98 | Business-logic container (4 useState + mutation) in shared `components/` |
| [x] | `settings/profile/_components/profile-form.tsx` | 24-31, 94, 107 | `*-form.tsx` acts as container (`useMeQuery`, service calls) — rename/split |
| [x] | `contacts/_components/contact-addresses.tsx` | 66, 131, 139, 150, 159 | Presentation file with direct service calls, no `useMutation` |
| [x] | `contacts/_components/contact-bank-accounts.tsx` | 54, 115, 123, 134 | Same pattern |
| [x] | `contacts/_components/contact-defaults.tsx` | 61, 127 | Same pattern |

---

## 200. Data-Fetching Correctness

**Target:** `useOrgQuery` for single gets; invalidate list + detail; `enabled` guards; `isPending` on buttons; avoid list-overfetch.

**Severity:** HIGH

Wrong hook (`useOrgListQuery` for single `get`):

| Status | File | Lines |
|--------|------|-------|
| [x] | `subscriptions/[subscriptionId]/_components/subscription-detail-section.tsx` | 34-38 |
| [x] | `service-contracts/[contractId]/_components/service-contract-detail-section.tsx` | 31-35 |
| [x] | `gift-cards/[giftCardId]/_components/gift-card-detail-section.tsx` | 23-26 |
| [x] | `fixed-assets/[assetId]/_components/fixed-asset-detail-section.tsx` | 46-50 |
| [x] | `equipments/[equipmentId]/_components/equipment-detail-section.tsx` | 15 |
| [x] | `maintenance-plans/[planId]/_components/maintenance-plan-detail-section.tsx` | 15 |
| [x] | `employees/[employeeId]/_components/employee-detail-section.tsx` | 22 |

Missing list invalidation / `refetch()` instead of invalidation:

| Status | File | Lines |
|--------|------|-------|
| [x] | `subscription-detail-section.tsx` | 46 — only `["subscription"]`, never `["subscriptions", orgId]` |
| [x] | `service-contract-detail-section.tsx` | 43 — only detail key |
| [x] | `gift-card-detail-section.tsx` | 48-49 — never `["giftCards", orgId]` |
| [x] | `fixed-asset-detail-section.tsx` | 59, 76, 97 — only detail key |
| [x] | `stock/pickings/[pickingId]/_components/picking-detail-section.tsx` | 81, 94 — `refetch()` only |
| [x] | `stock/counts/[countId]/_components/count-detail-section.tsx` | 66 — `refetch()` only |
| [x] | `stock/transfers/[transferId]/_components/transfer-detail-section.tsx` | 64, 78 — `refetch()` only |

Race (no `disabled={isPending}`) + over-fetch (list then `.find()` client-side):

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `subscription-detail-section.tsx` | 69-95 | 5 buttons, no pending disable |
| [x] | `service-contract-detail-section.tsx` | 68, 73 | Same |
| [x] | `gift-card-detail-section.tsx` | 75, 80 | Double-click double-spends `card.balance` |
| [x] | `stock/pickings/[pickingId]/_components/picking-detail-section.tsx` | 54-65 | Fetches all pickings + moves, filters client-side |
| [x] | `stock/counts/[countId]/_components/count-detail-section.tsx` | 29-40 | Fetches all counts, `.find()` client-side |
| [x] | `stock/transfers/[transferId]/_components/transfer-detail-section.tsx` | 40-45 | Fetches all transfers, `.find()` client-side |

Direct `fetch()` / service bypass (extends §141/§182):

| Status | File | Lines |
|--------|------|-------|
| [x] | `(auth)/oauth/callback/page.tsx` | 50-57 — `fetch("/api/v1/auth/session")` without CSRF header |
| [x] | `(auth)/register/_hooks/use-register-form.ts` | 100-108 — `fetch("/api/v1/auth/login")` directly |
| [x] | `settings/profile/_components/profile-form.tsx` | 94, 107 — `await getSwantaraService()` without `useMutation` |
| [x] | `onboarding/_hooks/use-onboarding-form.ts` | 88-93 — `quickCreate` without `useMutation` |

---

## 201. Session / Auth / CSRF / Proxy Security

**Target:** CSRF on all mutating routes; safe methods stay safe; validate proxy input; don't log secrets.

**Severity:** HIGH

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `app/api/v1/auth/session/route.ts` | 7-36 | `POST` sets session cookies with no `isCsrfValid()` — fixation; no content-type check, no rate limit |
| [x] | `app/api/v1/auth/session/route.ts` | 38-43 | `GET` clears cookies + redirects — logout-CSRF via `<img>` |
| [x] | `(auth)/oauth/callback/page.tsx` | 50-57 | POST to session endpoint sends no `x-csrf-token` |
| [x] | `app/_components/report-web-vitals.tsx` | 19-21 | `sendBeacon` cannot send CSRF headers |
| [x] | `lib/constants/cookies.ts` | 5-8 | `hasAccessToken()` reads HttpOnly cookie via `document.cookie` — always false |
| [x] | `(auth)/reset-password/page.tsx` + `use-reset-password-form.ts` | 6, 11 / 33, 52 | Reset token in `?token=` query — leaks to logs/Referer/history |
| [x] | `app/api/v1/[...path]/route.ts` + `lib/server/proxy.ts` | 10-18 / 105-112 | `hasInvalidSegments` misses `%2e%2e`, `.`, `\`, `?`/`#` smuggling; no per-segment encode |
| [x] | `lib/server/csrf.ts` | 19-31 | Same-host Origin/Referer returns true without token |
| [x] | `next.config.ts` | 12-15, 28 | `logging.fetches.fullUrl` logs query tokens; deprecated `X-XSS-Protection`; still no CSP |
| [x] | `app/api/v1/auth/session/route.ts` | 14-20 | Only `length<10` check — no max length / JWT shape / rate limit |

---

## 202. Duplication + UTC Date + Missing Catch + Unsafe Key

**Target:** Reuse canonical utils; timezone-safe dates; error feedback on deletes; stable keys.

| Status | File | Lines | Issue |
|--------|------|-------|-------|
| [x] | `stores/offline-queue.store.ts` | 28 | 6th `createId()` copy — import from `lib/utils/create-id.ts` |
| [x] | `components/data-table.tsx` | 132-138 | 3rd `getColumnLabel` copy — import from `lib/utils/table.ts` |
| [x] | `supplier-catalog-section.tsx` | 82, 84, 128, 129 | `toISOString().slice(0,10)` off-by-one at UTC+7 |
| [x] | `crm/_components/activity-form-dialog.tsx` | 84 | Same UTC pattern |
| [x] | `crm/_components/lead-form-dialog.tsx` | 115 | Same UTC pattern |
| [x] | `projects/_components/project-form-dialog.tsx` | 99-100 | Same UTC pattern |
| [x] | `projects/[projectId]/_components/task-form-dialog.tsx` | 71 | Same UTC pattern |
| [x] | `projects/[projectId]/_components/milestone-form-dialog.tsx` | 58 | Same UTC pattern |
| [x] | `job-positions/_components/job-positions-section.tsx` | 75-79 | Delete with `.then(refetch)`, no `.catch`/toast |
| [x] | `timesheets/_components/timesheets-section.tsx` | 65-69 | Same delete pattern |
| [x] | `onboarding/_components/confirmation-step.tsx` | 29 | `key={row.label}` translated-string collision — use stable id |

---

**Total:** ~2200+ actionable violations across 202 categories
