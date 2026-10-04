import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PosRegister } from "../pos-register-section";

const session = {
  id: 1,
  config_id: 5,
  cashier_id: 7,
  opened_at: "2026-01-01T08:00:00Z",
  closed_at: null,
  opening_balance: 100000,
  closing_balance: null,
  state: "opened",
};

const products = [
  { id: 1, organization_id: 1, name: "Arabica Beans", type: "stockable", list_price: 50000 },
  { id: 2, organization_id: 1, name: "Paper Cups", type: "consumable", list_price: 15000 },
];

function usePosRegisterHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/pos/sessions/:sessionId", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { session } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
  );
}

beforeEach(() => {
  usePosRegisterHandlers();
});

describe("PosRegister", () => {
  it("renders products with session label", async () => {
    renderWithProviders(<PosRegister orgId="1" sessionId="1" />);

    expect(await screen.findByText("Arabica Beans")).toBeInTheDocument();
    expect(screen.getByText("Paper Cups")).toBeInTheDocument();
    expect(screen.getByText(/Sesi: Sesi 1/)).toBeInTheDocument();
  });

  it("adds a item to the cart", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PosRegister orgId="1" sessionId="1" />);

    await screen.findByText("Arabica Beans");
    await user.click(screen.getByRole("button", { name: /Arabica Beans/ }));

    expect(await screen.findByText("1 item")).toBeInTheDocument();
  });
});
