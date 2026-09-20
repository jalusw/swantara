import { screen, waitFor, within } from "@testing-library/react";
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

function seedRegister(sessionData: unknown) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/pos/sessions/:sessionId", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { session: sessionData } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
  );
}

beforeEach(() => {
  seedRegister(session);
});

describe("PosRegister extra2", () => {
  it("keeps the pay button disabled with an empty cart", async () => {
    renderWithProviders(<PosRegister orgId="1" sessionId="1" />);

    await screen.findByText("Arabica Beans");
    expect(screen.getByRole("button", { name: /^Pay/ })).toBeDisabled();
    expect(screen.getByText("Cart is empty.")).toBeInTheDocument();
  });

  it("increments quantity when adding the same item twice", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PosRegister orgId="1" sessionId="1" />);

    await screen.findByText("Arabica Beans");
    await user.click(screen.getByRole("button", { name: /Arabica Beans/ }));
    await user.click(screen.getByRole("button", { name: /Arabica Beans/ }));

    expect(await screen.findByText("1 item(s)")).toBeInTheDocument();
  });

  it("clears the cart", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PosRegister orgId="1" sessionId="1" />);

    await screen.findByText("Arabica Beans");
    await user.click(screen.getByRole("button", { name: /Arabica Beans/ }));
    expect(await screen.findByText("1 item(s)")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Clear" }));
    expect(await screen.findByText("Cart is empty.")).toBeInTheDocument();
  });

  it("disables pay when the session is closed", async () => {
    seedRegister({ ...session, state: "closed" });
    const user = userEvent.setup();
    renderWithProviders(<PosRegister orgId="1" sessionId="1" />);

    await screen.findByText("Arabica Beans");
    await user.click(screen.getByRole("button", { name: /Arabica Beans/ }));
    expect(await screen.findByText("1 item(s)")).toBeInTheDocument();

    expect(screen.getByRole("button", { name: /^Pay/ })).toBeDisabled();
  });

  it("completes payment and shows the receipt", async () => {
    let orderCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/pos/orders", () => {
        orderCalls += 1;
        return HttpResponse.json(
          {
            success: true,
            message: "OK.",
            data: { order: { id: 7, name: "POS-0007", amount_total: 50000 } },
          },
          { status: 201 },
        );
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<PosRegister orgId="1" sessionId="1" />);

    await screen.findByText("Arabica Beans");
    await user.click(screen.getByRole("button", { name: /Arabica Beans/ }));
    await user.click(screen.getByRole("button", { name: /^Pay/ }));

    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Confirm" }));

    await waitFor(() => expect(orderCalls).toBe(1));
    expect(await screen.findByText("Receipt")).toBeInTheDocument();
    expect(screen.getByText("POS-0007")).toBeInTheDocument();
  });

  it("starts a new order from the receipt", async () => {
    server.use(
      http.post("*/api/v1/organizations/:organizationId/pos/orders", () =>
        HttpResponse.json(
          {
            success: true,
            message: "OK.",
            data: { order: { id: 7, name: "POS-0007", amount_total: 50000 } },
          },
          { status: 201 },
        ),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<PosRegister orgId="1" sessionId="1" />);

    await screen.findByText("Arabica Beans");
    await user.click(screen.getByRole("button", { name: /Arabica Beans/ }));
    await user.click(screen.getByRole("button", { name: /^Pay/ }));
    await user.click(
      within(await screen.findByRole("dialog")).getByRole("button", { name: "Confirm" }),
    );

    await screen.findByText("Receipt");
    await user.click(screen.getByRole("button", { name: "New order" }));

    expect(await screen.findByText("Arabica Beans")).toBeInTheDocument();
    expect(screen.getByText("Cart is empty.")).toBeInTheDocument();
  });

  it("blocks confirm when the payment does not cover the total", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PosRegister orgId="1" sessionId="1" />);

    await screen.findByText("Arabica Beans");
    await user.click(screen.getByRole("button", { name: /Arabica Beans/ }));
    await user.click(screen.getByRole("button", { name: /^Pay/ }));

    const dialog = await screen.findByRole("dialog");
    await user.clear(within(dialog).getByRole("spinbutton"));
    await user.type(within(dialog).getByRole("spinbutton"), "0");

    expect(within(dialog).getByRole("button", { name: "Confirm" })).toBeDisabled();
  });
});
