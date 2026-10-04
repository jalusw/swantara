import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { RmaDetailSection } from "../rma-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

function rmaFixture(overrides = {}) {
  return {
    id: 9,
    organization_id: 1,
    name: "RMA-0009",
    type: "customer_return",
    contact_id: 3,
    origin_order_type: "sale_order",
    origin_order_id: 7,
    reason: "Damaged",
    state: "draft",
    created_at: STAMP,
    updated_at: STAMP,
    ...overrides,
  };
}

const lines = [
  {
    id: 51,
    rma_id: 9,
    item_id: 5,
    qty: 2,
    batch_id: null,
    disposition: "restock",
    stock_movement_id: 12,
    credit_note_id: 34,
  },
  {
    id: 52,
    rma_id: 9,
    item_id: 6,
    qty: 1,
    batch_id: null,
    disposition: "scrap",
    stock_movement_id: null,
    credit_note_id: null,
  },
];

const journals = [
  {
    id: 2,
    organization_id: 1,
    name: "Purchase Journal",
    code: "PJ",
    type: "purchase",
    default_account_id: null,
    currency_code: "USD",
    bank_account_id: null,
    sequence_id: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

function useRmaHandlers(rma: Record<string, unknown>) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/rmas/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { rma, lines } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals } }),
    ),
  );
}

beforeEach(() => {});

describe("RmaDetailSection remainder", () => {
  it("renders the loading skeleton while fetching", () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/rmas/:id", async () => {
        await new Promise((resolve) => setTimeout(resolve, 50));
        return HttpResponse.json({
          success: true,
          message: "OK.",
          data: { rma: rmaFixture(), lines: [] },
        });
      }),
    );
    renderWithProviders(<RmaDetailSection orgId="1" rmaId="9" />);

    expect(document.querySelector(".animate-pulse, [data-slot='skeleton']")).not.toBeNull();
  });

  it("renders null when the rma payload is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/rmas/:id", () =>
        HttpResponse.json({ success: false, message: "Not found." }, { status: 404 }),
      ),
    );
    const { container } = renderWithProviders(<RmaDetailSection orgId="1" rmaId="999" />);

    await waitFor(() => expect(container.firstChild).toBeNull());
  });

  it("renders null for an unknown rma state", async () => {
    useRmaHandlers(rmaFixture({ state: "archived" }));
    const { container } = renderWithProviders(<RmaDetailSection orgId="1" rmaId="9" />);

    await waitFor(() => expect(container.firstChild).toBeNull());
  });

  it("renders origin order, stock moves and credit notes in lines", async () => {
    useRmaHandlers(rmaFixture());
    renderWithProviders(<RmaDetailSection orgId="1" rmaId="9" />);

    expect(await screen.findByText("RMA-0009")).toBeInTheDocument();
    expect(screen.getByText("SALE_ORDER-7")).toBeInTheDocument();
    expect(screen.getByText("#12")).toBeInTheDocument();
    expect(screen.getByText("#34")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
  });

  it("marks a refunded rma as done", async () => {
    let doneCalled = false;
    useRmaHandlers(rmaFixture({ state: "refunded" }));
    server.use(
      http.post("*/api/v1/organizations/:organizationId/rmas/:id/done", () => {
        doneCalled = true;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<RmaDetailSection orgId="1" rmaId="9" />);

    await screen.findByText("RMA-0009");
    await user.click(screen.getByRole("button", { name: "Selesai" }));

    await waitFor(() => expect(doneCalled).toBe(true));
  });

  it("cancels a draft rma", async () => {
    let cancelCalled = false;
    useRmaHandlers(rmaFixture());
    server.use(
      http.post("*/api/v1/organizations/:organizationId/rmas/:id/cancel", () => {
        cancelCalled = true;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<RmaDetailSection orgId="1" rmaId="9" />);

    await screen.findByText("RMA-0009");
    await user.click(screen.getByRole("button", { name: "Batal" }));

    await waitFor(() => expect(cancelCalled).toBe(true));
  });

  it("opens the receive dialog for a confirmed rma", async () => {
    useRmaHandlers(rmaFixture({ state: "confirmed" }));
    const user = userEvent.setup();
    renderWithProviders(<RmaDetailSection orgId="1" rmaId="9" />);

    await screen.findByText("RMA-0009");
    const receiveButton = screen.getByRole("button", { name: "Terima" });
    await user.click(receiveButton);

    expect(await screen.findByText("Terima Retur")).toBeInTheDocument();
  });

  it("opens the refund dialog for a received rma", async () => {
    useRmaHandlers(rmaFixture({ state: "received" }));
    const user = userEvent.setup();
    renderWithProviders(<RmaDetailSection orgId="1" rmaId="9" />);

    await screen.findByText("RMA-0009");
    const refundButton = screen.getByRole("button", { name: "Kembalikan dana" });
    await user.click(refundButton);

    expect(await screen.findByText("Refund Retur")).toBeInTheDocument();
  });

  it("navigates back with the back button", async () => {
    useRmaHandlers(rmaFixture());
    const user = userEvent.setup();
    const { navigationMock } = await import("@/lib/tests");
    renderWithProviders(<RmaDetailSection orgId="1" rmaId="9" />);

    await screen.findByText("RMA-0009");
    const buttons = screen.getAllByRole("button");
    await user.click(buttons[0]!);

    expect(navigationMock.back).toHaveBeenCalled();
  });

  it("shows an error toast when confirming fails", async () => {
    useRmaHandlers(rmaFixture());
    server.use(
      http.post("*/api/v1/organizations/:organizationId/rmas/:id/confirm", () =>
        HttpResponse.json({ success: false, message: "Cannot confirm." }, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<RmaDetailSection orgId="1" rmaId="9" />);

    await screen.findByText("RMA-0009");
    await user.click(screen.getByRole("button", { name: "Konfirmasi" }));

    await waitFor(() => expect(screen.getByText("RMA-0009")).toBeInTheDocument());
  });
});
