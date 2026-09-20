# Swantara Service

Backend API for **Swantara** — a generic open full-cycle ERP platform for small to medium businesses.

Go 1.25, Fiber v3, PostgreSQL (GORM), Redis + Asynq workers, JWT auth, Goose migrations, Swagger docs. Domain-oriented modular monolith with strict layering: `Request -> Handler -> Service -> DAO -> PostgreSQL`. No cross-domain imports between business modules.

Item overview, setup, and license live in `../README.md`. This file is the single source of truth for backend architecture and coding conventions.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.25 |
| HTTP Framework | Fiber v3 |
| Database | PostgreSQL (GORM) |
| Migrations | Goose |
| Background Jobs | Asynq (Redis-backed) |
| Auth | JWT |
| Validation | go-playground/validator |
| Email | SMTP via go-mail |
| Logging | slog + tint |
| Config | Viper (.env) |
| API Docs | Swagger (Swaggo) |
| Testing | httpexpect, gofakeit, go-sqlmock, testify |

## Layout

```
cmd/http/main.go        Fiber server entry point
cmd/queue/main.go       Asynq worker entry point
cmd/seeder/main.go      Database seeder
internal/<domain>/      models, errors, ports, dao, service, handler
internal/kernel/        Shared infrastructure (dao, model, query, amount, audit, state, sequence)
internal/httpx/         HTTP server, middleware, response helpers, context, validation
migrations/             Goose SQL migrations
test/                   Integration and e2e suites
```

## Setup

```bash
cp .env.example .env
just db-up
just migrate
just seed
just dev
just dev-queue
```

API at `http://localhost:8080`, Swagger at `/swagger/index.html`. Run `just lint` and `just test` before committing.

---
# Architecture
Patterns for organizing code in the Go backend (`service/`). Prioritize
simplicity, clarity, and elegance. Future agents: follow these patterns
unless there's a good reason not to.

---

## Core Principles

1. **Simple over clever** — Code should be obvious at a glance
2. **Consistent over optimal** — Follow existing patterns, even if slightly longer
3. **Readable over terse** — Prefer clarity over conciseness
4. **Domain isolation** — Each domain is self-contained; no cross-domain imports
5. **One responsibility per file** — Each file does one thing well
6. **Explicit error handling** — Never swallow errors; always propagate or log

---

## Layer Architecture

Domain-oriented modular monolith. Each business domain is a self-contained
package with clear layer boundaries.

```
Request → Handler → Service → DAO → Database
              ↓         ↓       ↓
         Response    Errors  Errors
```

### Layer Responsibilities

| Layer | Package | Responsibility | Knows About |
|-------|---------|----------------|-------------|
| Transport | `handler/` | Parse HTTP, write HTTP, call service | HTTP concepts only |
| Domain | `service.go` | Business rules, orchestration | Domain + DAO interfaces |
| Persistence | `dao.go` | Database queries | GORM + domain models |
| Cross-cutting | `httpx/`, `kernel/` | Auth, logging, pagination, amounts | Infrastructure |

### Dependency Rules

- Handler → Service → DAO (never reverse)
- Service never imports from `handler/`
- Handler never contains business logic
- DAO never contains business logic
- All layers may import from `kernel/` and `httpx/`
- Domains never import from each other (use interfaces or events)

---

## Domain Module Structure

Every business domain follows this exact structure:

```
internal/<domain>/
├── doc.go              ← Package-level doc comment
├── models.go           ← GORM entity structs (embed kernel/model.Base)
├── errors.go           ← Sentinel errors (var Err... = errors.New(...))
├── ports.go            ← Interface declarations (dependencies the service needs)
├── dao.go              ← DAO interface + GORM implementation
├── dao_mock.go         ← Mock DAO for unit tests
├── service.go          ← Service struct + business logic methods
├── service_test.go     ← Table-driven unit tests
├── fixtures.go         ← gofakeit-based test fixture factories
└── handler/
    ├── doc.go            ← Package doc
    ├── routes.go         ← Register(api fiber.Router, guards httpx.RouteGuards)
    └── <domain>_handler.go ← Handler struct + HTTP methods
```

### What Goes Where

| Concern | Location | Example |
|---------|----------|---------|
| HTTP request parsing | `handler/` | `httpx.BindAndValidate(c, &req)` |
| HTTP response writing | `handler/` | `httpx.CreateSuccessResponse(c, msg, data)` |
| Business rule validation | `service.go` | `if qty < 0 { return ErrNegativeQty }` |
| Database queries | `dao.go` | `s.db.WithContext(ctx).Where(...).Find(...)` |
| Domain error definitions | `errors.go` | `var ErrNotFound = errors.New("not found")` |
| GORM model definitions | `models.go` | `type Order struct { model.Base; ... }` |
| Interface declarations | `ports.go` | `type XDAO interface { ... }` |
| Route registration | `handler/routes.go` | `func (h H) Register(api, guards)` |

### Example: Products Domain

```
internal/products/
├── doc.go
├── models.go           ← Item, ItemVariant structs
├── errors.go           ← ErrProductNotFound, ErrVariantSKUConflict
├── ports.go            ← Interface declarations
├── dao.go              ← ProductDAO interface + productDAO impl
├── dao_mock.go         ← MockProductDAO
├── service.go          ← ProductService with Create/Update/Delete/List
├── service_test.go     ← Table-driven tests
├── fixtures.go         ← ProductFixture(), VariantFixture()
└── handler/
    ├── doc.go
    ├── routes.go
    └── product_handler.go ← ProductHandler with CRUD endpoints
```

---

## Handler Pattern

### Structure

Handlers are methods on a struct that holds service dependencies:

```go
type ProductHandler struct {
    service products.ProductService
}

func NewProductHandler(service products.ProductService) ProductHandler {
    return ProductHandler{service: service}
}
```

### Method Signature

Always `func (h Handler) Method(c fiber.Ctx) error`:

```go
func (h ProductHandler) List(c fiber.Ctx) error {
    // ...
}
```

### Request Flow

```
1. Parse path params    → c.Params("id")
2. Bind + validate body → httpx.BindAndValidate(c, &request)
3. Call service         → h.service.Get(c.Context(), id)
4. Handle errors        → switch on errors.Is(err, ...)
5. Write response       → httpx.CreateSuccessResponse(c, msg, data)
```

### Complete Handler Example

```go
func (h ProductHandler) Create(c fiber.Ctx) error {
    orgID, ok := httpx.CallerOrganizationID(c)
    if !ok {
        return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
    }

    var request CreateProductRequest
    if !httpx.BindAndValidate(c, &request) {
        return nil
    }

    item, err := h.service.Create(c.Context(), orgID, request.Name, request.SKU, request.Price)
    if err != nil {
        switch {
        case errors.Is(err, products.ErrAlreadyExists):
            return httpx.CreateConflictResponse(c, "Item already exists.", err)
        default:
            httpx.RequestLog(c).Error("create item failed", "org_id", orgID, "error", err)
            return httpx.CreateInternalServerErrorResponse(c, "Failed to create item.", err)
        }
    }

    return httpx.CreateCreatedResponse(c, "Item created.", newProductResponse(item))
}
```

### Route Registration

Each handler registers routes in `routes.go`:

```go
func (h ProductHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
    products := api.Group("/products", guards.AuthN)

    products.Get("/", guards.Guard("item", "view"), h.List)
    products.Post("/", guards.Guard("item", "create"), h.Create)
    products.Get("/:id", guards.Guard("item", "view"), h.Get)
    products.Put("/:id", guards.Guard("item", "update"), h.Update)
    products.Delete("/:id", guards.Guard("item", "delete"), h.Delete)
}
```

Route parameters are always `:id`. The resource name is inferred from
the group. Idempotency guards are added for mutation endpoints that
must not be duplicated.

---

## Service Pattern

### Structure

Services contain business logic. They hold DAO dependencies and
never import HTTP concepts:

```go
type ProductService struct {
    dao products.ProductDAO
}

func NewService(dao products.ProductDAO) *ProductService {
    return &ProductService{dao: dao}
}
```

### Method Pattern

Services return domain errors, never HTTP status codes:

```go
func (s *ProductService) Create(ctx context.Context, orgID uint64, name, sku string, price float64) (*Item, error) {
    existing, err := s.dao.FindBySKU(ctx, orgID, sku)
    if err != nil && !errors.Is(err, ErrNotFound) {
        return nil, err
    }
    if existing != nil {
        return nil, ErrAlreadyExists
    }

    item := &Item{
        OrganizationID: orgID,
        Name:           &name,
        SKU:            &sku,
        Price:          price,
        State:          ProductStateActive,
    }
    return s.dao.Create(ctx, item)
}
```

### Transactions

Use `db.Transactioner` for multi-table atomic operations:

```go
type OrderService struct {
    txer  db.Transactioner
    dao   OrderDAO
    lines OrderLineDAO
}

func (s *OrderService) CreateWithLines(ctx context.Context, ...) (*Order, error) {
    var result *Order
    err := s.txer.Run(ctx, func(tx *gorm.DB) error {
        order, err := s.dao.Create(ctx, &order)
        if err != nil {
            return err
        }
        for _, line := range lines {
            if _, err := s.lines.Create(ctx, &line); err != nil {
                return err
            }
        }
        result = order
        return nil
    })
    return result, err
}
```

---

## DAO Pattern

### Generic Base

Every domain DAO embeds `dao.Base[E]` which provides standard CRUD:

```go
type ProductDAO interface {
    dao.CRUD[Item]
    FindBySKU(ctx context.Context, orgID uint64, sku string) (*Item, error)
}

type productDAO struct {
    dao.Base[Item]
    db *gorm.DB
}

func NewProductDAO(db *gorm.DB) ProductDAO {
    return productDAO{Base: dao.NewBase[Item](db), db: db}
}
```

### Custom Queries

Domain-specific queries go in the DAO implementation:

```go
func (s productDAO) FindBySKU(ctx context.Context, orgID uint64, sku string) (*Item, error) {
    return s.Search(ctx, "organization_id", orgID, "sku", sku)
}
```

### Query Application

Use `query.ApplyQuery()` for standardized filtering, sorting,
and pagination:

```go
func (s productDAO) List(ctx context.Context, q *query.Query) (*query.Page[Item], error) {
    return s.List(ctx, q)
}
```

### Model Patterns

Embed `model.Base` in all entities:

```go
type Item struct {
    model.Base
    OrganizationID uint64    `gorm:"not null" json:"organization_id"`
    Name           *string   `json:"name"`
    SKU            *string   `json:"sku"`
    Price          float64   `gorm:"type:numeric(18,4)" json:"price"`
    State          string    `gorm:"type:text" json:"state"`
}
```

Rules:
- Nullable fields use pointers (`*uint64`, `*string`, `*time.Time`)
- GORM column types use `gorm:"type:..."` for precision (`numeric(18,4)`)
- Always include `json` tags
- Add `TableName()` method when GORM pluralization is wrong
- Use domain constants for state values (not string literals)

---

## Error Handling Architecture

### Error Flow

```
Domain Error (sentinel) → Handler (switch on errors.Is) → HTTP Response
                         ↓
                    Log with context
```

### Sentinel Errors

Define in `errors.go`. Named with `Err` prefix:

```go
var (
    ErrNotFound      = errors.New("resource not found")
    ErrAlreadyExists = errors.New("resource already exists")
    ErrInvalidState  = errors.New("invalid state transition")
)
```

### Handler Error Mapping

Handlers map domain errors to HTTP status codes via `errors.Is()`:

```go
switch {
case errors.Is(err, products.ErrNotFound):
    return httpx.CreateNotFoundResponse(c, "Item not found.")
case errors.Is(err, products.ErrAlreadyExists):
    return httpx.CreateConflictResponse(c, "Item already exists.", err)
default:
    httpx.RequestLog(c).Error("operation failed", "id", id, "error", err)
    return httpx.CreateInternalServerErrorResponse(c, "Failed.", err)
}
```

### Response Helpers

Always use `httpx` helpers. Never construct responses manually:

| Status | Helper |
|--------|--------|
| 200 | `httpx.CreateSuccessResponse(c, msg, data)` |
| 201 | `httpx.CreateCreatedResponse(c, msg, data)` |
| 204 | `httpx.CreateNoContentResponse(c)` |
| 400 | `httpx.CreateBadRequestResponse(c, msg, err)` |
| 401 | `httpx.CreateUnauthorizedErrorResponse(c, msg, err)` |
| 403 | `httpx.CreateForbiddenErrorResponse(c, msg, err)` |
| 404 | `httpx.CreateNotFoundResponse(c, msg)` |
| 409 | `httpx.CreateConflictResponse(c, msg, err)` |
| 422 | `httpx.CreateUnprocessableEntityErrorResponse(c, msg, fieldErrors)` |
| 500 | `httpx.CreateInternalServerErrorResponse(c, msg, err)` |

---

## Middleware Architecture

### Route Guards

Middleware is composed via `httpx.RouteGuards` which provides:

| Guard | Purpose |
|-------|---------|
| `AuthN` | Extracts user from JWT, sets `httpx.LocalUserID` |
| `Guard(resource, action)` | Checks membership permission in the organization |
| `IdempotencyGuard` | Prevents duplicate mutation execution |

### Usage in Routes

```go
products := api.Group("/products", guards.AuthN)
products.Get("/", guards.Guard("item", "view"), h.List)
products.Post("/", guards.Guard("item", "create"), h.Create)
```

### Custom Middleware

New middleware goes in `httpx/middleware.go`:

```go
func RateLimitGuard(limit int) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // ... rate limiting logic
        return c.Next()
    }
}
```

### Context Values

Middleware sets values via `c.Locals()`. Handlers read them:

```go
// Middleware sets:
c.Locals(httpx.LocalUserID, user.ID)
c.Locals(httpx.LocalOrganizationID, orgID)

// Handler reads:
actorID, ok := httpx.CallerID(c)
orgID, ok := httpx.CallerOrganizationID(c)
```

---

## Request/Response Patterns

### Request Binding

Always use `httpx.BindAndValidate`. Returns `false` and writes
error response automatically:

```go
var request CreateProductRequest
if !httpx.BindAndValidate(c, &request) {
    return nil
}
```

### Request Types

Define in `handler/` files, colocated with the handler:

```go
type CreateProductRequest struct {
    Name  string  `json:"name" validate:"required"`
    SKU   string  `json:"sku" validate:"required"`
    Price float64 `json:"price" validate:"required,min=0"`
}
```

### Response Types

Define response transformers as unexported functions:

```go
func newProductResponse(p *Item) ProductResponse {
    return ProductResponse{
        ID:    p.ID,
        Name:  p.Name,
        SKU:   p.SKU,
        Price: p.Price,
    }
}
```

### Pagination

Use `httpx.CreateSuccessResponseWithMeta` with `httpx.PaginateResponse`:

```go
page, err := h.dao.List(c.Context(), q)
if err != nil {
    return httpx.CreateInternalServerErrorResponse(c, "Failed.", err)
}
meta := httpx.PaginateResponse(page, q)
return httpx.CreateSuccessResponseWithMeta(c, "Retrieved.", newProductsResponse(page.Data), meta)
```

---

## Configuration Architecture

### Loading

Config loaded via Viper from `.env` files:

```go
cfg, err := config.New("", config.ConfigFilePath())
```

### Structure

Fields use `mapstructure` tags:

```go
type Config struct {
    ApplicationName string `mapstructure:"APP_NAME"`
    DatabaseHost    string `mapstructure:"DB_HOST"`
    DatabasePort    int    `mapstructure:"DB_PORT"`
    JWTSecret       string `mapstructure:"JWT_SECRET"`
}
```

### Access Pattern

Pass config to constructors:

```go
cfg, _ := config.New("", config.ConfigFilePath())
db, _ := database.New(cfg)
logger := logger.New(cfg)
server := httpx.New(cfg, logger)
```

---

## Testing Architecture

### Unit Tests

Table-driven tests in `_test.go` files:

```go
func TestProductService_Create(t *testing.T) {
    tests := []struct {
        name    string
        sku     string
        wantErr error
    }{
        {"valid item", "SKU-001", nil},
        {"duplicate SKU", "SKU-001", ErrAlreadyExists},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockDAO := &MockProductDAO{}
            svc := NewService(mockDAO)
            // ... test logic
        })
    }
}
```

### Mocks

DAO mocks in `dao_mock.go`. Used for service-level unit tests:

```go
mockDAO := &MockProductDAO{}
mockDAO.FindFunc = func(ctx context.Context, id uint64) (*Item, error) {
    return &Item{State: ProductStateActive}, nil
}
```

### Fixtures

`gofakeit`-based factories with functional options:

```go
func ProductFixture(opts ...func(*Item) *Item) *Item {
    p := &Item{
        Base:  model.Base{ID: gofakeit.Uint64()},
        Name:  gofakeit.Ptr(gofakeit.ProductName()),
        SKU:   gofakeit.Ptr(gofakeit.SKU()),
        Price: gofakeit.Price(10, 1000),
    }
    for _, opt := range opts {
        opt(p)
    }
    return p
}
```

### Integration Tests

In `test/integration/` with build tag `//go:build integration`:

```go
//go:build integration

func TestOrderFlow(t *testing.T) {
    db := testutil.SetupTestDB()
    defer testutil.CleanTables(t, db)
    // ... test with real database
}
```

### E2E Tests

In `test/e2e/` with build tag `//go:build e2e`. Uses `httpexpect`:

```go
//go:build e2e

func TestCreateOrder(t *testing.T) {
    e := httpexpect.New(t, baseURL)
    e.POST("/api/v1/organizations/{id}/orders").
        WithPath("id", orgID).
        WithJSON(payload).
        Expect().Status(http.StatusCreated)
}
```

---

## Queue Architecture

### Task Definitions

In `internal/queue/tasks/`:

```go
const TypeSendEmail = "send-email"

type SendEmailPayload struct {
    UserID uint64 `json:"user_id"`
    Token  string `json:"token"`
}
```

### Task Handlers

In `internal/queue/handlers/`:

```go
func SendEmailHandler(ctx context.Context, t *asynq.Task) error {
    var payload tasks.SendEmailPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return err
    }
    // ... process
    return nil
}
```

### Enqueuing

Use `TaskEnqueuer` interface:

```go
payload, _ := json.Marshal(tasks.SendEmailPayload{UserID: user.ID})
_, err := s.enqueuer.EnqueueContext(ctx, tasks.TypeSendEmail, payload,
    asynq.MaxRetry(5),
    asynq.Timeout(30*time.Second),
)
```

---

## Logging Architecture

### Logger

`slog` with optional `tint` for pretty output. Structured key-value
pairs for machine parsing:

```go
httpx.RequestLog(c).Info("order created", "id", order.ID, "amount", order.Amount)
httpx.RequestLog(c).Error("payment failed", "id", paymentID, "error", err)
```

### Levels

| Level | When |
|-------|------|
| `Info` | Successful operations, state transitions |
| `Warn` | Expected failures (invalid input, auth failures) |
| `Error` | Unexpected failures (DB errors, external service failures) |

### Sensitive Data

Never log passwords, tokens, or PII. Use `helper.MaskEmail()`:

```go
httpx.RequestLog(c).Warn("login failed", "email", helper.MaskEmail(email))
```

---

## Composition Patterns

### Middleware Chaining

Compose middleware via route groups and guards:

```go
// Public routes
api.Post("/login", h.Login)

// Authenticated routes
auth := api.Group("", guards.AuthN)

// Org-scoped routes
orgs := auth.Group("/organizations/:id")

// Domain routes with specific permissions
products := orgs.Group("/products")
products.Get("/", guards.Guard("item", "view"), h.List)
products.Post("/", guards.Guard("item", "create"), h.Create)
```

### DAO Composition

Services compose multiple DAOs for cross-domain operations:

```go
type InvoiceService struct {
    txer    db.Transactioner
    orders  OrderDAO
    entries EntryDAO
    pdf     PDFGenerator
}
```

### Generic CRUD

`dao.Base[E]` provides standard operations. Domains extend with
custom queries:

```go
type ProductDAO interface {
    dao.CRUD[Item]                    // Create, Find, Update, Delete, List, Search, Exists, Count
    FindBySKU(ctx context.Context, orgID uint64, sku string) (*Item, error)
    FindActive(ctx context.Context, orgID uint64) ([]Item, error)
}
```

### Response Composition

Build response types from domain models via transformer functions:

```go
func newOrderResponse(o *Order) OrderResponse {
    return OrderResponse{
        ID:     o.ID,
        Name:   o.Name,
        State:  o.State,
        Amount: o.Amount,
        Lines:  newOrderLineResponses(o.Lines),
    }
}
```

---

## Enforcement Checklist

Future agents: verify these before committing:

- [ ] Domain packages are self-contained (no cross-domain imports)
- [ ] Handler methods follow `func (h Handler) Method(c fiber.Ctx) error`
- [ ] `httpx.BindAndValidate` used for all request parsing
- [ ] `httpx` response helpers used for all HTTP responses
- [ ] Sentinel errors defined in `errors.go`, not in handler or service
- [ ] Handler maps domain errors to HTTP via `errors.Is()` switch
- [ ] Service returns domain errors, never HTTP status codes
- [ ] DAOs embed `dao.Base[E]` for standard CRUD
- [ ] Models embed `model.Base` with proper GORM/JSON tags
- [ ] Tests are table-driven with descriptive names
- [ ] Mock DAOs in `dao_mock.go` files
- [ ] Fixtures use `gofakeit` with functional options
- [ ] Logging uses `httpx.RequestLog(c)` for structured context
- [ ] No sensitive data in logs (use `helper.MaskEmail`)
- [ ] Context passed via `c.Context()`, not custom context keys
- [ ] Linter passes (`just lint` or `golangci-lint run`)

---

## Anti-Patterns to Avoid

1. **God handlers** — Handlers doing too much (>200 lines, >5 endpoints).
   Split into multiple handler files or extract shared logic to services.

2. **Business logic in handlers** — Validation rules, state checks, or
   calculations in handler methods. Extract to services.

3. **Stringly-typed errors** — Using string comparisons for error matching.
   Use sentinel errors with `errors.Is()`.

4. **Raw `c.JSON()` responses** — Constructing responses manually.
   Use `httpx` helpers for consistency.

5. **Cross-domain imports** — Domain A importing Domain B's internals.
   Use interfaces or events for cross-domain communication.

6. **Direct `db` access in handlers** — Handlers calling GORM directly.
   Always go through the DAO layer.

7. **Nil pointer panics** — Not checking for nil after `Find` operations.
   Always handle the case where the resource doesn't exist.

8. **Missing error logging** — Swallowing errors with `_ = err`.
   Log at the handler level with `httpx.RequestLog(c).Error(...)`.

9. **Hardcoded strings** — Magic strings in business logic.
   Extract to constants in `models.go` or domain-specific files.

10. **Untested business logic** — Service methods without corresponding
    tests. Every service method needs a table-driven test.

---

# Conventions
This document describes the established patterns and conventions for the
Go backend at `service/`. Follow these to maintain consistency across
the codebase.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Package Structure](#package-structure)
3. [File Naming](#file-naming)
4. [Error Handling](#error-handling)
5. [HTTP Handlers](#http-handlers)
6. [Database & DAOs](#database--daos)
7. [Validation](#validation)
8. [Logging](#logging)
9. [Context Handling](#context-handling)
10. [Configuration](#configuration)
11. [Testing](#testing)
12. [Queue & Background Jobs](#queue--background-jobs)
13. [Code Style](#code-style)

---

## Architecture Overview

Domain-oriented modular monolith. Each business domain is a
self-contained package under `internal/` with its own models,
data access, business logic, and HTTP handlers.

```
cmd/
  http/main.go        ← Fiber server entry point
  queue/main.go       ← Asynq worker entry point
  seeder/main.go      ← Database seeder
internal/
  <domain>/
    models.go         ← GORM entity structs
    errors.go         ← Sentinel errors
    service.go        ← Business logic (no HTTP awareness)
    dao.go            ← Data access interface + implementation
    dao_mock.go       ← Mock DAO for tests
    fixtures.go       ← Test fixtures (gofakeit)
    handler/
      routes.go       ← Route registration
      <domain>_handler.go ← HTTP handlers
      doc.go          ← Package documentation
kernel/               ← Shared infrastructure (dao, model, query, amount, audit, state, sequence)
httpx/                ← HTTP server, middleware, response helpers, context, validation
```

### Layer Responsibilities

| Layer | Location | Responsibility |
|-------|----------|----------------|
| Transport | `handler/` | Parse HTTP request, call service, write HTTP response |
| Domain | `service.go` | Business rules, orchestration, validation |
| Persistence | `dao.go` | Database queries via GORM |
| Cross-cutting | `httpx/`, `kernel/` | Auth, logging, pagination, audit, amounts |

Services never import from `handler/`. Handlers never contain
business logic.

---

## Package Structure

Each domain package follows this exact layout:

```
internal/<domain>/
  doc.go              ← Package-level doc comment
  models.go           ← GORM models (embed kernel/model.Base)
  errors.go           ← Sentinel errors (var Err... = errors.New(...))
  ports.go            ← Interface declarations (dependencies the service needs)
  dao.go              ← DAO interface + GORM implementation
  dao_mock.go         ← Mock implementation for tests
  service.go          ← Service struct + methods
  service_test.go     ← Unit tests (table-driven)
  fixtures.go         ← gofakeit-based test fixtures
  handler/
    doc.go            ← Package doc
    routes.go         ← Register(api fiber.Router, guards httpx.RouteGuards)
    <domain>_handler.go ← Handler struct + HTTP methods
```

### What Belongs Where

| Concern | Location | Example |
|---------|----------|---------|
| HTTP request parsing | `handler/` | `c.Bind().Body(&request)` |
| HTTP response writing | `handler/` | `httpx.CreateSuccessResponse(c, msg, data)` |
| Business rule validation | `service.go` | `if qty < 0 { return ErrNegativeQty }` |
| Database queries | `dao.go` | `s.db.WithContext(ctx).Where(...).Find(...)` |
| Domain error definitions | `errors.go` | `var ErrNotFound = errors.New("not found")` |
| Interface declarations | `ports.go` | `type XDAO interface { ... }` |
| GORM model definitions | `models.go` | `type Order struct { model.Base; ... }` |
| Query parameter parsing | `httpx/query.go` | `httpx.ParseQuery(c, allowlist)` |
| Request binding + validation | `httpx/request.go` | `httpx.BindAndValidate(c, &req)` |
| Route registration | `handler/routes.go` | `func (h H) Register(api, guards)` |

---

## File Naming

All files use `snake_case.go`. Suffixes indicate purpose:

| Suffix | Purpose | Example |
|--------|---------|---------|
| `_handler.go` | HTTP handler struct + methods | `product_handler.go` |
| `routes.go` | Route registration | `routes.go` |
| `models.go` | Domain model structs | `models.go` |
| `errors.go` | Sentinel error variables | `errors.go` |
| `service.go` | Business logic | `service.go` |
| `dao.go` | Data access objects | `dao.go` |
| `dao_mock.go` | Mock DAO for tests | `dao_mock.go` |
| `fixtures.go` | Test fixture factories | `fixtures.go` |
| `doc.go` | Package documentation | `doc.go` |
| `_test.go` | Standard Go tests | `service_test.go` |

---

## Error Handling

### Sentinel Errors

Define domain-specific sentinel errors in `errors.go`:

```go
var (
    ErrNotFound       = errors.New("resource not found")
    ErrAlreadyExists  = errors.New("resource already exists")
    ErrInvalidState   = errors.New("invalid state transition")
    ErrUnauthorized   = errors.New("unauthorized access")
)
```

### Service Layer

Services return sentinel errors. Never return HTTP status codes
from the service layer:

```go
func (s *OrderService) Confirm(ctx context.Context, id uint64) (*Order, error) {
    order, err := s.dao.Find(ctx, id)
    if err != nil {
        return nil, err // DAO returns ErrNotFound
    }
    if order.State != OrderStateDraft {
        return nil, ErrInvalidState
    }
    // ... business logic
    return order, nil
}
```

### Handler Layer

Handlers map domain errors to HTTP responses via `errors.Is()`:

```go
func (h OrderHandler) Confirm(c fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 64)
    if err != nil {
        return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid order id.", nil)
    }

    order, err := h.service.Confirm(c.Context(), id)
    if err != nil {
        switch {
        case errors.Is(err, sales.ErrNotFound):
            return httpx.CreateNotFoundResponse(c, "Order not found.")
        case errors.Is(err, sales.ErrInvalidState):
            return httpx.CreateConflictResponse(c, "Order is not in draft state.", err)
        default:
            httpx.RequestLog(c).Error("order confirm failed", "id", id, "error", err)
            return httpx.CreateInternalServerErrorResponse(c, "Failed to confirm order.", err)
        }
    }
    return httpx.CreateSuccessResponse(c, "Order confirmed.", newOrderResponse(order))
}
```

### Error Response Helpers

Always use the `httpx` helpers. Never construct error responses
manually:

| Status | Helper |
|--------|--------|
| 400 | `httpx.CreateBadRequestResponse(c, msg, err)` |
| 401 | `httpx.CreateUnauthorizedErrorResponse(c, msg, err)` |
| 403 | `httpx.CreateForbiddenErrorResponse(c, msg, err)` |
| 404 | `httpx.CreateNotFoundResponse(c, msg)` |
| 409 | `httpx.CreateConflictResponse(c, msg, err)` |
| 422 | `httpx.CreateUnprocessableEntityErrorResponse(c, msg, fieldErrors)` |
| 500 | `httpx.CreateInternalServerErrorResponse(c, msg, err)` |
| 503 | `httpx.CreateServiceUnavailableErrorResponse(c, msg, err)` |

### Error Logging

Log errors at the handler level, not in the service layer.
Use `httpx.RequestLog(c)` for structured context:

```go
httpx.RequestLog(c).Error("operation failed", "id", id, "error", err)
```

Log warnings for expected failures (e.g., invalid credentials):

```go
httpx.RequestLog(c).Warn("login failed", "email", helper.MaskEmail(email), "reason", "invalid_credentials")
```

---

## HTTP Handlers

### Handler Struct

Handlers are methods on a struct that holds DAO/service dependencies:

```go
type ProductHandler struct {
    service products.ProductService
}

func NewProductHandler(service products.ProductService) ProductHandler {
    return ProductHandler{service: service}
}
```

### Handler Signature

Always `func (h Handler) Method(c fiber.Ctx) error`:

```go
func (h ProductHandler) List(c fiber.Ctx) error {
    // ...
}
```

### Request Binding

Always use `httpx.BindAndValidate`. It returns `false` and writes
the error response automatically:

```go
func (h ProductHandler) Create(c fiber.Ctx) error {
    var request CreateProductRequest
    if !httpx.BindAndValidate(c, &request) {
        return nil
    }
    // ... use request
}
```

### Path Parameters

Parse manually with `strconv.ParseUint`:

```go
id, err := strconv.ParseUint(c.Params("id"), 10, 64)
if err != nil {
    return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
}
```

### Response Helpers

Always use `httpx` response helpers. Never write raw `c.JSON()`:

```go
// Single resource
return httpx.CreateSuccessResponse(c, "Retrieved.", data)

// Created
return httpx.CreateCreatedResponse(c, "Created.", data)

// Paginated list
return httpx.CreateSuccessResponseWithMeta(c, "Retrieved.", data, meta)

// No content (delete)
return httpx.CreateNoContentResponse(c)
```

### Swagger Annotations

Every handler must have swaggo comments:

```go
// ListProducts godoc
// @Summary      List products
// @Description  Get a paginated list of products
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        organization_id path int true "Organization ID"
// @Param        page query int false "Page number"
// @Param        size query int false "Page size"
// @Success      200 {object} ListProductsResponse
// @Failure      401 {object} httpx.ErrorResponse
// @Router       /organizations/{organization_id}/products [get]
```

### Route Registration

Each handler registers routes in its `routes.go`:

```go
func (h ProductHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
    products := api.Group("/products", guards.AuthN)

    products.Get("/", guards.Guard("item", "view"), h.List)
    products.Post("/", guards.Guard("item", "create"), h.Create)
    products.Get("/:id", guards.Guard("item", "view"), h.Get)
    products.Put("/:id", guards.Guard("item", "update"), h.Update)
    products.Delete("/:id", guards.Guard("item", "delete"), h.Delete)
}
```

Route parameters are always `:id` (never `:item_id` or
`:itemID`). The resource name is inferred from the group.

### Idempotency

For mutation endpoints that must not be duplicated, add the
idempotency guard:

```go
orders.Post("/:id/confirm", guards.Guard("sale_order", "update"),
    httpx.IdempotencyGuard(guards.Idempotency, "sale-order-confirm"), h.Confirm)
```

---

## Database & DAOs

### Generic Base

Every domain DAO embeds `dao.Base[E]` which provides standard
CRUD operations:

```go
type UserDAO interface {
    dao.CRUD[User]
    FindByEmail(ctx context.Context, email string) (*User, error)
}

type userDAO struct {
    dao.Base[User]
    db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
    return userDAO{Base: dao.NewBase[User](db), db: db}
}
```

### Custom Queries

Domain-specific queries go in the DAO implementation:

```go
func (s userDAO) FindByEmail(ctx context.Context, email string) (*User, error) {
    return s.Search(ctx, "email", email)
}
```

### Transactions

Use `db.Transactioner` for multi-table atomic operations:

```go
type OrderService struct {
    txer db.Transactioner
    dao  OrderDAO
}

func (s *OrderService) Create(ctx context.Context, ...) (*Order, error) {
    var result *Order
    err := s.txer.Run(ctx, func(tx *gorm.DB) error {
        order, err := s.dao.Create(ctx, &order)
        if err != nil {
            return err
        }
        // ... create lines, post entries, etc.
        result = order
        return nil
    })
    return result, err
}
```

### Query Application

Use `query.ApplyQuery()` for standardized filtering, sorting,
and pagination:

```go
func (s orderDAO) List(ctx context.Context, q *query.Query) (*query.Page[Order], error) {
    return s.List(ctx, q) // inherited from dao.Base
}
```

### Model Patterns

Embed `model.Base` in all entities:

```go
type Order struct {
    model.Base
    OrganizationID uint64    `gorm:"not null" json:"organization_id"`
    Name           *string   `json:"name"`
    Date           time.Time `gorm:"type:date;not null" json:"date"`
    State          string    `gorm:"type:text" json:"state"`
    Amount         float64   `gorm:"type:numeric(18,4)" json:"amount"`
}
```

Rules:
- Nullable fields use pointers (`*uint64`, `*string`, `*time.Time`)
- GORM column types use `gorm:"type:..."` for precision (`numeric(18,4)`)
- Always include `json` tags
- Add `TableName()` method when GORM pluralization is wrong
- Use domain constants for state values (not string literals)

---

## Validation

### Request Validation

Use `go-playground/validator` struct tags:

```go
type CreateOrderRequest struct {
    OrganizationID uint64  `json:"organization_id" validate:"required"`
    Name           string  `json:"name" validate:"required"`
    Date           string  `json:"date" validate:"required"`
    Notes          *string `json:"notes"`
}
```

### Common Validation Tags

| Tag | Purpose |
|-----|---------|
| `required` | Field must be present and non-zero |
| `email` | Must be valid email format |
| `min=N` | Minimum length (string) or minimum value (int) |
| `max=N` | Maximum length or maximum value |
| `oneof=a b c` | Must be one of the allowed values |

### Binding + Validation

Always use `httpx.BindAndValidate` which handles both:

```go
if !httpx.BindAndValidate(c, &request) {
    return nil // error response already written
}
```

### Path Parameters

Manual validation with `strconv.ParseUint`:

```go
id, err := strconv.ParseUint(c.Params("id"), 10, 64)
if err != nil {
    return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
}
```

---

## Logging

### Logger

Standard library `log/slog` with optional `tint` for pretty output.
Configured via `internal/logger/`.

### Request Logging

Use `httpx.RequestLog(c)` which returns a `*slog.Logger` pre-populated
with `request_id`, `user_id`, and `organization_id`:

```go
httpx.RequestLog(c).Info("order created", "id", order.ID)
httpx.RequestLog(c).Error("payment failed", "id", paymentID, "error", err)
httpx.RequestLog(c).Warn("rate limit exceeded", "ip", c.IP())
```

### Levels

| Level | When |
|-------|------|
| `Info` | Successful operations, state transitions |
| `Warn` | Expected failures (invalid input, auth failures) |
| `Error` | Unexpected failures (DB errors, external service failures) |

### Sensitive Data

Never log passwords, tokens, or PII. Use `helper.MaskEmail()` for
email masking in logs:

```go
httpx.RequestLog(c).Warn("login failed", "email", helper.MaskEmail(email))
```

---

## Context Handling

### Actor ID (who performed the action)

```go
// Set by middleware
c.Locals(httpx.LocalUserID, user.ID)

// Read in handlers/services
actorID, ok := httpx.CallerID(c)
```

### Tenant ID (which organization)

```go
// Set by Guard middleware
c.Locals(httpx.LocalOrganizationID, orgID)

// Read in handlers
orgID, ok := httpx.CallerOrganizationID(c)
```

### Request ID (correlation)

```go
// Automatically available via requestid middleware
requestID := requestid.FromContext(c)

// Structured logging includes it automatically
httpx.RequestLog(c).Info("done") // includes request_id
```

### Passing Context

Always use `c.Context()` to pass the Fiber context to services/DAOs:

```go
order, err := h.service.Get(c.Context(), id)
```

---

## Configuration

### Loading

Config is loaded via Viper from `.env` files:

```go
cfg, err := config.New("", config.ConfigFilePath())
```

### Structure

Config fields use `mapstructure` tags:

```go
type Config struct {
    ApplicationName string `mapstructure:"APP_NAME"`
    DatabaseHost    string `mapstructure:"DB_HOST"`
    // ...
}
```

### Environment Variables

All config is overridable via environment variables. The Viper
`AutomaticEnv` + `BindEnv` pattern is used.

---

## Testing

### Unit Tests

Table-driven tests in `_test.go` files alongside source:

```go
func TestOrderService_Confirm(t *testing.T) {
    tests := []struct {
        name    string
        state   string
        wantErr error
    }{
        {"draft order confirms", OrderStateDraft, nil},
        {"confirmed order fails", OrderStateConfirmed, ErrInvalidState},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... setup mock, call service, assert
        })
    }
}
```

### Mocks

DAO mocks are in `dao_mock.go` files. Use them for service-level
unit tests:

```go
mockDAO := &MockOrderDAO{}
mockDAO.FindFunc = func(ctx context.Context, id uint64) (*Order, error) {
    return &Order{State: OrderStateDraft}, nil
}
```

### Fixtures

Use `gofakeit`-based fixtures in `fixtures.go`:

```go
func OrderFixture(opts ...func(*Order) *Order) *Order {
    o := &Order{
        Base:  model.Base{ID: gofakeit.Uint64()},
        State: OrderStateDraft,
    }
    for _, opt := range opts {
        opt(o)
    }
    return o
}
```

### Integration Tests

In `test/integration/` with build tag `//go:build integration`:

```go
//go:build integration

func TestOrderFlow(t *testing.T) {
    db := testutil.SetupTestDB()
    defer testutil.CleanTables(t, db)
    // ... test with real database
}
```

### E2E Tests

In `test/e2e/` with build tag `//go:build e2e`:

```go
//go:build e2e

func TestCreateOrder(t *testing.T) {
    e := httpexpect.New(t, baseURL)
    e.POST("/api/v1/organizations/{id}/orders").
        WithPath("id", orgID).
        WithJSON(payload).
        Expect().Status(http.StatusCreated)
}
```

---

## Queue & Background Jobs

### Task Definitions

Define task types in `internal/queue/tasks/`:

```go
const TypeSendEmail = "send-email"

type SendEmailPayload struct {
    UserID uint64 `json:"user_id"`
    Token  string `json:"token"`
}
```

### Task Handlers

Implement `func(ctx context.Context, t *asynq.Task) error` in
`internal/queue/handlers/`:

```go
func SendEmailHandler(ctx context.Context, t *asynq.Task) error {
    var payload tasks.SendEmailPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return err
    }
    // ... process task
    return nil
}
```

### Enqueuing Tasks

Use the `TaskEnqueuer` interface:

```go
payload, _ := json.Marshal(tasks.SendEmailPayload{UserID: user.ID})
_, err := s.enqueuer.EnqueueContext(ctx, tasks.TypeSendEmail, payload,
    asynq.MaxRetry(5),
    asynq.Timeout(30*time.Second),
)
```

---

## Code Style

### Formatting

Standard `gofmt` + `goimports`. Tabs for indentation.

### Naming

| Element | Convention | Example |
|---------|-----------|---------|
| Packages | `lowercase`, no underscores | `products`, `httpx` |
| Exported types | `PascalCase` | `ProductHandler`, `OrderService` |
| Unexported types | `camelCase` | `userDAO`, `requestValidator` |
| Exported functions | `PascalCase` | `NewProductHandler` |
| Unexported functions | `camelCase` | `writeProductError` |
| Constants | `PascalCase` (exported) or `camelCase` (unexported) | `OrderStateDraft`, `defaultPageSize` |
| Interfaces | `PascalCase`, often verb-based | `CRUD[E]`, `TaskEnqueuer` |
| Methods | `PascalCase` (exported) or `camelCase` (unexported) | `h.List`, `writeAccessErrorResponse` |

### Export Rules

- Export everything consumed by other packages
- Keep helpers unexported within their package
- Types used in JSON/OpenAPI specs are always exported

### Struct Tags

Order: `json`, `validate`, `gorm`, `mapstructure`:

```go
type Item struct {
    model.Base
    Name   string  `json:"name" validate:"required" gorm:"not null"`
    Price  float64 `json:"price" gorm:"type:numeric(18,4)"`
}
```

### Pointer vs Value

- Use pointers for nullable database columns (`*uint64`, `*string`)
- Use values for required fields
- Use `time.Time` (not pointer) for required timestamps
- Use `*time.Time` for optional timestamps (e.g., `DeletedAt`)

### Comments

Keep comments minimal. Code should be self-documenting. Add doc
comments only on exported types and functions:

```go
// OrderService handles order lifecycle operations.
type OrderService struct { ... }

// Confirm transitions a draft order to confirmed state.
func (s *OrderService) Confirm(ctx context.Context, id uint64) (*Order, error) { ... }
```

### Linting

Enforced via `.golangci.yaml`:
- `errcheck` — unchecked errors
- `govet` — suspicious constructs
- `ineffassign` — dead assignments
- `staticcheck` — static analysis
- `unused` — dead code
- `gocritic` — opinionated checks
- `misspell` — typos
- `revive` — style enforcement

Run with `just lint` or `golangci-lint run`.
