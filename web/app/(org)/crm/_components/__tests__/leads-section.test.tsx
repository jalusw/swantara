import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { LeadsSection } from "../leads-section";

const leads = [
  {
    id: 1,
    organization_id: 1,
    name: "Acme Website Inquiry",
    type: "lead",
    contact_id: null,
    contact_name: "Aria Chen",
    email: "aria@example.com",
    phone: null,
    job_position: null,
    stage_id: 1,
    expected_revenue: 5000,
    probability: 10,
    priority: 1,
    salesperson_id: null,
    sales_group_id: null,
    source: "website",
    medium: null,
    campaign: null,
    lost_reason: null,
    expected_close: null,
    closed_at: null,
    is_won: false,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: 2,
    organization_id: 1,
    name: "Trade Show Follow-up",
    type: "lead",
    contact_id: null,
    contact_name: "June Park",
    email: "june@example.com",
    phone: null,
    job_position: null,
    stage_id: 1,
    expected_revenue: 12000,
    probability: 20,
    priority: 2,
    salesperson_id: null,
    sales_group_id: null,
    source: "event",
    medium: null,
    campaign: null,
    lost_reason: null,
    expected_close: null,
    closed_at: null,
    is_won: false,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

const stages = [
  { id: 1, name: "New", sequence: 1, is_won: false, probability: 10 },
  { id: 2, name: "Qualified", sequence: 2, is_won: false, probability: 30 },
];

function useLeadHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/crm/leads", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { leads } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/stages", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { stages } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/teams", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { teams: [] } }),
    ),
  );
}

beforeEach(() => {
  useLeadHandlers();
});

describe("LeadsSection", () => {
  it("renders leads with contacts and stage badges", async () => {
    renderWithProviders(<LeadsSection orgId="1" />);

    expect(await screen.findByText("Acme Website Inquiry")).toBeInTheDocument();
    expect(screen.getByText("Trade Show Follow-up")).toBeInTheDocument();
    expect(screen.getByText("Aria Chen")).toBeInTheDocument();
    expect(screen.getAllByText("New").length).toBeGreaterThan(0);
  });

  it("filters leads through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<LeadsSection orgId="1" />);

    await screen.findByText("Acme Website Inquiry");
    await user.type(screen.getByPlaceholderText("Cari prospek…"), "Trade Show");

    expect(await screen.findByText("Trade Show Follow-up")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Acme Website Inquiry")).not.toBeInTheDocument());
  });

  it("opens the lead dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<LeadsSection orgId="1" />);

    await screen.findByText("Acme Website Inquiry");
    await user.click(screen.getByRole("button", { name: "Tambah Prospek" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Prospek baru" })).toBeInTheDocument();
  });
});
