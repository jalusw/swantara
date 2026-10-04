import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { WarehouseTransferDetail } from "../warehouse-transfer-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

const order = {
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
};

let sendCalled = false;

beforeEach(() => {
  sendCalled = false;
  server.use(
    http.get("*/api/v1/organizations/:organizationId/warehouse-transfers/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { warehouseTransfer: order } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/warehouse-transfers/:id/send", () => {
      sendCalled = true;
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { warehouseTransfer: order },
      });
    }),
  );
});

describe("WarehouseTransferDetail", () => {
  it("renders the transfer with its legs", async () => {
    renderWithProviders(<WarehouseTransferDetail orgId="1" warehouseTransferId="31" />);

    expect((await screen.findAllByText("TO-0031")).length).toBeGreaterThan(0);
    expect(screen.getByText("Tahapan transfer")).toBeInTheDocument();
  });

  it("sends the draft transfer on button click", async () => {
    const user = userEvent.setup();
    renderWithProviders(<WarehouseTransferDetail orgId="1" warehouseTransferId="31" />);

    await screen.findAllByText("TO-0031");
    await user.click(screen.getByRole("button", { name: "Kirim" }));

    expect(sendCalled).toBe(true);
  });
});
