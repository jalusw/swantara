import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PurchaseRequestFormDialog } from "../purchase-request-form-dialog";

const STAMP = "2026-01-01T00:00:00Z";

const requester = {
  id: 7,
  organization_id: 1,
  name: "Aria Chen",
  display_name: "Aria Chen",
  is_organization: false,
  parent_id: null,
  email: null,
  phone: null,
  mobile: null,
  website: null,
  tax_id: null,
  industry: null,
  currency_code: null,
  lang: "en",
  active: true,
  created_at: STAMP,
  updated_at: STAMP,
};

const department = {
  id: 3,
  organization_id: 1,
  name: "Engineering",
  description: null,
  parent_id: null,
  manager_id: null,
  dimension_id: null,
};

const item = {
  id: 5,
  organization_id: 1,
  name: "Finished Widget",
  category_id: null,
  type: "stockable",
  unit_id: null,
  purchase_unit_id: null,
  list_price: 100,
  standard_cost: 60,
  is_purchasable: true,
  is_sellable: true,
  is_manufactured: false,
  tracking: "none",
  weight: 0,
  volume: 0,
  hs_code: null,
  description_sale: null,
  description_purchase: null,
  active: true,
  created_at: STAMP,
  updated_at: STAMP,
};

function seedForm() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [requester] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/departments", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { departments: [department] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [item] } }),
    ),
  );
}

async function fillRequired(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("combobox", { name: "Requester" }));
  await user.click(await screen.findByRole("option", { name: "Aria Chen" }));
  await user.click(screen.getByRole("combobox", { name: "Item 1" }));
  await user.click(await screen.findByRole("option", { name: "Finished Widget" }));
}

beforeEach(() => {
  seedForm();
});

describe("PurchaseRequestFormDialog extra2", () => {
  it("creates a request and notifies on save", async () => {
    const createBodies: unknown[] = [];
    server.use(
      http.post("*/api/v1/organizations/:organizationId/purchase-requests", async ({ request }) => {
        createBodies.push(await request.json());
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <PurchaseRequestFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await fillRequired(user);
    await user.click(screen.getByRole("combobox", { name: "Departemen" }));
    await user.click(await screen.findByRole("option", { name: "Engineering" }));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(createBodies).toHaveLength(1));
    expect(createBodies[0]).toMatchObject({ requester_id: 7, department_id: 3 });
    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("blocks saving when every line was removed", async () => {
    let createCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/purchase-requests", () => {
        createCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <PurchaseRequestFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await user.click(screen.getByRole("combobox", { name: "Requester" }));
    await user.click(await screen.findByRole("option", { name: "Aria Chen" }));
    await user.click(screen.getByRole("button", { name: "Hapus" }));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    expect(await screen.findByText("Tambahkan minimal satu baris.")).toBeInTheDocument();
    expect(createCalls).toBe(0);
    expect(onSave).not.toHaveBeenCalled();
  });

  it("blocks saving when a line quantity is not positive", async () => {
    let createCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/purchase-requests", () => {
        createCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <PurchaseRequestFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await fillRequired(user);
    const qtyInput = screen.getAllByRole("spinbutton")[0]!;
    await user.clear(qtyInput);
    await user.type(qtyInput, "0");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    expect(createCalls).toBe(0);
    expect(onSave).not.toHaveBeenCalled();
  });

  it("adds and removes request lines", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <PurchaseRequestFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    await user.click(screen.getByRole("button", { name: "Tambah baris" }));
    expect(screen.getByRole("combobox", { name: "Item 2" })).toBeInTheDocument();
    await user.click(screen.getAllByRole("button", { name: "Hapus" })[0]!);
    expect(screen.queryByRole("combobox", { name: "Item 2" })).not.toBeInTheDocument();
  });

  it("updates quantity, needed-by and description of a line", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <PurchaseRequestFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    const qtyInput = screen.getAllByRole("spinbutton")[0]!;
    await user.clear(qtyInput);
    await user.type(qtyInput, "4");
    expect(qtyInput).toHaveValue(4);

    const descriptionInput = screen.getByPlaceholderText("Deskripsi");
    await user.type(descriptionInput, "Spare parts");
    expect(descriptionInput).toHaveValue("Spare parts");
  });

  it("updates the header needed-by date", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <PurchaseRequestFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    const neededByInput = screen.getByLabelText("Dibutuhkan pada");
    await user.type(neededByInput, "2026-05-01");
    expect(neededByInput).toHaveValue("2026-05-01");
  });
});
