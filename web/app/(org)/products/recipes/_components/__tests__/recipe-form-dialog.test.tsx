import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import type { StubItem } from "../../../_components/products-data";
import { BomFormDialog } from "../recipe-form-dialog";

const STAMP = "2026-01-01T00:00:00Z";

const templates: StubItem[] = [
  {
    id: "5",
    name: "Finished Widget",
    categoryId: null,
    type: "stockable",
    unitId: null,
    purchaseUnitId: null,
    listPrice: 100,
    standardCost: 60,
    isPurchasable: true,
    isSellable: true,
    isManufactured: true,
    tracking: "none",
    active: true,
  },
  {
    id: "6",
    name: "Raw Material",
    categoryId: null,
    type: "consumable",
    unitId: null,
    purchaseUnitId: null,
    listPrice: 10,
    standardCost: 8,
    isPurchasable: true,
    isSellable: false,
    isManufactured: false,
    tracking: "none",
    active: true,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/units", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          units: [
            {
              id: 1,
              category_id: 1,
              name: "Units",
              factor: 1,
              unit_type: "reference",
              rounding: 0.01,
              created_at: STAMP,
              updated_at: STAMP,
            },
          ],
        },
      }),
    ),
  );
});

describe("BomFormDialog", () => {
  it("renders the create form when open", () => {
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        onSave={vi.fn()}
      />,
    );

    expect(screen.getByText("New bill of materials")).toBeInTheDocument();
    expect(screen.getByText("Components")).toBeInTheDocument();
  });

  it("requires a item before saving", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <BomFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        templates={templates}
        onSave={onSave}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Select a item.")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });
});
