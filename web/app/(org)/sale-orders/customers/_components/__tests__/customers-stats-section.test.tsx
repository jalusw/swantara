import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { CustomersStats } from "../customers-stats-section";

const contacts = [
  {
    id: 1,
    organization_id: 1,
    name: "Acme Corporation",
    display_name: "Acme Corp",
    is_organization: true,
    parent_id: null,
    email: "hello@acme.co",
    phone: null,
    mobile: null,
    website: null,
    tax_id: null,
    industry: null,
    currency_code: "USD",
    lang: "en",
    active: true,
    created_at: "2026-01-05T00:00:00Z",
    updated_at: "2026-01-05T00:00:00Z",
  },
  {
    id: 2,
    organization_id: 1,
    name: "Globex Corporation",
    display_name: "Globex Inc",
    is_organization: true,
    parent_id: null,
    email: "ops@globex.co",
    phone: null,
    mobile: null,
    website: null,
    tax_id: null,
    industry: null,
    currency_code: "USD",
    lang: "en",
    active: true,
    created_at: "2026-01-06T00:00:00Z",
    updated_at: "2026-01-06T00:00:00Z",
  },
  {
    id: 3,
    organization_id: 1,
    name: "Initech LLC",
    display_name: "Initech",
    is_organization: true,
    parent_id: null,
    email: null,
    phone: null,
    mobile: null,
    website: null,
    tax_id: null,
    industry: null,
    currency_code: "USD",
    lang: "en",
    active: false,
    created_at: "2026-01-07T00:00:00Z",
    updated_at: "2026-01-07T00:00:00Z",
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts } }),
    ),
  );
});

describe("CustomersStats", () => {
  it("renders total, active and new-this-month labels", async () => {
    renderWithProviders(<CustomersStats />);

    expect(await screen.findByText("Total customers")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByText("New this month")).toBeInTheDocument();
  });

  it("keeps counts visible when stats are clicked", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CustomersStats />);

    await screen.findByText("Total customers");
    await user.click(screen.getByText("Total customers"));

    expect(screen.getByText("3")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
  });
});
