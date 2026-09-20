import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { RmaDetailSection } from "../rma-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

const rma = {
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
};

const lines = [
  {
    id: 51,
    rma_id: 9,
    item_id: 5,
    qty: 2,
    batch_id: null,
    disposition: "restock",
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
  {
    id: 3,
    organization_id: 1,
    name: "General Journal",
    code: "GJ",
    type: "general",
    default_account_id: null,
    currency_code: "USD",
    bank_account_id: null,
    sequence_id: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

let confirmCalled = false;

beforeEach(() => {
  confirmCalled = false;
  server.use(
    http.get("*/api/v1/organizations/:organizationId/rmas/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { rma, lines } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/rmas/:id/confirm", () => {
      confirmCalled = true;
      return HttpResponse.json({ success: true, message: "OK.", data: { rma } });
    }),
  );
});

describe("RmaDetailSection", () => {
  it("renders the RMA summary with lines", async () => {
    renderWithProviders(<RmaDetailSection orgId="1" rmaId="9" />);

    expect(await screen.findByText("RMA-0009")).toBeInTheDocument();
    expect(screen.getByText("#5")).toBeInTheDocument();
  });

  it("confirms the draft RMA on button click", async () => {
    const user = userEvent.setup();
    renderWithProviders(<RmaDetailSection orgId="1" rmaId="9" />);

    await screen.findByText("RMA-0009");
    await user.click(screen.getByRole("button", { name: "Confirm" }));

    expect(confirmCalled).toBe(true);
  });
});
