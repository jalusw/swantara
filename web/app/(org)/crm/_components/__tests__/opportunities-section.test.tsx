import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { OpportunitiesSection } from "../opportunities-section";

const opportunities = [
  {
    id: 11,
    organization_id: 1,
    name: "Acme Enterprise Deal",
    type: "opportunity",
    contact_id: 1,
    contact_name: "Aria Chen",
    email: "aria@example.com",
    phone: null,
    job_position: null,
    stage_id: 2,
    expected_revenue: 50000,
    probability: 60,
    priority: 1,
    salesperson_id: null,
    sales_group_id: null,
    source: "referral",
    medium: null,
    campaign: null,
    lost_reason: null,
    expected_close: "2026-06-30",
    closed_at: null,
    is_won: false,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: 12,
    organization_id: 1,
    name: "Nusantara Rollout",
    type: "opportunity",
    contact_id: 2,
    contact_name: "June Park",
    email: "june@example.com",
    phone: null,
    job_position: null,
    stage_id: 3,
    expected_revenue: 80000,
    probability: 100,
    priority: 1,
    salesperson_id: null,
    sales_group_id: null,
    source: "website",
    medium: null,
    campaign: null,
    lost_reason: null,
    expected_close: "2026-05-31",
    closed_at: "2026-04-01T00:00:00Z",
    is_won: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-04-01T00:00:00Z",
  },
];

const stages = [
  { id: 2, name: "Proposal", sequence: 2, is_won: false, probability: 60 },
  { id: 3, name: "Won", sequence: 4, is_won: true, probability: 100 },
];

function useOpportunityHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/crm/opportunities", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { opportunities } }),
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
  useOpportunityHandlers();
});

describe("OpportunitiesSection", () => {
  it("renders opportunities with stage badges", async () => {
    renderWithProviders(<OpportunitiesSection orgId="1" />);

    expect(await screen.findByText("Acme Enterprise Deal")).toBeInTheDocument();
    expect(screen.getByText("Nusantara Rollout")).toBeInTheDocument();
    expect(screen.getByText("Proposal")).toBeInTheDocument();
    expect(screen.getByText("Create quotation")).toBeInTheDocument();
  });

  it("filters opportunities through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<OpportunitiesSection orgId="1" />);

    await screen.findByText("Acme Enterprise Deal");
    await user.type(screen.getByPlaceholderText("Search opportunities…"), "Nusantara");

    expect(await screen.findByText("Nusantara Rollout")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Acme Enterprise Deal")).not.toBeInTheDocument());
  });

  it("opens the opportunity dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<OpportunitiesSection orgId="1" />);

    await screen.findByText("Acme Enterprise Deal");
    await user.click(screen.getByRole("button", { name: "Add opportunity" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("New opportunity")).toBeInTheDocument();
  });
});
