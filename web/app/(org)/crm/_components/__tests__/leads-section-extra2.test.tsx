import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { LeadsSection } from "../leads-section";

const STAMP = "2026-01-01T00:00:00Z";

function lead(id: number, patch: Record<string, unknown>) {
  return {
    id,
    organization_id: 1,
    name: `Lead ${id}`,
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
    created_at: STAMP,
    updated_at: STAMP,
    ...patch,
  };
}

const stages = [{ id: 1, name: "New", sequence: 1, is_won: false, probability: 10 }];

function seedLeads(leads: unknown[]) {
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
  seedLeads([lead(1, {}), lead(2, { contact_name: null, stage_id: null })]);
});

describe("LeadsSection extra2", () => {
  it("falls back to dash for missing contact and stage", async () => {
    renderWithProviders(<LeadsSection orgId="1" />);

    expect(await screen.findByText("Lead 1")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("shows the raw stage id when the stage is unknown", async () => {
    seedLeads([lead(1, {}), lead(2, { stage_id: 777 })]);
    renderWithProviders(<LeadsSection orgId="1" />);

    expect(await screen.findByText("777")).toBeInTheDocument();
  });

  it("opens the promote dialog from the promote action", async () => {
    const user = userEvent.setup();
    renderWithProviders(<LeadsSection orgId="1" />);

    await screen.findByText("Lead 1");
    await user.click(screen.getAllByRole("button", { name: "Promosikan ke peluang" })[0]!);

    expect(await screen.findByText("Promosikan prospek menjadi peluang")).toBeInTheDocument();
  });

  it("hides the promote action for closed or non-lead rows", async () => {
    seedLeads([
      lead(1, {}),
      lead(3, { type: "opportunity" }),
      lead(4, { closed_at: "2026-03-01T00:00:00Z" }),
    ]);
    renderWithProviders(<LeadsSection orgId="1" />);

    await screen.findByText("Lead 3");
    expect(screen.getAllByRole("button", { name: "Promosikan ke peluang" })).toHaveLength(1);
  });

  it("deletes a lead from the row actions", async () => {
    let deleteCalls = 0;
    server.use(
      http.delete("*/api/v1/organizations/:organizationId/crm/leads/:id", () => {
        deleteCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<LeadsSection orgId="1" />);

    await screen.findByText("Lead 1");
    const deleteTriggers = screen
      .getAllByRole("button", { name: "Hapus prospek" })
      .filter((el) => el.getAttribute("aria-haspopup") === "dialog");
    await user.click(deleteTriggers[0]!);
    await user.click(await screen.findByRole("button", { name: "Hapus" }));

    await waitFor(() => expect(deleteCalls).toBe(1));
  });

  it("opens the edit dialog from the row actions", async () => {
    const user = userEvent.setup();
    renderWithProviders(<LeadsSection orgId="1" />);

    await screen.findByText("Lead 1");
    await user.click(screen.getAllByRole("button", { name: "Ubah prospek" })[0]!);

    expect(await screen.findByRole("heading", { name: "Ubah prospek" })).toBeInTheDocument();
  });

  it("shows the error state with retry", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/crm/leads", () =>
        HttpResponse.json({ success: false, message: "lead boom" }, { status: 500 }),
      ),
    );
    renderWithProviders(<LeadsSection orgId="1" />);

    expect(await screen.findByText("lead boom")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Coba lagi" })).toBeInTheDocument();
  });
});
