import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PosConfigsSection } from "../pos-configs-section";

const configs = [
  {
    id: 1,
    organization_id: 1,
    name: "Main Store",
    warehouse_id: 1,
    journal_id: 1,
    price_book_id: 1,
  },
  {
    id: 2,
    organization_id: 1,
    name: "Outlet Kiosk",
    warehouse_id: 2,
    journal_id: 1,
    price_book_id: null,
  },
];

function useConfigHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/pos/configs", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { configs } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          warehouses: [
            { id: 1, organization_id: 1, name: "Central Warehouse" },
            { id: 2, organization_id: 1, name: "Kiosk Storage" },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { journals: [{ id: 1, organization_id: 1, name: "Cash Journal" }] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/price_books", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { price_books: [{ id: 1, organization_id: 1, name: "Retail PriceBook" }] },
      }),
    ),
  );
}

beforeEach(() => {
  useConfigHandlers();
});

describe("PosConfigsSection", () => {
  it("renders configs with resolved warehouse and journal names", async () => {
    renderWithProviders(<PosConfigsSection orgId="1" />);

    expect(await screen.findByText("Main Store")).toBeInTheDocument();
    expect(screen.getByText("Outlet Kiosk")).toBeInTheDocument();
    expect(screen.getByText("Central Warehouse")).toBeInTheDocument();
    expect(screen.getAllByText("Cash Journal").length).toBe(2);
    expect(screen.getByText("Retail PriceBook")).toBeInTheDocument();
  });

  it("filters configs through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PosConfigsSection orgId="1" />);

    await screen.findByText("Main Store");
    await user.type(screen.getByPlaceholderText("Cari konfigurasi…"), "Kiosk");

    expect(await screen.findByText("Outlet Kiosk")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Main Store")).not.toBeInTheDocument());
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PosConfigsSection orgId="1" />);

    await screen.findByText("Main Store");
    await user.click(screen.getByRole("button", { name: "Konfigurasi baru" }));

    expect(await screen.findByRole("heading", { name: "Konfigurasi baru" })).toBeInTheDocument();
  });
});
