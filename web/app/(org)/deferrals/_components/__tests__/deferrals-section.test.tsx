import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { DeferralsSection } from "../deferrals-section";

const schedules = [
  {
    id: 1,
    organization_id: 1,
    type: "deferred_revenue",
    source_type: "invoice",
    source_id: 12,
    total_amount: 1200,
    balance_sheet_account_id: 1,
    pl_account_id: 2,
    method: "linear",
    date_start: "2026-02-01",
    state: "running",
    recognized_amount: 200,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/deferrals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { schedules } }),
    ),
  );
});

describe("DeferralsSection", () => {
  it("renders seeded deferral schedules", async () => {
    renderWithProviders(<DeferralsSection orgId="1" />);

    expect(await screen.findByText("invoice #12")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DeferralsSection orgId="1" />);

    await screen.findByText("invoice #12");
    await user.click(screen.getByRole("button", { name: "Buat penangguhan" }));

    expect(await screen.findByRole("heading", { name: "Buat penangguhan" })).toBeInTheDocument();
  });
});
