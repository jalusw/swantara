import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { DeferralDetail } from "../deferral-detail-section";

const schedule = {
  id: 7,
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
};

const lines = [
  {
    id: 3,
    schedule_id: 7,
    sequence: 1,
    recognition_date: "2026-02-01",
    amount: 200,
    posted: true,
    entry_id: 44,
  },
  {
    id: 4,
    schedule_id: 7,
    sequence: 2,
    recognition_date: null,
    amount: 200,
    posted: false,
    entry_id: null,
  },
];

let recognizeCalled = false;

function seedRunning() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/deferrals/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { schedule } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/deferrals/:id/lines", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { lines } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/deferrals/recognize", () => {
      recognizeCalled = true;
      return HttpResponse.json({ success: true, message: "OK.", data: { posted: 2 } });
    }),
  );
}

beforeEach(() => {
  recognizeCalled = false;
  seedRunning();
});

describe("DeferralDetail", () => {
  it("renders the seeded schedule overview", async () => {
    renderWithProviders(<DeferralDetail orgId="1" deferralId="7" />);

    expect((await screen.findAllByText("invoice #12")).length).toBeGreaterThan(0);
    expect(screen.getByText("Total amount")).toBeInTheDocument();
    expect(screen.getByText("1,200")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Recognize" })).toBeInTheDocument();
  });

  it("recognizes on button click and shows seeded lines", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DeferralDetail orgId="1" deferralId="7" />);

    await screen.findAllByText("invoice #12");
    await user.click(screen.getByRole("button", { name: "Recognize" }));
    await waitFor(() => expect(recognizeCalled).toBe(true));

    await user.click(screen.getByRole("tab", { name: "Recognition lines" }));

    expect(await screen.findByText("✓")).toBeInTheDocument();
    expect(screen.getByText("#44")).toBeInTheDocument();
  });

  it("hides recognize and shows empty lines for a done schedule", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/deferrals/:id", () =>
        HttpResponse.json({
          success: true,
          message: "OK.",
          data: { schedule: { ...schedule, state: "done" } },
        }),
      ),
      http.get("*/api/v1/organizations/:organizationId/deferrals/:id/lines", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { lines: [] } }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<DeferralDetail orgId="1" deferralId="7" />);

    await screen.findAllByText("invoice #12");
    expect(screen.queryByRole("button", { name: "Recognize" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("tab", { name: "Recognition lines" }));

    expect(await screen.findByText("No recognition lines.")).toBeInTheDocument();
  });

  it("shows notFound when the schedule is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/deferrals/:id", () =>
        HttpResponse.json({ success: true, message: "OK.", data: {} }),
      ),
    );
    renderWithProviders(<DeferralDetail orgId="1" deferralId="7" />);

    expect(await screen.findByText("Deferral schedule not found.")).toBeInTheDocument();
  });
});
