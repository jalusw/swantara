import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import type { CommissionEntry } from "@/lib/services/swantara/types";
import { renderWithProviders, server } from "@/lib/tests";
import { CommissionEntriesSection } from "../commission-entries-section";

function entry(id: number, state: CommissionEntry["state"]): Record<string, unknown> {
  return {
    id,
    organization_id: 1,
    salesperson_id: 10 + id,
    plan_id: 3,
    source_type: "invoice",
    source_id: 100 + id,
    base_amount: 1000,
    commission_amount: 100 * id,
    state,
    period_id: null,
    payslip_id: null,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  };
}

function seedEntries(payload: unknown, status = 200) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/commission-entries", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Boom." }, { status });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: payload,
      });
    }),
  );
}

beforeEach(() => {});

describe("CommissionEntriesSection", () => {
  it("renders every state tone branch", async () => {
    seedEntries({
      commission_entries: [
        entry(1, "draft"),
        entry(2, "confirmed"),
        entry(3, "paid"),
        entry(4, "cancelled"),
      ],
    });
    renderWithProviders(<CommissionEntriesSection orgId="1" />);

    expect(await screen.findByText("#11")).toBeInTheDocument();
    expect(screen.getByText("#12")).toBeInTheDocument();
    expect(await screen.findByText("Draf")).toBeInTheDocument();
    expect(screen.getByText("Dikonfirmasi")).toBeInTheDocument();
    expect(screen.getByText("Lunas")).toBeInTheDocument();
    expect(screen.getByText("Dibatalkan")).toBeInTheDocument();
  });

  it("totals commission amounts across entries", async () => {
    seedEntries({ commission_entries: [entry(1, "confirmed"), entry(2, "paid")] });
    renderWithProviders(<CommissionEntriesSection orgId="1" />);

    await screen.findByText("#11");
    expect(screen.getByText("Rp 300,00")).toBeInTheDocument();
  });

  it("shows the empty state when no entries exist", async () => {
    seedEntries({ commission_entries: [] });
    renderWithProviders(<CommissionEntriesSection orgId="1" />);

    expect(await screen.findByText("Belum ada entri komisi")).toBeInTheDocument();
  });

  it("shows the error state with retry on load failure", async () => {
    seedEntries({}, 500);
    renderWithProviders(<CommissionEntriesSection orgId="1" />);

    expect(await screen.findByText("Boom.")).toBeInTheDocument();
  });

  it("falls back to empty list when the envelope omits entries", async () => {
    seedEntries({});
    renderWithProviders(<CommissionEntriesSection orgId="1" />);

    expect(await screen.findByText("Belum ada entri komisi")).toBeInTheDocument();
  });
});
