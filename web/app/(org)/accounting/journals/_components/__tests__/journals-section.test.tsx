import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { JournalsSection } from "../journals-section";

function useLocalJournals() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          journals: [
            {
              id: 1,
              organization_id: 1,
              name: "Sales Journal",
              code: "SAJ",
              type: "sale",
              default_account_id: 1,
              bank_account_id: null,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
            {
              id: 2,
              organization_id: 1,
              name: "Bank Journal",
              code: "BNK",
              type: "bank",
              default_account_id: 2,
              bank_account_id: null,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/accounts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          accounts: [
            {
              id: 1,
              organization_id: 1,
              code: "1000",
              name: "Kas",
              type: "cash",
              reconcilable: false,
              currency_code: null,
              parent_id: null,
              active: true,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
            {
              id: 2,
              organization_id: 1,
              code: "1100",
              name: "Bank",
              type: "bank",
              reconcilable: true,
              currency_code: null,
              parent_id: null,
              active: true,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalJournals();
});

describe("JournalsSection", () => {
  it("renders seeded journals", async () => {
    renderWithProviders(<JournalsSection orgId="1" />);

    expect(await screen.findByText("Sales Journal")).toBeInTheDocument();
    expect(screen.getByText("Bank Journal")).toBeInTheDocument();
  });

  it("filters journals by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<JournalsSection orgId="1" />);

    await screen.findByText("Sales Journal");
    await user.type(screen.getByPlaceholderText("Cari jurnal..."), "Bank");

    expect(await screen.findByText("Bank Journal")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<JournalsSection orgId="1" />);

    await screen.findByText("Sales Journal");
    await user.click(screen.getByRole("button", { name: "Tambah jurnal" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Buat jurnal")).toBeInTheDocument();
  });
});
