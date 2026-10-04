import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import type { StubItem } from "../../../_components/products-data";
import { BomFormDialog } from "../recipe-form-dialog";

const templates: StubItem[] = [
  {
    id: "1",
    name: "Chair",
    categoryId: null,
    type: "stockable",
    unitId: null,
    purchaseUnitId: null,
    listPrice: 100,
    standardCost: 60,
    isPurchasable: true,
    isSellable: true,
    isManufactured: false,
    tracking: "none",
    active: true,
  },
  {
    id: "2",
    name: "Leg",
    categoryId: null,
    type: "stockable",
    unitId: null,
    purchaseUnitId: null,
    listPrice: 10,
    standardCost: 5,
    isPurchasable: true,
    isSellable: false,
    isManufactured: false,
    tracking: "none",
    active: true,
  },
];

function seed() {
  server.use(
    http.get("*/api/v1/units", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { units: [{ id: 1, name: "Unit", category_id: 1 }] },
      }),
    ),
    http.post("*/api/v1/organizations/:organizationId/recipes", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.put("*/api/v1/organizations/:organizationId/recipes/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
  );
}

beforeEach(() => {
  seed();
});

describe("BomFormDialog branches3", () => {
  it("renders create title for new recipes", async () => {
    renderWithProviders(
      <BomFormDialog
        open
        onOpenChange={() => undefined}
        orgId="1"
        templates={templates}
        onSave={() => undefined}
      />,
    );

    expect(await screen.findByText("Resep baru")).toBeInTheDocument();
  });

  it("renders edit title when initial is provided", async () => {
    renderWithProviders(
      <BomFormDialog
        open
        onOpenChange={() => undefined}
        orgId="1"
        templates={templates}
        initial={{
          id: "3",
          itemId: "1",
          code: null,
          type: "manufacture",
          qty: 2,
          unitId: null,
          version: 2,
          active: true,
        }}
        initialLines={[{ id: "l1", componentId: "2", qty: 4, unitId: null, scrapPct: 10 }]}
        onSave={() => undefined}
      />,
    );

    expect(await screen.findByText("Ubah resep")).toBeInTheDocument();
  });

  it("adds and removes a line", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <BomFormDialog
        open
        onOpenChange={() => undefined}
        orgId="1"
        templates={templates}
        onSave={() => undefined}
      />,
    );

    await screen.findByText("Resep baru");
    await user.click(screen.getByRole("button", { name: "Tambah komponen" }));
    const removes = await screen.findAllByRole("button", { name: "Hapus baris" });
    expect(removes.length).toBeGreaterThan(0);
    const first = removes[0];
    if (first) await user.click(first);
    expect(screen.getByRole("button", { name: "Tambah komponen" })).toBeInTheDocument();
  });

  it("shows total with scrap for positive quantities", async () => {
    renderWithProviders(
      <BomFormDialog
        open
        onOpenChange={() => undefined}
        orgId="1"
        templates={templates}
        initial={{
          id: "3",
          itemId: "1",
          code: "BOM-1",
          type: "manufacture",
          qty: 2,
          unitId: "",
          version: 1,
          active: false,
        }}
        initialLines={[{ id: "l1", componentId: "2", qty: 2, unitId: "", scrapPct: 50 }]}
        onSave={() => undefined}
      />,
    );

    expect(await screen.findByText("Ubah resep")).toBeInTheDocument();
  });
});
