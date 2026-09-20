import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { TaxesSection } from "../taxes-section";

function useLocalTaxes() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/taxes", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          taxes: [
            {
              id: 1,
              organization_id: 1,
              name: "PPN 11%",
              amount: 11,
              type: "percent",
              scope: "sale",
              price_include: false,
              tax_account_id: null,
              refund_tax_account_id: null,
              active: true,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
            {
              id: 2,
              organization_id: 1,
              name: "PPh 23",
              amount: 2,
              type: "percent",
              scope: "purchase",
              price_include: false,
              tax_account_id: null,
              refund_tax_account_id: null,
              active: true,
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
        data: { accounts: [] },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalTaxes();
});

describe("TaxesSection", () => {
  it("renders seeded taxes", async () => {
    renderWithProviders(<TaxesSection orgId="1" />);

    expect(await screen.findByText("PPN 11%")).toBeInTheDocument();
    expect(screen.getByText("PPh 23")).toBeInTheDocument();
  });

  it("filters taxes by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TaxesSection orgId="1" />);

    await screen.findByText("PPN 11%");
    await user.type(screen.getByPlaceholderText("Search taxes..."), "PPh");

    expect(await screen.findByText("PPh 23")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TaxesSection orgId="1" />);

    await screen.findByText("PPN 11%");
    await user.click(screen.getByRole("button", { name: "Add tax" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Create tax")).toBeInTheDocument();
  });
});
