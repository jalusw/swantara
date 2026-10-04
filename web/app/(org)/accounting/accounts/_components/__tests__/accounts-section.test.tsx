import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { AccountsSection } from "../accounts-section";

function useLocalAccounts() {
  server.use(
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
              name: "Cash on Hand",
              type: "cash",
              reconcilable: true,
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
              name: "Bank BCA",
              type: "bank",
              reconcilable: true,
              currency_code: null,
              parent_id: null,
              active: true,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
            {
              id: 3,
              organization_id: 1,
              code: "2000",
              name: "Accounts Payable",
              type: "payable",
              reconcilable: false,
              currency_code: null,
              parent_id: null,
              active: false,
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
  useLocalAccounts();
});

describe("AccountsSection", () => {
  it("renders seeded accounts with codes", async () => {
    renderWithProviders(<AccountsSection orgId="1" />);

    expect(await screen.findByText("Cash on Hand")).toBeInTheDocument();
    expect(screen.getByText("Bank BCA")).toBeInTheDocument();
    expect(screen.getByText("1000")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<AccountsSection orgId="1" />);

    await screen.findByText("Cash on Hand");
    await user.click(screen.getByRole("button", { name: "Tambah akun" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Buat akun")).toBeInTheDocument();
  });

  it("opens the edit dialog from a row action", async () => {
    const user = userEvent.setup();
    renderWithProviders(<AccountsSection orgId="1" />);

    await screen.findByText("Cash on Hand");
    const editButton = screen.getAllByRole("button", { name: "Ubah" })[0];
    if (editButton === undefined) {
      throw new Error("Expected an Edit button");
    }
    await user.click(editButton);

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Ubah akun")).toBeInTheDocument();
  });
});
