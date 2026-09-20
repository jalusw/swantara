# Swantara

Generic ERP for small to medium businesses. Covers the boring business cycle: sell, buy, make, stock, bill, pay, close.

## Product

### What it covers

- Sell: prospect, quote, order, delivery, invoice, payment, POS, returns
- Buy: purchase request, quote request, PO, receipt, bill, payment, expenses
- Make: recipe, production order, QC
- Stock: items, stock movements, transfers, counts
- Money: ledger, payables, receivables, tax, reports
- People and assets: employees, payroll, assets, service, subscriptions, projects

### Modules

| Flow | Modules |
|------|---------|
| O2C | Sales, POS, Checkout (Xendit), Gift Card, Commission |
| P2P | Procurement, Expense |
| P2M | Manufacturing, Quality |
| I2F | Inventory, Products |
| R2R | Accounting, Reference, Localization (ID), Reporting |
| H2R | Payroll |
| A2R | Asset |
| S2C | Service, Returns (RMA) |
| Sub | Subscription |
| P2Pft | Project |
| Cross-cutting | IAM (JWT, RBAC), Inter-Organization, Approvals, Audit, Document vault |

### Flows

- O2C: Quote -> Order -> Delivery -> Invoice -> Payment
- P2P: Purchase request -> Quote request/PO -> Receive -> Bill -> Pay
- P2M: Demand -> Production order -> Consume -> Produce
- I2F: Inbound, outbound, transfer, count, return
- R2R: Post to ledger, close period, report
- H2R: Hire -> Contract -> Attendance -> Payroll run -> Pay
- A2R: Acquire -> Depreciate -> Dispose
- S2C / Sub / Project: shop task -> bill, recurring bill, project WIP vs billing

See `docs/glossary.md` for the generic business terms used across the codebase.

## Technical

Modular monolith + web client, containerized for local dev. No staging/prod yet.

### Service (`service/`)

| Component | Technology |
|-----------|-----------|
| Language | Go 1.25 |
| HTTP | Fiber v3 |
| Database | PostgreSQL 16 (GORM) |
| Migrations | Goose |
| Queue | Redis 7 + Asynq workers |
| Auth | JWT, RBAC |
| Validation | go-playground/validator |
| Payment | Xendit |
| Mail | SMTP via go-mail |
| Logging | slog + tint |
| Config | Viper (.env) |
| Docs | Swagger (Swaggo) |
| Test | httpexpect, gofakeit, go-sqlmock, testify |
| Lint | golangci-lint |

Architecture per domain: `Request -> Handler -> Service -> DAO -> PostgreSQL`. Handlers parse HTTP and write responses. Services hold business rules. DAOs embed generic `dao.Base[E]`. Shared code lives in `kernel/` and `httpx/`. No cross-domain imports.

See `service/README.md` for layer rules, error handling, testing, and queue patterns.

### Web (`web/`)

Next.js 16, React 19, Tailwind 4, shadcn + Base UI, TanStack Query/Table, Zustand, React Hook Form + Zod, next-intl, Serwist PWA for low-end/offline use. Biome + `tsc` + Vitest + Playwright.

See `web/README.md` for feature module structure and container/presentation patterns.

### Layout

```
service/            Go API + workers + migrations
web/                Next.js client
docker-compose.yml  postgres, redis, api, worker, web
justfile            dev, lint, test, build entrypoints
LICENSE             AGPL-3.0
```

Infra services: `postgres` (5432), `redis` (6379), `api` (8080), `worker` (Asynq), `web` (3000). Test profile adds `postgres-test` (5433), `api-e2e`, `worker-e2e`, `k6`.

## Getting Started

### Prerequisites

- Node 20+, pnpm 9+
- Go 1.25+
- Docker + Docker Compose
- `just`

Run `just doctor` to verify toolchain.

### Setup

```sh
cp service/.env.example service/.env
cp web/.env.example web/.env
just setup
docker compose up -d postgres redis
```

Run migrations and seed:

```sh
just service-migrate
just service-seed
```

Or from `service/` directly:

```sh
cp .env.example .env
just db-up
just migrate
just seed
```

### Run

Full stack via compose:

```sh
docker compose up
```

Or split dev servers:

```sh
just dev-web
just dev-service
```

Service-only with live reload:

```sh
cd service
just dev
just dev-queue
```

### Verify

- Web: http://localhost:3000
- API health: http://localhost:8080/api/v1/health
- Swagger: http://localhost:8080/swagger/index.html

### Useful Commands

```sh
just lint
just typecheck
just test
just build
just service-lint
just service-test
```

## License

Copyright (C) 2026 J Satriani Wijaya

Swantara is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version. See `LICENSE` for details.
