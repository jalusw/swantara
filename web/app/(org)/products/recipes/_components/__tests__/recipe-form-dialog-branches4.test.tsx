import { screen, waitFor, within } from "@testing-library/react";
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

const created: unknown[] = [];
const updated: unknown[] = [];

function seed() {
  created.length = 0;
  updated.length = 0;
  server.use(
    http.get("*/api/v1/units", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          units: [
            { id: 1, name: "Unit", category_id: 1 },
            { id: 2, name: "Kg", category_id: 1 },
          ],
        },
      }),
    ),
    http.post("*/api/v1/organizations/:organizationId/recipes", async ({ request }) => {
      created.push(await request.json());
      return HttpResponse.json({ success: true, message: "OK.", data: {} });
    }),
    http.put("*/api/v1/organizations/:organizationId/recipes/:id", async ({ request }) => {
      updated.push(await request.json());
      return HttpResponse.json({ success: true, message: "OK.", data: {} });
    }),
  );
}

beforeEach(() => {
  seed();
});

async function selectComboboxOption(
  user: ReturnType<typeof userEvent.setup>,
  combobox: HTMLElement,
  optionName: string,
) {
  await user.click(combobox);
  await user.click(await screen.findByRole("option", { name: optionName }));
}

describe("BomFormDialog branches4", () => {
  it("submits a create with empty optionals mapped to null", async () => {
    const user = userEvent.setup();
    let saved = false;
    renderWithProviders(
      <BomFormDialog
        open
        onOpenChange={() => undefined}
        orgId="abc"
        templates={templates}
        onSave={() => {
          saved = true;
        }}
      />,
    );

    await screen.findByText("Resep baru");
    const dialog = screen.getByRole("dialog");
    const boxes = within(dialog).getAllByRole("combobox");
    const productBox = boxes[0];
    const lineComponentBox = boxes[boxes.length - 1];
    if (!productBox || !lineComponentBox) throw new Error("Expected selects");
    await selectComboboxOption(user, productBox, "Chair");
    await selectComboboxOption(user, lineComponentBox, "Leg");

    const spins = within(dialog).getAllByRole("spinbutton");
    const lineQty = spins[spins.length - 2];
    if (!lineQty) throw new Error("Expected line qty input");
    await user.type(lineQty, "2");

    await user.click(within(dialog).getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(saved).toBe(true));
    expect(created.length).toBe(1);
    const body = created[0] as Record<string, unknown>;
    expect(body.organization_id).toBe(0);
    expect(body.code).toBeNull();
    expect(body.unit_id).toBeNull();
    const lines = body.lines as Record<string, unknown>[];
    expect(lines[0]?.unit_id).toBeNull();
  });

  it("submits a create with unit and cleared scrap", async () => {
    const user = userEvent.setup();
    let saved = false;
    renderWithProviders(
      <BomFormDialog
        open
        onOpenChange={() => undefined}
        orgId="1"
        templates={templates}
        onSave={() => {
          saved = true;
        }}
      />,
    );

    await screen.findByText("Resep baru");
    const dialog = screen.getByRole("dialog");
    const boxes = within(dialog).getAllByRole("combobox");
    const productBox = boxes[0];
    const uomBox = boxes[2];
    const lineComponentBox = boxes[boxes.length - 1];
    if (!productBox || !uomBox || !lineComponentBox) throw new Error("Expected selects");
    await selectComboboxOption(user, productBox, "Chair");
    await selectComboboxOption(user, uomBox, "Unit");
    await selectComboboxOption(user, lineComponentBox, "Leg");

    const spins = within(dialog).getAllByRole("spinbutton");
    const lineQty = spins[spins.length - 2];
    const scrap = spins[spins.length - 1];
    if (!lineQty || !scrap) throw new Error("Expected line inputs");
    await user.type(lineQty, "4");
    await user.clear(scrap);

    await user.click(within(dialog).getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(saved).toBe(true));
    const body = created[created.length - 1] as Record<string, unknown>;
    expect(body.unit_id).toBe(1);
    const lines = body.lines as Record<string, unknown>[];
    expect(lines[0]?.scrap_pct).toBe(0);
  });

  it("submits an update with fallback ids and a chosen unit", async () => {
    const user = userEvent.setup();
    let saved = false;
    renderWithProviders(
      <BomFormDialog
        open
        onOpenChange={() => undefined}
        orgId="abc"
        templates={templates}
        initial={{
          id: "0",
          itemId: "1",
          code: null,
          type: "manufacture",
          qty: 2,
          unitId: null,
          version: 2,
          active: true,
        }}
        initialLines={[{ id: "l1", componentId: "2", qty: 4, unitId: null, scrapPct: 10 }]}
        onSave={() => {
          saved = true;
        }}
      />,
    );

    await screen.findByText("Ubah resep");
    const dialog = screen.getByRole("dialog");
    const boxes = within(dialog).getAllByRole("combobox");
    const uomBox = boxes[2];
    if (!uomBox) throw new Error("Expected unit select");
    await selectComboboxOption(user, uomBox, "Kg");

    await user.click(within(dialog).getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(saved).toBe(true));
    expect(updated.length).toBe(1);
    const body = updated[0] as Record<string, unknown>;
    expect(body.organization_id).toBe(0);
    expect(body.code).toBeNull();
    expect(body.unit_id).toBe(2);
  });
});
