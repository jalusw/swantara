import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { CustomersTable } from "../customers-table-section";

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
    active: false,
    created_at: "2026-01-06T00:00:00Z",
    updated_at: "2026-01-06T00:00:00Z",
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts } }),
    ),
  );
});

describe("CustomersTable", () => {
  it("renders seeded customers", async () => {
    renderWithProviders(<CustomersTable orgId="1" />);

    expect(await screen.findByText("Acme Corp")).toBeInTheDocument();
    expect(screen.getByText("Globex Inc")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CustomersTable orgId="1" />);

    await screen.findByText("Acme Corp");
    await user.click(screen.getByRole("button", { name: "Add customer" }));

    expect(await screen.findByRole("heading", { name: "New customer" })).toBeInTheDocument();
  });
});
