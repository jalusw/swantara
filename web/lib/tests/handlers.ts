import { HttpResponse, http } from "msw";

// MSW serves hooks/components over fetch. Route handlers are tested directly via
// `callRoute` from "./route-handler" — no HTTP layer involved.
//
// Responses mirror the backend envelope and use snake_case keys; the service
// layer camel-cases them via its response interceptor.

const profile = {
  id: 1,
  username: "alex",
  first_name: "Alex",
  last_name: "Rivera",
  email: "alex@acme.com",
  phone: null,
  avatar: null,
  bio: null,
  birthday: null,
  active: true,
  sex: null,
  address: null,
  city: null,
  postal_code: null,
  system_role_id: 1,
  email_verified_at: null,
  phone_verified_at: null,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

const organizations = [
  { id: 1, name: "Acme Inc" },
  { id: 2, name: "PT Nusantara" },
];

const defaultModules = [
  "crm",
  "products",
  "inventory",
  "procurement",
  "hr",
  "finance",
  "projects",
  "quality",
  "subscriptions",
  "pos",
  "service",
].map((module_id) => ({ module_id, active: true }));

const permissionCodes = [
  "contact.view",
  "contact.create",
  "employee.view",
  "journal_entry.view",
  "reporting.view",
  "organization.view",
  "member.view",
  "subscription.view",
  "prospect.view",
  "sale_order.view",
  "production_order.view",
  "purchase_order.view",
  "item.view",
  "leave_request.view",
  "attendance.view",
  "timesheet.view",
  "payroll_run.view",
  "project.view",
  "expense_report.view",
  "fixed_asset.view",
  "quality_check.view",
  "service_order.view",
  "commission_plan.view",
];

const permissions = permissionCodes.map((code, index) => {
  const [resource, action] = code.split(".");
  return {
    id: index + 1,
    name: `${action} ${resource}`,
    code,
    description: `Permission to ${action} ${resource}`,
    resource,
    action,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  };
});

const members = [
  {
    id: 1,
    user_id: 1,
    organization_id: 1,
    position: null,
    created_at: "2026-01-05T00:00:00Z",
    updated_at: "2026-01-05T00:00:00Z",
    user: {
      id: 1,
      username: "alex",
      first_name: "Alex",
      last_name: "Rivera",
      email: "alex@acme.com",
      active: true,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
    roles: [
      {
        id: 1,
        organization_id: 1,
        name: "Owner",
        code: "owner",
        description: null,
        created_at: "2026-01-05T00:00:00Z",
        updated_at: "2026-01-05T00:00:00Z",
      },
    ],
  },
  {
    id: 2,
    user_id: 2,
    organization_id: 1,
    position: "Head of Operations",
    created_at: "2026-02-10T00:00:00Z",
    updated_at: "2026-02-10T00:00:00Z",
    user: {
      id: 2,
      username: "june",
      first_name: "June",
      last_name: "Park",
      email: "june@acme.com",
      active: true,
      created_at: "2026-02-01T00:00:00Z",
      updated_at: "2026-02-01T00:00:00Z",
    },
    roles: [
      {
        id: 2,
        organization_id: 1,
        name: "Admin",
        code: "admin",
        description: null,
        created_at: "2026-02-10T00:00:00Z",
        updated_at: "2026-02-10T00:00:00Z",
      },
    ],
  },
  {
    id: 3,
    user_id: 3,
    organization_id: 1,
    position: null,
    created_at: "2026-03-01T00:00:00Z",
    updated_at: "2026-03-01T00:00:00Z",
    roles: [],
  },
];

type DimensionFixture = {
  id: number;
  organization_id: number;
  name: string;
  code: string | null;
  kind: string | null;
  parent_id: number | null;
  active: boolean;
  created_at: string;
  updated_at: string;
};

type FxRateFixture = {
  id: number;
  organization_id: number;
  currency_code: string;
  rate: number;
  rate_type: string;
  valid_from: string;
  created_at: string;
  updated_at: string;
};

type UnitCategoryFixture = {
  id: number;
  name: string;
  created_at: string;
  updated_at: string;
};

type UnitFixture = {
  id: number;
  category_id: number;
  name: string;
  factor: number;
  unit_type: string;
  rounding: number;
  created_at: string;
  updated_at: string;
};

type PaymentTermFixture = {
  id: number;
  name: string;
  note: string | null;
  created_at: string;
  updated_at: string;
};

function createDimensionFixtures(): DimensionFixture[] {
  return [
    {
      id: 1,
      organization_id: 1,
      name: "Operating costs",
      code: "OPEX",
      kind: "expense",
      parent_id: null,
      active: true,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
    {
      id: 2,
      organization_id: 1,
      name: "Salaries",
      code: "OPEX-SAL",
      kind: "expense",
      parent_id: 1,
      active: true,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
    {
      id: 3,
      organization_id: 1,
      name: "Revenue",
      code: "REV",
      kind: "revenue",
      parent_id: null,
      active: true,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
  ];
}

function createFxRateFixtures(): FxRateFixture[] {
  return [
    {
      id: 1,
      organization_id: 1,
      currency_code: "IDR",
      rate: 16000,
      rate_type: "spot",
      valid_from: "2026-01-01",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
    {
      id: 2,
      organization_id: 1,
      currency_code: "EUR",
      rate: 0.92,
      rate_type: "avg",
      valid_from: "2026-01-01",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
  ];
}

function createUnitCategoryFixtures(): UnitCategoryFixture[] {
  return [
    {
      id: 1,
      name: "Length",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
    {
      id: 2,
      name: "Weight",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
    {
      id: 3,
      name: "Volume",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
  ];
}

function createUnitFixtures(): UnitFixture[] {
  return [
    {
      id: 1,
      category_id: 1,
      name: "Meter",
      factor: 1,
      unit_type: "reference",
      rounding: 0.001,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
    {
      id: 2,
      category_id: 2,
      name: "Kilogram",
      factor: 1,
      unit_type: "reference",
      rounding: 0.001,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
  ];
}

function createPaymentTermFixtures(): PaymentTermFixture[] {
  return [
    {
      id: 1,
      name: "Net 30",
      note: "Due in 30 days",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
    {
      id: 2,
      name: "50/50 split",
      note: "Half upfront, half on delivery",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
  ];
}

let dimensionFixtures = createDimensionFixtures();
let fxRateFixtures = createFxRateFixtures();
let unitGroupFixtures = createUnitCategoryFixtures();
let uomFixtures = createUnitFixtures();
let paymentTermFixtures = createPaymentTermFixtures();

function nextFixtureId(fixtures: { id: number }[]) {
  return fixtures.reduce((max, item) => Math.max(max, item.id), 0) + 1;
}

export function resetReferenceFixtures() {
  dimensionFixtures = createDimensionFixtures();
  fxRateFixtures = createFxRateFixtures();
  unitGroupFixtures = createUnitCategoryFixtures();
  uomFixtures = createUnitFixtures();
  paymentTermFixtures = createPaymentTermFixtures();
}

export const handlers = [
  http.post("*/api/v1/auth/login", () =>
    HttpResponse.json({
      success: true,
      message: "Authenticated.",
    }),
  ),
  http.post("*/api/v1/auth/logout", () =>
    HttpResponse.json({
      success: true,
      message: "Logout success.",
    }),
  ),
  http.post("*/api/v1/auth/password-reset/request", () =>
    HttpResponse.json({
      success: true,
      message: "Password reset link sent.",
    }),
  ),
  http.post("*/api/v1/auth/password-reset", () =>
    HttpResponse.json({
      success: true,
      message: "Password reset successful.",
    }),
  ),
  http.post("*/api/v1/auth/email/check", () =>
    HttpResponse.json({
      success: true,
      message: "Email available.",
      data: { available: true },
    }),
  ),
  http.post("*/api/v1/auth/register", () =>
    HttpResponse.json(
      {
        success: true,
        message: "Registration successful.",
      },
      { status: 201 },
    ),
  ),
  http.post("*/api/v1/organizations", () =>
    HttpResponse.json(
      {
        success: true,
        message: "Organization created.",
      },
      { status: 201 },
    ),
  ),
  http.post("*/api/v1/organizations/quick", () =>
    HttpResponse.json(
      {
        success: true,
        message: "Organization created.",
      },
      { status: 201 },
    ),
  ),
  http.get("*/api/v1/me", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: { user: profile },
    }),
  ),
  http.get("*/api/v1/me/organizations", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: { organizations },
    }),
  ),
  http.get("*/api/v1/me/organizations/:organizationId/permissions", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: { permissions },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/members", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: { members },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/modules", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: { modules: defaultModules },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/kpis/finance", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        kpi: {
          revenue: 84320,
          expenses: 52300,
          gross_margin_pct: 0.38,
          net_margin_pct: 0.22,
          ebitda: 32020,
          current_ratio: 1.8,
        },
      },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/kpis/sales", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        kpi: {
          bookings: 120000,
          revenue: 84320,
          cogs: 41000,
          gross_margin: 43320,
          gross_margin_pct: 0.514,
          win_rate: 0.64,
        },
      },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/kpis/pipeline", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        kpi: {
          total_expected_revenue: 180000,
          weighted_pipeline: 95000,
          win_rate: 0.42,
          stages: 5,
        },
      },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/kpis/inventory", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        kpi: {
          on_hand_value: 245000,
          on_hand_quantity: 3200,
          product_count: 96,
        },
      },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/kpis/subscription", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        kpi: {
          mrr: 12500,
          arr: 150000,
          churned: 3,
          churn_rate: 0.02,
          ltv: 45000,
        },
      },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/kpis/projects", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        kpi: {
          project_count: 12,
          total_margin: 85000,
          total_cost: 120000,
          total_billed: 205000,
          utilization: 0.75,
        },
      },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/kpis/payroll", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        kpi: {
          gross_cost: 45000,
          net_cost: 38000,
        },
      },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/kpis/procurement", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        kpi: {
          avg_cycle_days: 12.5,
          on_time_delivery_pct: 0.88,
          price_variance_pct: 0.025,
          purchase_count: 48,
        },
      },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/kpis/manufacturing", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        kpi: {
          oee_pct: 0.85,
          yield_pct: 0.96,
          scrap_pct: 0.02,
          cost_variance_pct: 0.015,
          order_count: 24,
        },
      },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/kpis/ar-ap", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        kpi: {
          dso: 32.5,
          dpo: 28.3,
          overdue_ar_pct: 0.12,
          overdue_ap_pct: 0.08,
        },
      },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/kpis/cash", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        kpi: {
          position: 125000,
          burn: 15000,
          forecast: 140000,
        },
      },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/kpis/inventory-ratio", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        kpi: {
          turnover: 4.2,
          days_on_hand: 87.5,
          stockout_count: 3,
        },
      },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/contacts", () => {
    const initial = [
      {
        id: 1,
        organization_id: 1,
        name: "Bluebird Trading Pte. Ltd.",
        display_name: "Bluebird Trading",
        is_organization: true,
        parent_id: null,
        email: "billing@bluebird.sg",
        phone: "+65 6123 4567",
        mobile: "+65 8123 4567",
        website: "https://bluebird.sg",
        tax_id: "202012345K",
        industry: "retail",
        currency_code: "SGD",
        lang: "en",
        active: true,
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
      {
        id: 2,
        organization_id: 1,
        name: "PT Nusantara Logistics",
        display_name: "Nusantara Logistics",
        is_organization: true,
        parent_id: null,
        email: "finance@nusantaralog.id",
        phone: "+62 21 5550 1188",
        mobile: "+62 811 5550 1188",
        website: "https://nusantaralog.id",
        tax_id: "02.345.678.9-002.000",
        industry: "logistics",
        currency_code: "IDR",
        lang: "id",
        active: true,
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
      {
        id: 3,
        organization_id: 1,
        name: "Klima Foods GmbH",
        display_name: "Klima Foods",
        is_organization: true,
        parent_id: null,
        email: "team@klimafoods.de",
        phone: "+49 30 1204 550",
        mobile: "",
        website: "https://klimafoods.de",
        tax_id: "DE312345678",
        industry: "food",
        currency_code: "EUR",
        lang: "de",
        active: false,
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
      {
        id: 4,
        organization_id: 1,
        name: "Aria Chen",
        display_name: "Aria Chen",
        is_organization: false,
        parent_id: null,
        email: "aria.chen@example.com",
        phone: "+1 415 555 0134",
        mobile: "+1 415 555 0134",
        website: "",
        tax_id: "",
        industry: "",
        currency_code: "USD",
        lang: "en",
        active: true,
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
    ];
    const generic = Array.from({ length: 1280 }, (_, i) => ({
      id: i + 5,
      organization_id: 1,
      name: `Contact ${i + 5}`,
      display_name: null,
      is_organization: i % 3 === 0,
      parent_id: null,
      email: `contact${i + 5}@example.com`,
      phone: null,
      mobile: null,
      website: null,
      tax_id: null,
      industry: null,
      currency_code: null,
      lang: "en",
      active: true,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    }));
    const contacts = [...initial, ...generic];
    return HttpResponse.json({
      success: true,
      message: "OK.",
      data: { contacts },
      meta: {
        pagination: { page: 1, perPage: 50, total: 1284, totalPages: 26 },
      },
    });
  }),
  http.get("*/api/v1/organizations/:organizationId/employees", () => {
    const employees = Array.from({ length: 48 }, (_, i) => ({
      id: i + 1,
      organization_id: 1,
      contact_id: i + 1,
      employee_number: `EMP-${String(i + 1).padStart(4, "0")}`,
      job_position_id: null,
      department_id: null,
      manager_id: null,
      hire_date: "2024-01-01",
      active: true,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    }));
    return HttpResponse.json({
      success: true,
      message: "OK.",
      data: { employees },
      meta: { pagination: { page: 1, perPage: 50, total: 48, totalPages: 1 } },
    });
  }),
  http.get("*/api/v1/organizations/:organizationId/dimensions", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: { accounts: dimensionFixtures },
    }),
  ),
  http.post("*/api/v1/organizations/:organizationId/dimensions", async ({ request, params }) => {
    const body = (await request.json()) as {
      name: string;
      code?: string;
      kind?: string;
      parent_id?: number | null;
      active?: boolean;
    };
    const account = {
      id: nextFixtureId(dimensionFixtures),
      organization_id: Number(params.organizationId),
      name: body.name,
      code: body.code ?? null,
      kind: body.kind ?? null,
      parent_id: body.parent_id ?? null,
      active: body.active ?? true,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    };
    dimensionFixtures.push(account);
    return HttpResponse.json(
      {
        success: true,
        message: "Created.",
        data: { account },
      },
      { status: 201 },
    );
  }),
  http.put(
    "*/api/v1/organizations/:organizationId/dimensions/:accountId",
    async ({ request, params }) => {
      const body = (await request.json()) as {
        name?: string;
        code?: string;
        kind?: string;
        parent_id?: number | null;
        active?: boolean;
      };
      const account = dimensionFixtures.find((item) => item.id === Number(params.accountId));
      if (!account) {
        return HttpResponse.json({ success: false, message: "Not found." }, { status: 404 });
      }
      if (body.name !== undefined) account.name = body.name;
      if (body.code !== undefined) account.code = body.code;
      if (body.kind !== undefined) account.kind = body.kind;
      if (body.parent_id !== undefined) account.parent_id = body.parent_id;
      if (body.active !== undefined) account.active = body.active;
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { account },
      });
    },
  ),
  http.delete("*/api/v1/organizations/:organizationId/dimensions/:accountId", ({ params }) => {
    dimensionFixtures = dimensionFixtures.filter((item) => item.id !== Number(params.accountId));
    return HttpResponse.json({
      success: true,
      message: "Deleted.",
    });
  }),
  http.get("*/api/v1/currencies", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        currencies: [
          {
            code: "USD",
            name: "US Dollar",
            symbol: "$",
            decimal_places: 2,
            rounding: 0.01,
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z",
          },
          {
            code: "IDR",
            name: "Indonesian Rupiah",
            symbol: "Rp",
            decimal_places: 0,
            rounding: 1,
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z",
          },
          {
            code: "EUR",
            name: "Euro",
            symbol: "€",
            decimal_places: 2,
            rounding: 0.01,
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z",
          },
        ],
      },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/fx-rates", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: { fx_rates: fxRateFixtures },
    }),
  ),
  http.post("*/api/v1/organizations/:organizationId/fx-rates", async ({ request, params }) => {
    const body = (await request.json()) as {
      currency_code: string;
      rate: number;
      rate_type: string;
      valid_from: string;
    };
    const fxRate = {
      id: nextFixtureId(fxRateFixtures),
      organization_id: Number(params.organizationId),
      currency_code: body.currency_code,
      rate: body.rate,
      rate_type: body.rate_type,
      valid_from: body.valid_from,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    };
    fxRateFixtures.push(fxRate);
    return HttpResponse.json(
      {
        success: true,
        message: "Created.",
        data: { fx_rate: fxRate },
      },
      { status: 201 },
    );
  }),
  http.put(
    "*/api/v1/organizations/:organizationId/fx-rates/:fxRateId",
    async ({ request, params }) => {
      const body = (await request.json()) as {
        rate?: number;
        rate_type?: string;
        valid_from?: string;
      };
      const fxRate = fxRateFixtures.find((item) => item.id === Number(params.fxRateId));
      if (!fxRate) {
        return HttpResponse.json({ success: false, message: "Not found." }, { status: 404 });
      }
      if (body.rate !== undefined) fxRate.rate = body.rate;
      if (body.rate_type !== undefined) fxRate.rate_type = body.rate_type;
      if (body.valid_from !== undefined) fxRate.valid_from = body.valid_from;
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { fx_rate: fxRate },
      });
    },
  ),
  http.delete("*/api/v1/organizations/:organizationId/fx-rates/:fxRateId", ({ params }) => {
    fxRateFixtures = fxRateFixtures.filter((item) => item.id !== Number(params.fxRateId));
    return HttpResponse.json({
      success: true,
      message: "Deleted.",
    });
  }),
  http.get("*/api/v1/unit-categories", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: { categories: unitGroupFixtures },
    }),
  ),
  http.post("*/api/v1/unit-categories", async ({ request }) => {
    const body = (await request.json()) as { name: string };
    const category = {
      id: nextFixtureId(unitGroupFixtures),
      name: body.name,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    };
    unitGroupFixtures.push(category);
    return HttpResponse.json(
      {
        success: true,
        message: "Created.",
        data: { category },
      },
      { status: 201 },
    );
  }),
  http.get("*/api/v1/units", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: { units: uomFixtures },
    }),
  ),
  http.post("*/api/v1/units", async ({ request }) => {
    const body = (await request.json()) as {
      category_id: number;
      name: string;
      factor: number;
      unit_type?: string;
      rounding?: number;
    };
    const unit = {
      id: nextFixtureId(uomFixtures),
      category_id: body.category_id,
      name: body.name,
      factor: body.factor,
      unit_type: body.unit_type ?? "reference",
      rounding: body.rounding ?? 0.001,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    };
    uomFixtures.push(unit);
    return HttpResponse.json(
      {
        success: true,
        message: "Created.",
        data: { unit },
      },
      { status: 201 },
    );
  }),
  http.put("*/api/v1/units/:unitId", async ({ request, params }) => {
    const body = (await request.json()) as {
      name?: string;
      factor?: number;
      unit_type?: string;
      rounding?: number;
    };
    const unit = uomFixtures.find((item) => item.id === Number(params.unitId));
    if (!unit) {
      return HttpResponse.json({ success: false, message: "Not found." }, { status: 404 });
    }
    if (body.name !== undefined) unit.name = body.name;
    if (body.factor !== undefined) unit.factor = body.factor;
    if (body.unit_type !== undefined) unit.unit_type = body.unit_type;
    if (body.rounding !== undefined) unit.rounding = body.rounding;
    return HttpResponse.json({
      success: true,
      message: "OK.",
      data: { unit },
    });
  }),
  http.delete("*/api/v1/units/:unitId", ({ params }) => {
    uomFixtures = uomFixtures.filter((item) => item.id !== Number(params.unitId));
    return HttpResponse.json({
      success: true,
      message: "Deleted.",
    });
  }),
  http.get("*/api/v1/payment-terms", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: { payment_terms: paymentTermFixtures },
    }),
  ),
  http.post("*/api/v1/payment-terms", async ({ request }) => {
    const body = (await request.json()) as { name: string; note?: string };
    const paymentTerm = {
      id: nextFixtureId(paymentTermFixtures),
      name: body.name,
      note: body.note ?? null,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    };
    paymentTermFixtures.push(paymentTerm);
    return HttpResponse.json(
      {
        success: true,
        message: "Created.",
        data: { payment_term: paymentTerm },
      },
      { status: 201 },
    );
  }),
  http.put("*/api/v1/payment-terms/:paymentTermId", async ({ request, params }) => {
    const body = (await request.json()) as { name?: string; note?: string };
    const paymentTerm = paymentTermFixtures.find(
      (item) => item.id === Number(params.paymentTermId),
    );
    if (!paymentTerm) {
      return HttpResponse.json({ success: false, message: "Not found." }, { status: 404 });
    }
    if (body.name !== undefined) paymentTerm.name = body.name;
    if (body.note !== undefined) paymentTerm.note = body.note;
    return HttpResponse.json({
      success: true,
      message: "OK.",
      data: { payment_term: paymentTerm },
    });
  }),
  http.delete("*/api/v1/payment-terms/:paymentTermId", ({ params }) => {
    paymentTermFixtures = paymentTermFixtures.filter(
      (item) => item.id !== Number(params.paymentTermId),
    );
    return HttpResponse.json({
      success: true,
      message: "Deleted.",
    });
  }),
  http.get("*/api/v1/organizations/:organizationId/contacts/:contactId/addresses", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: { addresses: [] },
    }),
  ),
  http.get("*/api/v1/organizations/:organizationId/contacts/:contactId/bank-accounts", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: { bank_accounts: [] },
    }),
  ),
];
