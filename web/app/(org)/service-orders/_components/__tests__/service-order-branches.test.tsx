import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ServiceOrderFormDialog } from "../service-order-form-dialog";

const contacts = [{ id: 61, name: "Acme Corp" }];
const equipments = [{ id: 71, name: "Forklift A" }];
const contracts = [{ id: 81, name: "Gold SLA" }];
const employees = [{ id: 91, employee_number: "EMP-0091" }];
const products = [{ id: 5, name: "Finished Widget" }];

function seedLists() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/equipments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { equipments } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/service-contracts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { service_contracts: contracts } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/employees", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { employees } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
  );
}

function seedCreate(status = 200) {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/service-orders", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { service_order: { id: 9 } },
      });
    }),
  );
}

async function fillRequired(user: ReturnType<typeof userEvent.setup>, name: string) {
  await user.type(screen.getByLabelText("Nama"), name);
  await user.click(screen.getByRole("combobox", { name: "Jenis pesanan" }));
  await user.click(await screen.findByRole("option", { name: "Perbaikan" }));
}

beforeEach(() => {
  seedLists();
});

describe("ServiceOrderFormDialog branches", () => {
  it("blocks submit when name and type are missing", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ServiceOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Pesanan baru");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    expect(await screen.findByText("Nama wajib diisi.")).toBeInTheDocument();
    expect(await screen.findByText("Jenis wajib dipilih.")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("creates an order with only required fields", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ServiceOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Pesanan baru");
    await fillRequired(user, "Fix forklift");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("9"));
  });

  it("creates an order with customer, equipment, contract, and technician", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ServiceOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Pesanan baru");
    await fillRequired(user, "Quarterly maintenance");
    await user.click(screen.getByRole("combobox", { name: "Pelanggan" }));
    await user.click(await screen.findByRole("option", { name: "Acme Corp" }));
    await user.click(screen.getByRole("combobox", { name: "Peralatan" }));
    await user.click(await screen.findByRole("option", { name: "Forklift A" }));
    await user.click(screen.getByRole("combobox", { name: "Kontrak" }));
    await user.click(await screen.findByRole("option", { name: "Gold SLA" }));
    await user.click(screen.getByRole("combobox", { name: "Technician" }));
    await user.click(await screen.findByRole("option", { name: "EMP-0091" }));
    await user.type(screen.getByLabelText("Priority"), "3");
    await user.type(screen.getByLabelText("Masalah yang dilaporkan"), "Engine overheats");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("9"));
  });

  it("creates an order with a part line and item", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ServiceOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Pesanan baru");
    await fillRequired(user, "Replace filter");
    await user.click(screen.getByRole("combobox", { name: "Jenis baris" }));
    await user.click(await screen.findByRole("option", { name: "Suku Cadang" }));
    await user.click(screen.getByRole("combobox", { name: "Item" }));
    await user.click(await screen.findByRole("option", { name: "Finished Widget" }));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("9"));
  });

  it("adds and removes service lines", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <ServiceOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Pesanan baru");
    expect(screen.getByRole("button", { name: "x" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "Tambah baris" }));

    const removeButtons = screen.getAllByRole("button", { name: "x" });
    expect(removeButtons).toHaveLength(2);
    const firstRemove = removeButtons[0];
    const secondRemove = removeButtons[1];
    if (firstRemove === undefined || secondRemove === undefined)
      throw new Error("expected remove buttons");
    expect(firstRemove).toBeEnabled();
    await user.click(secondRemove);
    expect(screen.getAllByRole("button", { name: "x" })).toHaveLength(1);
  });

  it("does not save when creation fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate(500);
    renderWithProviders(
      <ServiceOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Pesanan baru");
    await fillRequired(user, "Fix forklift");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(screen.getByText("Pesanan baru")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("closes without saving on cancel", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <ServiceOrderFormDialog open={true} onOpenChange={onOpenChange} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Pesanan baru");
    await user.click(screen.getByRole("button", { name: "Batal" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
