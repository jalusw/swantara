# Swantara Web

Web client for **Swantara** — a generic open full-cycle ERP platform for small to medium businesses.

Next.js 16, React 19, Tailwind 4, shadcn + Base UI, TanStack Query/Table, Zustand, React Hook Form + Zod, next-intl, Serwist PWA for low-end and offline use. Feature-oriented App Router structure with container/presentation split and co-located tests.

Item overview, setup, and license live in `../README.md`. This file is the single source of truth for frontend architecture and coding conventions.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Framework | Next.js 16 (App Router) |
| UI | React 19, Tailwind 4, shadcn + Base UI |
| Data | TanStack Query/Table, Zustand, Axios |
| Forms | React Hook Form + Zod |
| i18n | next-intl |
| PWA | Serwist (offline, low-end devices) |
| Test | Vitest + Testing Library + MSW, Playwright |
| Lint | Biome, tsc |

## Layout

```
app/                  Next.js App Router pages and layouts
components/           Shared UI (no barrel files)
lib/                  client, server, services, queries, hooks, utils
providers/            React context providers
stores/               Zustand stores (*.store.ts)
types/                Global type augmentations (*.d.ts)
styles/               CSS entry points and tokens
e2e/                  Playwright suites
```

## Setup

```bash
cp .env.example .env
pnpm install
pnpm dev
```

Web at `http://localhost:3000`, API base `http://localhost:8080/api/v1`. Run `pnpm lint`, `pnpm typecheck`, and `pnpm test` before committing.

---
# Architecture
Patterns for organizing code in this codebase. Prioritize simplicity,
clarity, and elegance. Future agents: follow these patterns unless there's
a good reason not to.

---

## Core Principles

1. **Simple over clever** — Code should be obvious at a glance
2. **Consistent over optimal** — Follow existing patterns, even if slightly longer
3. **Readable over terse** — Prefer clarity over conciseness
4. **Colocate related code** — Keep things close to where they're used
5. **One responsibility per file** — Each file does one thing well

---

## Feature Module Structure

Every feature follows this structure:

```
feature/
├── _components/          # UI components
│   ├── feature-section.tsx      # Main section/container
│   ├── feature-utils.ts         # Helper functions
│   ├── feature-form-dialog.tsx  # Form dialogs
│   ├── __tests__/               # Co-located tests
│   └── ...
├── _hooks/               # Custom hooks (optional, only if reused)
├── _types.ts             # Feature-specific types (optional)
└── page.tsx              # Route page
```

### Rules

- `_components/` contains ALL UI for this feature
- `_utils.ts` or `*-utils.ts` contains helpers, formatters, constants
- `_hooks/` only for hooks used by multiple components in this feature
- `_types.ts` only for types shared across multiple files in this feature
- Tests go in `__tests__/` inside `_components/`

### Example

```
sales/
├── _components/
│   ├── sales-section.tsx           # Main container
│   ├── sales-utils.ts              # Formatters, helpers
│   ├── sale-order-form-dialog.tsx  # Form
│   ├── sale-order-detail.tsx       # Detail view
│   └── __tests__/
│       └── sales-utils.test.ts
├── _hooks/
│   └── use-sale-orders.ts          # Query hook (if shared)
└── page.tsx
```

---

## Container/Presentation Pattern

Split components into two types:

### Container Components (Smart)

- Handles data fetching, state, business logic
- Named with `*-section.tsx` or `*-container.tsx`
- Lives in `feature/_components/`
- Contains `useQuery`, `useState`, event handlers
- Passes data and callbacks to presentation component

### Presentation Components (Dumb)

- Pure UI, receives data via props
- Named with descriptive name (e.g., `sales-table.tsx`, `sale-card.tsx`)
- Lives in `components/` (shared) or `feature/_components/` (feature-specific)
- No data fetching, minimal state
- Easy to test, reuse, and reason about

### File Naming

```
feature/_components/
├── sales-section.tsx          # Container (data + logic)
├── sales-table.tsx            # Presentation (UI only)
├── sales-filters.tsx          # Presentation (UI only)
└── sales-utils.ts             # Helpers
```

### Container Pattern

```tsx
// sales-section.tsx (Container)
export function SalesSection({ orgId }: { orgId: number }) {
  const { data, isLoading } = useSalesQuery(orgId)
  const [filters, setFilters] = useState<SalesFilters>({})

  if (isLoading) return <SalesSectionSkeleton />
  if (!data) return <SalesSectionEmpty />

  return (
    <SalesTable
      orders={data.orders}
      filters={filters}
      onFilterChange={setFilters}
      onOrderClick={handleOrderClick}
    />
  )
}
```

### Presentation Pattern

```tsx
// sales-table.tsx (Presentation)
type SalesTableProps = {
  orders: SaleOrder[]
  filters: SalesFilters
  onFilterChange: (filters: SalesFilters) => void
  onOrderClick: (order: SaleOrder) => void
}

export function SalesTable({ orders, filters, onFilterChange, onOrderClick }: SalesTableProps) {
  return (
    <Table>
      {/* Pure UI, no data fetching */}
    </Table>
  )
}
```

---

## Component Organization

All shared components live in `components/`:

```
components/
├── ui/                    # Primitives (shadcn/ui style)
│   ├── button.tsx
│   ├── input.tsx
│   └── ...
├── data-table.tsx         # Business components
├── combobox.tsx
├── pdf-viewer.tsx
└── ...
```

### Rules

- `components/ui/` — Primitives, no business logic
- `components/` — Business components, reusable across features
- Feature-specific components stay in `feature/_components/`
- Container components never go in `components/` (they belong to features)

---

## Form Pattern

### Schema Location

```typescript
// feature/_components/feature-utils.ts
export const featureFormSchema = z.object({
  name: z.string().min(1),
  email: z.string().email(),
  // ...
})

export type FeatureFormValues = z.infer<typeof featureFormSchema>
```

### Form Component

```tsx
// feature/_components/feature-form-dialog.tsx
import { featureFormSchema, type FeatureFormValues } from './feature-utils'

type FeatureFormDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialData?: Feature
  onSave: (data: FeatureFormValues) => void
}

export function FeatureFormDialog({ open, onOpenChange, initialData, onSave }: FeatureFormDialogProps) {
  const form = useForm<FeatureFormValues>({
    resolver: zodResolver(featureFormSchema),
    defaultValues: initialData ?? { name: '', email: '' },
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <Form {...form}>
          {/* Form fields */}
        </Form>
      </DialogContent>
    </Dialog>
  )
}
```

---

## Data Fetching Pattern

### Query Hook

```typescript
// feature/_hooks/use-feature-query.ts
const featureQueryKey = ['features'] as const

export function featureListQueryKey(orgId: number) {
  return [...featureQueryKey, 'list', orgId] as const
}

export function featureDetailQueryKey(orgId: number, id: number) {
  return [...featureQueryKey, 'detail', orgId, id] as const
}

export function useFeatureQuery(orgId: number, id: number) {
  return useQuery({
    queryKey: featureDetailQueryKey(orgId, id),
    queryFn: () => api.getFeature(orgId, id),
  })
}
```

### Mutation Pattern

```typescript
// feature/_components/feature-section.tsx
const updateMutation = useOrgMutation({
  endpoint: "/api/features",
  method: "PUT",
  onSuccess: () => {
    toast.success("Saved.");
  },
  onError: (error) => {
    toast.error("Something went wrong. Please try again.");
  },
})
```

Always use `useOrgMutation`. Never call
`queryClient.invalidateQueries` directly in component bodies.
Invalidate both list and detail keys in `onSuccess`. Always provide
`onError` with user feedback.

---

## Loading & Empty States

Every section uses `SectionWrapper` or the DataTable `status` prop.
Never return `null` while loading. Never render inline empty text.

```tsx
export function FeatureSection({ orgId }: { orgId: number }) {
  const { data, isLoading, isError, refetch, error } = useFeatureQuery(orgId)

  return (
    <SectionWrapper
      isLoading={isLoading}
      isError={isError}
      isEmpty={!data || data.length === 0}
      error={error?.message}
      onRetry={() => void refetch()}
    >
      <FeatureTable data={data ?? []} />
    </SectionWrapper>
  )
}
```

List views map query state to DataTable `status` with the shared
status helper. Detail views use `DetailPageSkeleton`. Route groups
with heavy fetching add `loading.tsx` and `error.tsx` boundaries.

### Skeleton Pattern

```tsx
function FeatureSectionSkeleton() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-64 w-full" />
    </div>
  )
}
```

### Empty Pattern

```tsx
function FeatureSectionEmpty() {
  return (
    <EmptyState
      icon={<FeatureIcon className="size-12" />}
      title="No items yet"
      description="Get started by creating your first record."
      action={<Button onClick={handleCreate}>Create</Button>}
    />
  )
}
```

---

## Type Patterns

### Props Types

```typescript
// Always use type, not interface
type FeatureCardProps = {
  feature: Feature
  onSelect: (id: number) => void
}

// Destructure in function signature
export function FeatureCard({ feature, onSelect }: FeatureCardProps) {
  // ...
}
```

### Component Props Location

```typescript
// Option 1: At top of file (preferred for small components)
type FeatureCardProps = { /* ... */ }

export function FeatureCard({ /* ... */ }: FeatureCardProps) { /* ... */ }

// Option 2: In _types.ts (only if shared across multiple files)
```

---

## Error Handling

### Server Layer

```typescript
// lib/server/handler.ts
export function handler(fn: (req: NextRequest) => Promise<Response>) {
  return async (req: NextRequest) => {
    try {
      return await fn(req)
    } catch (error) {
      console.error(error)
      return createErrorResponse(500, 'Internal server error')
    }
  }
}
```

### Client Layer

```typescript
// lib/client/error.ts
export function isServerError(error: unknown): error is AxiosError {
  return axios.isAxiosError(error) && error.response?.status !== undefined
}
```

### Component Layer

```tsx
// Use ErrorBoundary for unexpected errors
// Use try/catch for expected errors (form validation, API calls)
// Use error state for query failures
```

---

## File Naming Rules

| Pattern | Example | Use Case |
|---------|---------|----------|
| `kebab-case.tsx` | `feature-section.tsx` | Components |
| `kebab-case.ts` | `feature-utils.ts` | Utilities |
| `*.store.ts` | `auth.store.ts` | Zustand stores |
| `*.test.ts` | `feature.test.ts` | Tests |
| `use-*.ts` | `use-feature-query.ts` | Custom hooks |
| `*-utils.ts` | `feature-utils.ts` | Feature helpers |
| `*-section.tsx` | `feature-section.tsx` | Container components |

---

## Export Rules

### Components

```typescript
// Named export (preferred)
export function FeatureCard({ /* ... */ }: FeatureCardProps) { /* ... */ }

// Default export only for Next.js boundaries (requirement)
export default function FeaturePage() { /* ... */ }
```

Only `page.tsx`, `layout.tsx`, `loading.tsx`, `error.tsx`,
`not-found.tsx`, and `template.tsx` may use `export default`.
Components, hooks, stores, providers, and all `lib/` modules use
named exports. `lib/i18n/request.ts` keeps its default per next-intl
convention.

### Utilities

```typescript
// Named export
export function formatDate(date: Date): string { /* ... */ }
export function formatCurrency(amount: number): string { /* ... */ }
```

### Types

```typescript
// Named export
export type Feature = { id: number; name: string }
export type FeatureFormValues = z.infer<typeof featureSchema>
```

---

## Import Rules

```typescript
// 1. React/Next.js imports
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'

// 2. External libraries
import { useQuery } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'

// 3. Internal utilities
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/utils/formatters'

// 4. Shared components
import { Button } from '@/components/ui/button'
import { DataTable } from '@/components/data-table'

// 5. Feature-specific components
import { FeatureFormDialog } from './feature-form-dialog'
```

Use `@/*` alias only. Never `@lib/*`. Use named React imports,
never `import * as React`. Import shared factories (`createId`,
table column helpers, `formatHours`) from `lib/utils/`, never
duplicate them per feature.

---

## Composition Patterns

Build complex UIs from simple, composable pieces. When facing similar
(not identical) situations, compose existing patterns instead of rewriting.

---

### Compound Components

Build components that work together via shared context. Users compose
them in JSX, not via complex props.

**Pattern:**

```typescript
// components/data-table.tsx
type DataTableContextValue = {
  table: TanStackTable<unknown>
}

const DataTableContext = createContext<DataTableContextValue | null>(null)

function useDataTable() {
  const ctx = useContext(DataTableContext)
  if (!ctx) throw new Error('DataTable components must be used within <DataTable>')
  return ctx
}

// Root provider
type DataTableProps<TData> = {
  table: TanStackTable<TData>
  children: ReactNode
}

export function DataTable<TData>({ table, children }: DataTableProps<TData>) {
  return (
    <DataTableContext.Provider value={{ table }}>
      <div className="relative overflow-auto">{children}</div>
    </DataTableContext.Provider>
  )
}

// Sub-components
export function DataTableHeader() {
  const { table } = useDataTable()
  // Renders table headers from table.getHeaderGroups()
}

export function DataTableBody() {
  const { table } = useDataTable()
  // Renders rows from table.getRowModel().rows
}

export function DataTableRow<TData>({ row }: { row: TanStackRow<TData> }) {
  // Individual row rendering
}
```

**Usage (composable):**

```tsx
<DataTable table={table}>
  <DataTableHeader />
  <DataTableBody />
</DataTable>

// Or customize:
<DataTable table={table}>
  <DataTableHeader />
  <tbody>
    {table.getRowModel().rows.map(row => (
      <MyCustomRow key={row.id} row={row} />
    ))}
  </tbody>
</DataTable>
```

**Benefits:**
- Users control markup structure
- Easy to add/remove parts
- Shared state via context (no prop drilling)
- Each sub-component is independently testable

---

### Headless Hooks

Extract state logic into hooks that return data + actions. Components
handle only rendering. This separates concerns and enables reuse across
different UIs.

**Pattern:**

```typescript
// lib/hooks/use-data-table.ts
type UseDataTableOptions<TData> = {
  data: TData[]
  columns: ColumnDef<TData>[]
  filters?: FilterConfig[]
}

type UseDataTableReturn<TData> = {
  table: TanStackTable<TData>
  filters: FilterState
  setFilter: (key: string, value: FilterValue) => void
  resetFilters: () => void
  selectedRows: TData[]
  globalFilter: string
  setGlobalFilter: (value: string) => void
}

export function useDataTable<TData>(options: UseDataTableOptions<TData>): UseDataTableReturn<TData> {
  const [filters, setFilters] = useState<FilterState>({})
  const [globalFilter, setGlobalFilter] = useState('')

  const table = useReactTable({
    data: options.data,
    columns: options.columns,
    state: { filters, globalFilter },
    // ... table config
  })

  return {
    table,
    filters,
    setFilter: (key, value) => setFilters(prev => ({ ...prev, [key]: value })),
    resetFilters: () => setFilters({}),
    selectedRows: table.getSelectedRowModel().rows.map(r => r.original),
    globalFilter,
    setGlobalFilter,
  }
}
```

**Usage (compose with any UI):**

```tsx
// With DataTable component
function SalesSection() {
  const { data } = useSalesQuery()
  const { table, filters, setFilter, resetFilters } = useDataTable({
    data,
    columns: salesColumns,
  })

  return (
    <div>
      <SalesFilters filters={filters} onFilterChange={setFilter} onReset={resetFilters} />
      <DataTable table={table}>
        <DataTableHeader />
        <DataTableBody />
      </DataTable>
    </div>
  )
}

// With custom UI (same hook, different rendering)
function SalesCards() {
  const { data } = useSalesQuery()
  const { table, filters, setFilter } = useDataTable({ data, columns: salesColumns })

  return (
    <div>
      <SalesFilters filters={filters} onFilterChange={setFilter} />
      <div className="grid grid-cols-3 gap-4">
        {table.getRowModel().rows.map(row => (
          <SalesCard key={row.id} row={row} />
        ))}
      </div>
    </div>
  )
}
```

---

### Section Wrappers

Reusable patterns for loading, empty, and error states. Every feature
section follows this composition.

**Pattern:**

```typescript
// components/section-wrapper.tsx
type SectionWrapperProps = {
  isLoading?: boolean
  isError?: boolean
  isEmpty?: boolean
  skeleton?: ReactNode
  empty?: ReactNode
  error?: ReactNode
  children: ReactNode
}

export function SectionWrapper({ isLoading, isError, isEmpty, skeleton, empty, error, children }: SectionWrapperProps) {
  if (isLoading) return skeleton ?? <SectionSkeleton />
  if (isError) return error ?? <SectionError />
  if (isEmpty) return empty ?? <SectionEmpty />
  return <>{children}</>
}

// Pre-built variants
export function SectionSkeleton() {
  return (
    <div className="space-y-4">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-64 w-full" />
    </div>
  )
}

export function SectionEmpty({ title, description, action }: { title: string; description?: string; action?: ReactNode }) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <p className="text-lg font-medium">{title}</p>
      {description && <p className="text-muted-foreground">{description}</p>}
      {action}
    </div>
  )
}

export function SectionError({ message }: { message?: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center text-destructive">
      <p className="text-lg font-medium">Something went wrong</p>
      {message && <p className="text-muted-foreground">{message}</p>}
    </div>
  )
}
```

**Usage:**

```tsx
function FeaturesSection() {
  const { data, isLoading, isError } = useFeaturesQuery()

  return (
    <SectionWrapper
      isLoading={isLoading}
      isError={isError}
      isEmpty={!data || data.length === 0}
    >
      <FeaturesTable data={data} />
    </SectionWrapper>
  )
}
```

---

### CRUD Pattern

Standardize create, read, update, delete operations across features.

**Pattern:**

```typescript
// feature/_components/feature-section.tsx
export function FeaturesSection({ orgId }: { orgId: number }) {
  const [dialogOpen, setDialogOpen] = useState(false)
  const [selectedFeature, setSelectedFeature] = useState<Feature | null>(null)

  const { data, isLoading, isError } = useFeaturesQuery(orgId)
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: (data: FeatureFormValues) => api.createFeature(orgId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: featureListQueryKey(orgId) })
      setDialogOpen(false)
    },
  })

  const updateMutation = useMutation({
    mutationFn: (data: FeatureFormValues) => api.updateFeature(orgId, selectedFeature!.id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: featureListQueryKey(orgId) })
      setDialogOpen(false)
      setSelectedFeature(null)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => api.deleteFeature(orgId, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: featureListQueryKey(orgId) })
    },
  })

  const handleEdit = (feature: Feature) => {
    setSelectedFeature(feature)
    setDialogOpen(true)
  }

  const handleCreate = () => {
    setSelectedFeature(null)
    setDialogOpen(true)
  }

  return (
    <SectionWrapper isLoading={isLoading} isError={isError} isEmpty={!data?.length}>
      <div className="flex items-center justify-between">
        <h2>Features</h2>
        <Button onClick={handleCreate}>Add Feature</Button>
      </div>

      <FeaturesTable
        features={data}
        onEdit={handleEdit}
        onDelete={(id) => deleteMutation.mutate(id)}
      />

      <FeatureFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        initialData={selectedFeature}
        onSave={(data) => selectedFeature ? updateMutation.mutate(data) : createMutation.mutate(data)}
        isPending={createMutation.isPending || updateMutation.isPending}
      />
    </SectionWrapper>
  )
}
```

---

### Composable Form Pattern

Forms compose validation, layout, and submission. Each part is
independently reusable.

**Pattern:**

```typescript
// feature/_components/feature-form-dialog.tsx
type FeatureFormDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialData?: Feature
  onSave: (data: FeatureFormValues) => void
  isPending?: boolean
}

export function FeatureFormDialog({ open, onOpenChange, initialData, onSave, isPending }: FeatureFormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{initialData ? 'Edit' : 'Create'} Feature</DialogTitle>
        </DialogHeader>
        <FeatureForm initialData={initialData} onSubmit={onSave} isPending={isPending} />
      </DialogContent>
    </Dialog>
  )
}

// Reusable form (can be used standalone or in dialog)
type FeatureFormProps = {
  initialData?: Feature
  onSubmit: (data: FeatureFormValues) => void
  isPending?: boolean
}

export function FeatureForm({ initialData, onSubmit, isPending }: FeatureFormProps) {
  const form = useForm<FeatureFormValues>({
    resolver: zodResolver(featureFormSchema),
    defaultValues: initialData ?? { name: '', description: '' },
  })

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField control={form.control} name="name" render={({ field }) => (
          <FormItem>
            <FormLabel>Name</FormLabel>
            <FormControl><Input {...field} /></FormControl>
            <FormMessage />
          </FormItem>
        )} />
        {/* More fields */}
        <Button type="submit" disabled={isPending}>
          {isPending ? 'Saving...' : 'Save'}
        </Button>
      </form>
    </Form>
  )
}
```

---

### Composition Rules

1. **Keep primitives simple** — UI components (Button, Input) should be
   thin wrappers, not complex logic holders

2. **Context for shared state** — When multiple sub-components need the
   same data, use context (not prop drilling)

3. **Hooks for logic** — Extract stateful logic into hooks, keep components
   focused on rendering

4. **Render when needed** — Only use render props/callbacks when the parent
   needs to customize rendering. Don't over-abstract

5. **Compose, don't inherit** — Build complex components from simple ones,
   don't extend or wrap deeply

6. **Defaults with escape hatches** — Provide sensible defaults, allow
   overrides via props or composition

---

## Enforcement Checklist

Future agents: verify these before committing:

- [ ] No `export default` except Next.js boundaries (`page/layout/loading/error/not-found/template`) and `lib/i18n/request.ts`
- [ ] No `export default` in `lib/`, `providers/`, `stores/`, components, or hooks
- [ ] No `interface` for props (use `type`); no arrow components; no `import * as React`
- [ ] No local `formatDate`/`formatCurrency` copies; no `toLocale*` without locale; no hardcoded `en-US`/`USD`; no `$` prefix
- [ ] No hardcoded English UI, Badge, aria, defaults, or label functions (use `t()`); no `humanizeKey` for translated strings
- [ ] Container components named `*-section.tsx`; never in shared `components/`
- [ ] Presentation components have no data fetching; hooks before early returns
- [ ] Forms use module-scope schema from `*-utils.ts` with `.trim()`/`.refine()`; buttons disabled on `isPending`; state reset on reopen
- [ ] Queries use centralized keys with `orgId`; `useOrgQuery` for single, `useOrgListQuery` for lists; `enabled` guards; numbers validated; conversion at page boundary
- [ ] Mutations use `useOrgMutation` with list+detail invalidation and `onError` toast; no `refetch()` instead of invalidation; no `.sort()` on cache; no service bypass; no missing `.catch()`
- [ ] Every section has loading/empty/error states (use SectionWrapper/DataTable `status`); route groups have `loading.tsx`/`error.tsx`; never return `null` on load
- [ ] `cn()` for classes, no template classNames, no inline styles in loops, no IIFE in JSX, stable keys, no nested ternaries, no `any`/`as never`/`as unknown`
- [ ] No `console.*`; no unused imports/exports; no hardcoded URLs/emails/timeouts; `EMPTY_FALLBACK` for empty display; clean up effects/blob URLs
- [ ] CSRF on mutating routes; proxy input validated; no secret/URL logging; `Secure` cookies in prod; single refresh authority
- [ ] A11y: `lang`, `id="main"`, `aria-describedby`, 44px targets; Metadata: `generateMetadata` with `openGraph`/`alternates`/`robots`; `<Suspense>` in layouts
- [ ] Tests co-located in `__tests__/` inside `_components/`
- [ ] DataTable uses compound component pattern
- [ ] Headless hooks for state logic, components for rendering only
- [ ] CRUD operations follow standard pattern (create/update/delete mutations)
- [ ] Reuse canonical `EntityFormDialog`, `EntitySection`, `ActiveBadge`, `JournalDateDialog`, `DetailPageSkeleton`, `createId`, table helpers; no new duplicates

---

## Anti-Patterns to Avoid

1. **God components** — Components doing too much (>150 lines, >5 state vars)
2. **Prop drilling** — Passing data through 3+ levels (use context or hooks)
3. **Inline logic** — Business logic in JSX (extract to utils/hooks)
4. **Magic strings/numbers** — Extract to constants
5. **Copy-paste code** — Extract shared patterns to utils
6. **Premature abstraction** — Don't abstract until you have 3+ uses
7. **Clever one-liners** — Clarity over conciseness
8. **Dead code** — Unused components, hooks, stores, utils, or CSS shipped in bundle. Remove it.
9. **Bypass layers** — Direct `fetch`/service calls skipping `SwantaraService`/`useOrgMutation`, or `useState`+service where React Query applies.

---

# Conventions
Single source of truth for coding patterns in `web/`. All new code and
refactors must follow these conventions. For architecture patterns
(component organization, data fetching, container/presentation), see
the Architecture section above.

## File & Folder Naming

| What | Convention | Example |
|------|------------|---------|
| All source files | `kebab-case` | `empty-state.tsx`, `use-login-form.ts` |
| Components | `kebab-case.tsx` | `confirm-dialog.tsx` |
| Hooks | `use-kebab-case.ts` | `use-data-table.ts` |
| Stores | `kebab-case.store.ts` | `theme.store.ts` |
| Tests | `*.test.ts(x)` co-located in `__tests__/` | `components/__tests__/button.test.tsx` |
| CSS | `kebab-case.css` | `tokens.css`, `globals.css` |
| Folders | `kebab-case` | `lib/utils/`, `lib/hooks/` |
| Private folders | `_` prefix | `_components/`, `_hooks/` |
| Route groups | `(kebab-case)` | `(auth)`, `(org)` |

## Component Patterns

### Function Style

Use **function declarations** for components. Never arrow functions.

```tsx
// Good
function Button({ ... }: ButtonProps) { ... }

// Bad
const Button = ({ ... }: ButtonProps) => { ... }
```

### Props Typing

Use the `type` keyword (not `interface`).

**Primitive wrappers** — extend `React.ComponentProps`:
```tsx
function Card({ className, ...props }: React.ComponentProps<"div">) { ... }
```

**Composite / domain components** — separate named type:
```tsx
export type ConfirmDialogProps = {
  title: string;
  onConfirm: () => void;
  children: React.ReactNode;
};

function ConfirmDialog({ title, onConfirm, children }: ConfirmDialogProps) { ... }
```

### Exports

**Named exports only.** No `export default` for components, hooks,
stores, or providers.

```tsx
// Good
export function Button({ ... }: ButtonProps) { ... }

// Bad
export default function Button({ ... }: ButtonProps) { ... }
```

Exception: Next.js boundaries (`page.tsx`, `layout.tsx`, `loading.tsx`,
`error.tsx`, `not-found.tsx`, `template.tsx`) and
`lib/i18n/request.ts` per next-intl convention.

### Client / Server

Place `"use client"` only when the file uses React hooks, event handlers, or
browser APIs. Prefer server components when possible.

```tsx
// First line of the file, before any imports
"use client";
```

**Do NOT add `"use client"` when:**
- Wrapping a primitive component that already marks itself as client
  (e.g., `Separator`, `Label`, `Tooltip`, `ScrollArea`, `Progress`)
- The file contains only pure utility functions with no hooks or DOM access
- The file is a `.ts` module (not `.tsx`) with no client-side dependencies

### data-slot

Every UI component root element gets a `data-slot` attribute using kebab-case
matching the component name:

```tsx
function Card({ className, ...props }: React.ComponentProps<"div">) {
  return <div data-slot="card" className={cn("...", className)} {...props} />;
}

function DataTable({ ... }: DataTableProps) {
  return <div data-slot="data-table" ...>;
}
```

### Refs

Use the React 19 ref-as-prop pattern. Never use `React.forwardRef`.

```tsx
// Good
function Button({ ref, ...props }: ButtonProps & { ref?: React.Ref<HTMLButtonElement> }) {
  return <button ref={ref} {...props} />;
}

// Bad
const Button = React.forwardRef<HTMLButtonElement, ButtonProps>((props, ref) => {
  return <button ref={ref} {...props} />;
});
```

### Props Conventions

**Controlled dialog state** — use `open` + `onOpenChange` (Radix pattern):

```tsx
type ConfirmDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
};
```

**Close handlers** — use semantically appropriate names:

| Pattern | Use for |
|---------|---------|
| `onOpenChange` | Controlled open/close state on overlays |
| `onClose` | Explicit close action (e.g., close a panel) |
| `onDismiss` | Dismiss informational content (e.g., callout, banner) |
| `onCancel` | Abort an in-progress operation |

### Touch Targets

All interactive components (buttons, links, form inputs, clickable elements)
must have a minimum touch target size of 44×44px. Use padding inside elements
to meet this requirement rather than relying on visual size alone.

```tsx
// Good — padding ensures 44px touch target
<button className="p-3 min-h-[44px] min-w-[44px]">Click</button>

// Bad — too small for touch
<button className="p-1 text-sm">Click</button>
```

### Component Size

Keep page components under 150 lines. If a page exceeds this, extract
sub-components, data fetching hooks, and mock data into `_components/` or
`_utils.ts`.

```tsx
// Good — page is thin, logic lives in sub-components
export default function AccountingPage() {
  return (
    <div>
      <AccountingSummary />
      <AccountingTable />
    </div>
  );
}

// Bad — page is 300+ lines with inline data and logic
export default function AccountingPage() {
  const data = [...]; // 50 lines of mock data
  return (
    <div>
      {/* 250 lines of JSX */}
    </div>
  );
}
```

Extract shared utilities (formatters, mappers, mock data) into `_utils.ts`
alongside the feature. Keep page files focused on composition.

## Styling

- **Tailwind CSS v4** — utility classes only.
- **`cn()`** — merge class names (`clsx` + `tailwind-merge`). Import from `@/lib/utils`. Never use template literal `className={`... ${cond ? ...}`}` for conditional classes.
- **CVA** — component variants via `class-variance-authority`.
- No inline styles. No `style=` props except dynamic data-driven values (width/height from data). Never put style objects inside render loops; hoist static styles or use CSS variables.
- No IIFE in JSX. Extract computations to variables or sub-components.

## Directory Structure

```
app/                  # Next.js App Router pages and layouts
components/           # Shared UI components (no barrel files)
lib/
  client/             # Client-side utilities (axios, logger)
  constants/          # App constants (UPPER_SNAKE_CASE)
  hooks/              # Shared custom hooks
  middleware/         # Server middleware helpers
  queries/            # TanStack Query query option factories
  server/             # Server-side utilities (API, CSRF, session)
  services/           # API service layer
  tests/              # Test utilities, mocks, handlers
  utils/              # General utilities
providers/            # React context providers
stores/               # Zustand stores (*.store.ts)
types/                # Global type augmentations only (*.d.ts)
styles/               # CSS entry points and tokens
```

## Imports

Use `@/*` alias only. Never `@lib/*`.

```tsx
// Good
import { cn } from "@/lib/utils";
import { Button } from "@/components/button";

// Bad
import { cn } from "@lib/utils";
```

## Constants

`UPPER_SNAKE_CASE` for all constants. Co-locate in `lib/constants/` or module file.

```tsx
export const STORAGE_KEY = "swantara:theme" as const;
export const SHOW_DELAY_MS = 500;
export const EMPTY_FALLBACK = "—" as const;
```

Use `EMPTY_FALLBACK` for all null/empty display fallbacks, never
inline `"–"`/`"—"`. Extract timeout durations, support email, map tile
URL, and hardcoded API URLs to constants or env. Never inline them.

### UI Copy

All user-visible strings use `next-intl` `t()`. This includes component
text, `Badge` content, `aria-label`/`title`, default prop values,
`defaultLabels` objects, and status label functions in `*-utils.ts`.
Never hardcode English fallbacks that silently defeat i18n.

```tsx
// Good
const t = useTranslations("sales");
return <p>{t("title")}</p>;
<span>{t(`status.${order.state}`)}</span>

// Bad
return <p>Dashboard</p>;
<span>{humanizeKey(status)}</span>
```

`humanizeKey` from `@/lib/utils/case` is only for dynamic enum-like
values with no translation key yet. Support email, map tile URL, and
other environment-specific strings live in constants or env, never
inline.

## Testing

- **Unit**: Vitest + Testing Library + MSW. Co-located in `__tests__/`.
- **E2E**: Playwright. Tests in `e2e/`.
- Test files focus on behavior and expected outcomes. Stubs and setup extracted
  to separate files. Reuse `renderWithProviders` and shared utilities from
  `lib/tests/` instead of duplicating setup. Assert meaningful behavior,
  never only `toBeDefined()` / `not.toBeNull()`.

## Formatting

Enforced by Biome (see `biome.json`):

- 2-space indent, spaces (not tabs)
- Double quotes
- Semicolons always
- Trailing commas (all)
- Line width: 100
- LF line endings

## Error Handling

### Route-Level

Data-heavy route groups add all three boundaries:

- `loading.tsx` — skeleton UI (never blank screen)
- `error.tsx` — error UI (must be client component)
- `not-found.tsx` — 404 UI

Use the existing `ErrorState` component from `@/components/error-state` for
consistent error display across routes. Wrap server layout prefetch
(`prefetchPermissionsData`, `prefetchMeData`) in try/catch so transient
API errors do not crash the section.

### Service Layer

Use the typed error hierarchy from `@/lib/services/swantara/errors`:

```ts
try {
  await service.catalog.products.list(orgId);
} catch (error) {
  if (error instanceof SwantaraUnprocessableError) {
    // Handle validation errors — error.fieldErrors contains details
  } else if (error instanceof SwantaraNotFoundError) {
    // Handle not found
  }
}
```

Never swallow errors silently. At minimum, log them or show user feedback.

## Utility Functions

**Single source of truth.** All formatting utilities live in `@/lib/utils/formatters`.
Never redefine `formatDate`, `formatCurrency`, `formatNumber`, etc. locally.
Never call `toLocaleDateString`, `toLocaleTimeString`, or `toLocaleString`
without a locale argument. Never hardcode `"en-US"` or `"USD"`.

```ts
import { formatDate, formatCurrency } from "@/lib/utils/formatters";

// Good — uses centralized formatter with flexible options
formatDate(date, { nullFallback: "—" });
formatCurrency(amount, { currency: "IDR" });

// Bad — local redefinition with hardcoded values
function formatDate(date: string | null) {
  return date ? new Intl.DateTimeFormat("en-US", { ... }).format(new Date(date)) : "—";
}
```

When a feature needs different formatting behavior, extend the centralized
function with options rather than creating a new copy.

### Mock Data

Never hardcode mock arrays/objects inline in page components. Extract to
`_utils.ts` alongside the feature.

```tsx
// Bad — mock data inline in page (300+ lines)
export default function SettingsPage() {
  const billingPlans = [
    { name: "Starter", price: 99 },
    { name: "Professional", price: 299 },
    // ... 20 more lines
  ];
  return <div>...</div>;
}

// Good — mock data in _utils.ts
// app/(org)/settings/_utils.ts
export const BILLING_PLANS = [
  { name: "Starter", price: 99 },
  { name: "Professional", price: 299 },
] as const;

// settings/page.tsx
import { BILLING_PLANS } from "./_utils";
```

## HTTP Client

Use `SwantaraService` from `@/lib/services/swantara/service` as the primary
HTTP client. It handles CSRF, token refresh, snake_case/camelCase
transformation, and error mapping automatically.

```ts
import { getSwantaraService } from "@/lib/services/swantara/service";

const service = getSwantaraService();
const data = await service.catalog.products.list(orgId);
```

For ad-hoc client-side requests outside the service layer (e.g., auth
endpoints), use the shared Axios instance from `@/lib/client/axios`. Never
create new Axios instances.

Raw `fetch()` is reserved for middleware and server contexts where Axios is
not available.

## State Management (Zustand)

- File: `*.store.ts`
- Separate state and action types, combine into store type.
- Hook: `use{Name}Store`.
- Always include `"use client"` at the top of store files.
- Store files live in `stores/`. Co-locate form stores in `_hooks/` only if
  feature-specific and not shared.

```tsx
type ThemeState = { theme: "light" | "dark" };
type ThemeActions = { setTheme: (theme: ThemeState["theme"]) => void };
export type ThemeStore = ThemeState & ThemeActions;
```

### Persistence

Use Zustand's `persist` middleware for stores that need localStorage sync.
Never manage `localStorage.getItem/setItem` manually inside actions.

```ts
import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";

export const useThemeStore = create<ThemeStore>()(
  persist(
    (set) => ({ ... }),
    {
      name: "swantara:theme",
      storage: createJSONStorage(() => localStorage),
    },
  ),
);
```

For stores that also need cookie sync (e.g., density), use `persist` with a
custom storage adapter or sync in `onRehydrateStorage`.

## Loading & Empty States

Use `SectionWrapper` or the DataTable `status` prop with the shared
status helper across all list and detail views. Never return `null`
while loading. Never render inline empty text.

### Loading State

For table views, use the DataTable `status` prop:

```tsx
<DataTable
  status={isLoading ? { type: "loading" } : isError ? { type: "error", message: "Failed to load.", onRetry: refetch } : undefined}
/>
```

### Empty State

Use the `EmptyState` component from `@/components/empty-state`. Never render
inline text for empty states.

```tsx
// Good
<EmptyState title="No items" description="Nothing here yet." />

// Bad — inline text
<p className="text-sm text-muted-foreground">No items</p>
```

### Not-Found State

Use `text-muted-foreground` for not-found states. Never use `text-destructive`
(not-found is informational, not an error).

```tsx
// Good
<p className="text-sm text-muted-foreground">Not found</p>

// Bad — destructive styling for a non-error state
<p className="text-sm text-destructive">{t("notFound")}</p>
```

## Data Formatting

### Currency

Always use `formatCurrency` from `@/lib/utils/formatters`. Never manually
prefix values with `$` or other currency symbols.

```tsx
// Good
import { formatCurrency } from "@/lib/utils/formatters";
<span>{formatCurrency(amount, { currency: orgCurrency })}</span>

// Bad — manual prefix
<span>{`$${formatNumber(amount)}`}</span>

// Bad — hardcoded currency code
<span>{formatCurrency(amount, { currency: "USD" })}</span>
```

### Dates

Always use `formatDate` from `@/lib/utils/formatters`. Never call
`toLocaleDateString()` directly without a locale argument. Never use
`new Date().toISOString().slice(0, 10)` for default form values (off
by one in non-UTC timezones); use the timezone-safe local-date helper.

```tsx
// Good
import { formatDate } from "@/lib/utils/formatters";
<span>{formatDate(date)}</span>

// Bad — no locale, format varies by browser
<span>{new Date(date).toLocaleDateString()}</span>

// Bad — hardcoded locale
<span>{new Date(date).toLocaleDateString("en-US", { ... })}</span>
```

## Forms

All forms use React Hook Form + Zod.

```tsx
function CustomerFormDialog() {
  const form = useForm<CustomerFormValues>({
    resolver: zodResolver(schema),
  });
  // ...
  return <Form form={form} onSubmit={handleSubmit}>...</Form>;
}
```

### Schema Location

Define schemas at module scope in `*-utils.ts` or the dialog module,
never inside the component body. Add `.trim()` to required strings
and `.refine()` for cross-field rules.

```ts
// feature/_components/feature-utils.ts
export const featureFormSchema = z.object({
  name: z.string().trim().min(1, { message: "Name is required." }),
  startDate: z.string().min(1),
  endDate: z.string().min(1),
}).refine((v) => v.endDate >= v.startDate, {
  message: "End date must be after start date.",
  path: ["endDate"],
});
```

**Auth forms** — expose the module-scope schema through dedicated
`_hooks/` files. **Org feature forms** — same module-scope pattern.
Do not mix inline render-scope schemas within the same feature area.

Submit buttons use `isPending`/`isSubmitting` for `disabled` state.
Reset form state on dialog open with different data, or key the
dialog to force remount.

## Data Fetching

### Query Keys

Use centralized query key factories from `lib/hooks/use-org-query.ts`.
Never define query keys inline in components or hooks.

```ts
// Good — centralized key factory
import { orgListQueryKey } from "@/lib/hooks/use-org-query";
const queryKey = orgListQueryKey({ endpoint: "/api/items", params });

// Good — ad-hoc hooks use consistent structure
const queryKey = ["products", orgId, itemId];

// Bad — raw strings in components with no orgId
queryClient.invalidateQueries({ queryKey: ["crmLeads"] });
```

**Key structure:** Always include `organizationId` for org-scoped data.
Global resources (currencies, Units) use a simple key without orgId.
Use `useOrgQuery` for single-item fetches, `useOrgListQuery` for lists.
Never fetch a list then `.find()` client-side for a detail view.

```ts
// Org-scoped
["products", orgId, itemId]
["partners", orgId, partnerId]

// Global
["currencies"]
["units"]
```

Guard queries with `enabled` when params may be empty. Validate
route-param numbers before sending to the API (`Number("")` is `0`).
Convert `orgId: string` to `number` once at the layout/page boundary
and pass `number` down, never `Number(orgId)` at every call site.
Guard `.find()` targets with `?? []` and null-check nested access.

### Query Options

Global defaults are set in `providers/tanstack-query.tsx` (staleTime: 2min,
gcTime: 30min, refetchOnWindowFocus: false, retry: 3). Do not override these
in individual hooks unless there is a documented reason.

### Query Invalidation

**Always use `useOrgMutation`** for mutations that need to invalidate
queries. Never call `queryClient.invalidateQueries` directly in component
bodies, even if it seems simpler.

```tsx
// Good — mutation with automatic invalidation
const mutation = useOrgMutation({
  endpoint: "/api/items",
  method: "POST",
  onSuccess: () => {
    toast.success("Saved.");
  },
});

// Bad — manual invalidation in component
const save = async () => {
  await fetch("/api/items", { method: "POST" });
  queryClient.invalidateQueries({ queryKey: ["items"] });
};
```

If you need to invalidate multiple query keys after a mutation, define them
in the mutation's `onSuccess` callback rather than scattering
`invalidateQueries` calls across component logic. Invalidate both list
and detail keys. Never use `refetch()` instead of invalidation, or the
list page stays stale. Never mutate cached arrays with `.sort()`; copy
first with `[...arr].sort()`.

**Never bypass the service layer** for API calls. The `useOrgMutation` hook
handles CSRF, token refresh, snake_case/camelCase transformation, and error
mapping automatically. Raw `fetch()` skips all of this. Never combine
manual `getSwantaraService()` + `useState` where React Query caching,
deduplication, and retry apply. Wrap service calls in `useMutation`
with `isPending` to prevent double-click races.

### Error Feedback

Always provide user feedback on failed mutations. At minimum, show a toast
with the error message. Never silently ignore errors. Prefer
`toast.promise()` for mutations to keep pending/success/error together.
Always add `.catch()` with `toast.error()` to fire-and-forget service
calls. Wrap async event handlers in try/catch.

```tsx
// Good
onSuccess: () => toast.success("Saved."),
onError: (error) => toast.error("Something went wrong. Please try again."),

// Bad — no error handling
onSuccess: () => {},
```

## Security

CSRF validation on all mutating API routes. Safe methods stay safe:
never clear cookies or mutate state on `GET` (logout-CSRF via `<img>`).
Validate proxy `[...path]` segments against a safe pattern and encode
per segment; reject `..`, null bytes, and encoded traversal. Never log
full upstream URLs with query params, tokens, or PII. Never forward
raw upstream bodies to clients; use generic messages. Distinguish
network errors from invalid refresh tokens in middleware; never clear
session on transient failure. Wrap `intlMiddleware` so refresh cookies
still apply on locale failure. Tokens only from URL hash fragment,
never query string. Validate token presence and shape before setting
cookies. Set `Secure` in production on session, locale, and density
cookies. Single token-refresh authority in middleware; no competing
401-retry loops in client and proxy.

## React Rules

Hooks before any early return, never after. All components use function
declarations, never arrow functions. Use `type`, never `interface`,
for props. No `any`, no `as never`, no `as unknown`; fix root type
causes instead. No deeply nested ternaries; use lookup maps or
helpers. Keep components under 150 lines; extract hooks and
sub-components. One `activeDialog` discriminator instead of multiple
boolean dialog states. Stable keys (`id`), never translatable labels
that can collide. Never store derived state in `useEffect`; compute
inline or with `useMemo`. Memoize `Map` derivations and inline
object/array JSX props that trigger re-renders. Use `React.memo` only
for pure leaf components in lists. `next/dynamic` only for heavy
components actually used in pages. Parallelize independent `await`s
with `Promise.all`. Clean up listeners, intervals, timeouts, blob
URLs, and injected iframes in `useEffect` return. Sync
URL-derived state with `useEffect` on param change. Reset dialog form
state on open or key to force remount.

## Accessibility

`<html lang>` reflects the active locale. `<main id="main">` backs the
skip link. `FormField` passes `aria-describedby` to inputs. No
contradictory `aria-hidden` parent with `sr-only` child. All touch
targets minimum 44x44 via padding.

## Metadata & Server Rendering

Every page exports locale-aware `generateMetadata` with title,
description, `openGraph`, and `alternates` for `en`/`id`. Auth and org
pages set `robots: noindex`; public pages stay indexable. Wrap client
components in `<Suspense>` for streaming. Never use server
`staleTime: "static"` for user-scoped data.

## Hygiene

No `console.*` in production code; use `@/lib/client/logger` or
`@/lib/server/logger`. No unused imports, exports, hooks, stores,
providers, utils, components, or CSS. Remove dead modules instead of
shipping them. No magic numbers; extract timeout durations to named
constants. Standardize detail navigation (`BackLink` + detail) and
error display (`ErrorState`, `toast.promise`). No hardcoded mock
business data in production components. No `isLoading` in React Query
v5; use `isPending`.
