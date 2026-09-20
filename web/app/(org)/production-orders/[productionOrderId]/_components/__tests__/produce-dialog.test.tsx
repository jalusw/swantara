import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ProductionOrder } from "@/lib/services/swantara";
import { renderWithProviders, server } from "@/lib/tests";
import { ProduceDialog } from "../produce-dialog";

const productionOrder: ProductionOrder = {
  id: 1,
  organizationId: 1,
  name: "MO-0001 Chair Assembly",
  itemId: 5,
  bomId: null,
  qtyToProduce: 100,
  qtyProduced: 20,
  unitId: null,
  srcLocationId: null,
  dstLocationId: null,
  state: "in_progress",
  datePlannedStart: null,
  datePlannedFinish: null,
  dateStart: null,
  dateFinished: null,
  origin: "manual",
  priority: 1,
  createdAt: new Date("2026-01-01T00:00:00Z"),
  updatedAt: new Date("2026-01-01T00:00:00Z"),
};

beforeEach(() => {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/production-orders/:id/produce", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          production_order: {
            id: 1,
            organization_id: 1,
            name: "MO-0001 Chair Assembly",
            item_id: 5,
            qty_to_produce: 100,
            qty_produced: 30,
            state: "in_progress",
          },
        },
      }),
    ),
  );
});

describe("ProduceDialog", () => {
  it("renders produce title with quantity field", async () => {
    renderWithProviders(
      <ProduceDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        productionOrderId="1"
        productionOrder={productionOrder}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Produce Goods")).toBeInTheDocument();
    expect(screen.getByLabelText("Quantity to Produce")).toBeInTheDocument();
  });

  it("submits the produced quantity", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <ProduceDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        productionOrderId="1"
        productionOrder={productionOrder}
        onSave={onSave}
      />,
    );

    await screen.findByText("Produce Goods");
    await user.click(screen.getByRole("button", { name: "Produce" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });
});
