import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ServiceOrderDetail } from "../service-order-detail-section";

function order(patch: Record<string, unknown> = {}) {
  return {
    id: 1,
    organization_id: 1,
    name: "AC Repair Visit",
    contact_id: null,
    equipment_id: null,
    contract_id: null,
    type: "repair",
    priority: 1,
    state: "new",
    scheduled_date: null,
    technician_id: null,
    invoice_id: null,
    dimension_id: null,
    reported_issue: "AC not cooling",
    resolution: "",
    ...patch,
  };
}

const lines = [
  {
    id: 1,
    service_order_id: 1,
    type: "labor",
    item_id: null,
    description: "Brake pad replacement",
    qty: 2,
    unit_id: null,
    unit_cost: 10,
    unit_price: 25,
    stock_movement_id: null,
    billable: true,
    covered_by_warranty: false,
  },
];

function seedOrder(orderData: unknown, orderLines: unknown[] = []) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/service-orders/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { service_order: orderData } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/service-orders/:id/lines", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { service_order_lines: orderLines },
      }),
    ),
  );
}

beforeEach(() => {
  seedOrder(order(), lines);
});

describe("ServiceOrderDetail extra2", () => {
  it("shows the not-found state for a missing order", async () => {
    seedOrder(null, []);
    renderWithProviders(<ServiceOrderDetail orgId="1" orderId="1" />);

    expect(await screen.findByText("Pesanan tidak ditemukan.")).toBeInTheDocument();
  });

  it("schedules a new order", async () => {
    let scheduleCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/service-orders/:id/schedule", () => {
        scheduleCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ServiceOrderDetail orgId="1" orderId="1" />);

    await screen.findByRole("heading", { name: "AC Repair Visit" });
    await user.click(screen.getByRole("button", { name: "Jadwalkan" }));

    await waitFor(() => expect(scheduleCalls).toBe(1));
  });

  it("starts a scheduled order", async () => {
    let startCalls = 0;
    seedOrder(order({ state: "scheduled" }), []);
    server.use(
      http.post("*/api/v1/organizations/:organizationId/service-orders/:id/start", () => {
        startCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ServiceOrderDetail orgId="1" orderId="1" />);

    await screen.findByRole("heading", { name: "AC Repair Visit" });
    expect(screen.queryByRole("button", { name: "Jadwalkan" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Mulai" }));

    await waitFor(() => expect(startCalls).toBe(1));
  });

  it("completes an in-progress order", async () => {
    let completeCalls = 0;
    seedOrder(order({ state: "in_progress" }), []);
    server.use(
      http.post("*/api/v1/organizations/:organizationId/service-orders/:id/complete", () => {
        completeCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ServiceOrderDetail orgId="1" orderId="1" />);

    await screen.findByRole("heading", { name: "AC Repair Visit" });
    await user.click(screen.getByRole("button", { name: "Selesaikan" }));

    await waitFor(() => expect(completeCalls).toBe(1));
  });

  it("bills a done order", async () => {
    let billCalls = 0;
    seedOrder(order({ state: "done" }), []);
    server.use(
      http.post("*/api/v1/organizations/:organizationId/service-orders/:id/bill", () => {
        billCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ServiceOrderDetail orgId="1" orderId="1" />);

    await screen.findByRole("heading", { name: "AC Repair Visit" });
    expect(screen.queryByRole("button", { name: "Selesaikan" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Tagih" }));

    await waitFor(() => expect(billCalls).toBe(1));
  });

  it("cancels a new order", async () => {
    let cancelCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/service-orders/:id/cancel", () => {
        cancelCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ServiceOrderDetail orgId="1" orderId="1" />);

    await screen.findByRole("heading", { name: "AC Repair Visit" });
    await user.click(screen.getByRole("button", { name: "Batal" }));

    await waitFor(() => expect(cancelCalls).toBe(1));
  });

  it("shows the empty lines state", async () => {
    seedOrder(order(), []);
    const user = userEvent.setup();
    renderWithProviders(<ServiceOrderDetail orgId="1" orderId="1" />);

    await screen.findByRole("heading", { name: "AC Repair Visit" });
    await user.click(screen.getByRole("tab", { name: "Baris" }));

    expect(await screen.findByText("Belum ada baris")).toBeInTheDocument();
  });

  it("falls back to dash for missing references and resolution", async () => {
    seedOrder(
      order({ contact_id: 7, equipment_id: 8, contract_id: 9, technician_id: 3, resolution: null }),
      [],
    );
    renderWithProviders(<ServiceOrderDetail orgId="1" orderId="1" />);

    await screen.findByRole("heading", { name: "AC Repair Visit" });
    expect(screen.getByText("#7")).toBeInTheDocument();
    expect(screen.getByText("#8")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });
});
