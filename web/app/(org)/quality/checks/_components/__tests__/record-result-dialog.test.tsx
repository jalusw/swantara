import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { QualityCheck } from "@/lib/services/swantara";
import { renderWithProviders, server } from "@/lib/tests";
import { RecordResultDialog } from "../record-result-dialog";

const check: QualityCheck = {
  id: 11,
  pointId: 1,
  itemId: 5,
  batchId: null,
  shipmentId: 7,
  productionOrderId: null,
  measuredValue: null,
  result: "pending",
  checkedBy: null,
  checkedAt: null,
};

let recordedPass: boolean | null = null;

beforeEach(() => {
  recordedPass = null;
  server.use(
    http.post("*/api/v1/organizations/:organizationId/quality-checks/:id/result", () => {
      recordedPass = false;
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          check: {
            id: 11,
            point_id: 1,
            item_id: 5,
            batch_id: null,
            shipment_id: 7,
            production_order_id: null,
            measured_value: null,
            result: "fail",
            checked_by: null,
            checked_at: null,
          },
        },
      });
    }),
  );
});

describe("RecordResultDialog", () => {
  it("renders the result options", async () => {
    renderWithProviders(
      <RecordResultDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        check={check}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Record Result")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Fail" })).toBeInTheDocument();
  });

  it("records a fail result", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <RecordResultDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        check={check}
        onSave={onSave}
      />,
    );

    await screen.findByText("Record Result");
    await user.click(screen.getByRole("button", { name: "Fail" }));
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
    expect(recordedPass).toBe(false);
  });
});
