import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { MOComponent } from "@/lib/services/swantara";
import { renderWithProviders, server } from "@/lib/tests";
import { ConsumeDialog } from "../consume-dialog";

const components: MOComponent[] = [
  {
    id: 9,
    productionOrderId: 1,
    itemId: 3,
    qtyPlanned: 10,
    qtyConsumed: 2,
    unitId: null,
    stockMovementId: null,
    createdAt: new Date("2026-01-01T00:00:00Z"),
    updatedAt: new Date("2026-01-01T00:00:00Z"),
  },
];

beforeEach(() => {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/production-orders/:id/consume", () =>
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
            qty_produced: 20,
            state: "in_progress",
          },
        },
      }),
    ),
  );
});

describe("ConsumeDialog", () => {
  it("renders consume title with component lines", async () => {
    renderWithProviders(
      <ConsumeDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        productionOrderId="1"
        components={components}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Konsumsi bahan")).toBeInTheDocument();
    expect(screen.getByText("#3")).toBeInTheDocument();
  });

  it("submits the consumption lines", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <ConsumeDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        productionOrderId="1"
        components={components}
        onSave={onSave}
      />,
    );

    await screen.findByText("Konsumsi bahan");
    await user.click(screen.getByRole("button", { name: "Konsumsi" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });
});
