import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PipelineSection } from "../pipeline-section";

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
];

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
];

const stages = [
  { id: 1, name: "New", sequence: 1, is_won: false, probability: 10 },
  { id: 2, name: "Proposal", sequence: 2, is_won: false, probability: 60 },
  { id: 3, name: "Won", sequence: 3, is_won: true, probability: 100 },
];

const pipeline = {
  total_expected_revenue: 50000,
  weighted_pipeline: 30000,
  win_rate: 0,
  stages: [
    {
      stage_id: 2,
      stage_name: "Proposal",
      probability: 60,
      opportunity_count: 1,
      expected_revenue: 50000,
      weighted_revenue: 30000,
    },
  ],
};

let winCalls = 0;

function usePipelineHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/crm/leads", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { leads } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/opportunities", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { opportunities } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/stages", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { stages } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/pipeline", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { pipeline } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/crm/opportunities/:id/win", () => {
      winCalls += 1;
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { opportunity: { ...opportunities[0], is_won: true } },
      });
    }),
  );
}

beforeEach(() => {
  winCalls = 0;
  usePipelineHandlers();
});

describe("PipelineSection", () => {
  it("renders KPIs, stage forecast wall, and the pipeline board", async () => {
    renderWithProviders(<PipelineSection orgId="1" />);

    expect((await screen.findAllByText("Weighted pipeline")).length).toBeGreaterThan(0);
    expect(screen.getByText("Pipeline board")).toBeInTheDocument();
    expect(screen.getByText("Acme Enterprise Deal")).toBeInTheDocument();
    expect(screen.getByText("New")).toBeInTheDocument();
  });

  it("marks the opportunity won from the quick win action", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PipelineSection orgId="1" />);

    await screen.findByText("Acme Enterprise Deal");
    await user.click(screen.getByRole("button", { name: "Mark won: Acme Enterprise Deal" }));

    await waitFor(() => expect(winCalls).toBe(1));
  });
});
