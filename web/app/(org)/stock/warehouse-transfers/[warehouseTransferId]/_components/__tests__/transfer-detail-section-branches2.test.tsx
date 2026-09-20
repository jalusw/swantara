import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { WarehouseTransferDetail } from "../warehouse-transfer-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

function transferOrder(overrides: Record<string, unknown> = {}) {
  return {
    id: 31,
    organization_id: 1,
    name: "TO-0031",
    src_warehouse_id: 1,
    dst_warehouse_id: 2,
    state: "draft",
    out_shipment_id: null,
    in_shipment_id: null,
    is_interorganization: false,
    scheduled_date: "2026-02-01",
    created_at: STAMP,
    updated_at: STAMP,
    ...overrides,
  };
}

function seedTransfer(order: unknown) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/warehouse-transfers/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { warehouseTransfer: order } }),
    ),
  );
}

beforeEach(() => {});

describe("WarehouseTransferDetail branches2", () => {
  it("renders the not-found branch for unknown transfers", async () => {
    seedTransfer(null);
    renderWithProviders(<WarehouseTransferDetail orgId="1" warehouseTransferId="99" />);

    expect(await screen.findByText("Transfer not found.")).toBeInTheDocument();
  });

  it("shows out and in shipment branches when linked", async () => {
    seedTransfer(transferOrder({ out_shipment_id: 11, in_shipment_id: 12, state: "in_transit" }));
    renderWithProviders(<WarehouseTransferDetail orgId="1" warehouseTransferId="31" />);

    await screen.findAllByText("TO-0031");
    expect(screen.getByText("PK-11")).toBeInTheDocument();
    expect(screen.getByText("PK-12")).toBeInTheDocument();
  });

  it("receives an in-transit transfer", async () => {
    let receiveCalls = 0;
    seedTransfer(transferOrder({ state: "in_transit" }));
    server.use(
      http.post("*/api/v1/organizations/:organizationId/warehouse-transfers/:id/receive", () => {
        receiveCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<WarehouseTransferDetail orgId="1" warehouseTransferId="31" />);

    await screen.findAllByText("TO-0031");
    await user.click(screen.getByRole("button", { name: "Receive" }));

    await waitFor(() => expect(receiveCalls).toBe(1));
  });

  it("hides send for non-draft and shows name fallback", async () => {
    seedTransfer(transferOrder({ name: null, state: "received" }));
    renderWithProviders(<WarehouseTransferDetail orgId="1" warehouseTransferId="31" />);

    expect((await screen.findAllByText("TO-31")).length).toBeGreaterThan(0);
    expect(screen.queryByRole("button", { name: "Send" })).not.toBeInTheDocument();
  });
});
