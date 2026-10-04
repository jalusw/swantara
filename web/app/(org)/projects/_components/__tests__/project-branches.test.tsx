import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ProjectFormDialog } from "../project-form-dialog";

const contacts = [
  { id: 10, organization_id: 1, name: "Acme Corp", display_name: "Acme Corp" },
  { id: 11, organization_id: 1, name: "Globex", display_name: "Globex" },
];
const accounts = [{ id: 21, name: "Operating costs" }];
const orders = [{ id: 31, name: "SO-001" }];

function seedLists() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/dimensions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { accounts } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/sale-orders", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { orders } }),
    ),
  );
}

function seedCreate(status = 200) {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/projects", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({ success: true, message: "OK.", data: { project: { id: 9 } } });
    }),
  );
}

function seedUpdate(status = 200) {
  server.use(
    http.put("*/api/v1/organizations/:organizationId/projects/:id", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({ success: true, message: "OK.", data: { project: { id: 4 } } });
    }),
  );
}

const editInitial = {
  id: 4,
  name: "Website Revamp",
  contactId: 10,
  managerId: null,
  billingType: "time_material",
  billableRate: 50,
  dimensionId: null,
  saleOrderId: null,
  dateStart: null,
  dateEnd: null,
} as never;

const editInitialFilled = {
  id: 5,
  name: "Mobile App",
  contactId: 11,
  managerId: 10,
  billingType: "fixed",
  billableRate: 100,
  dimensionId: 21,
  saleOrderId: 31,
  dateStart: "2026-01-05T00:00:00Z",
  dateEnd: "2026-02-05T00:00:00Z",
} as never;

beforeEach(() => {
  seedLists();
});

describe("ProjectFormDialog branches", () => {
  it("shows edit title with prefilled name", async () => {
    renderWithProviders(
      <ProjectFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFilled}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Ubah proyek")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Mobile App")).toBeInTheDocument();
  });

  it("blocks submit when name and customer are missing", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ProjectFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Proyek baru");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    expect(await screen.findByText("Nama wajib diisi.")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("blocks submit when end date precedes start date", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ProjectFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Proyek baru");
    await user.type(screen.getByLabelText("Nama"), "Bad Dates");
    await user.click(screen.getByRole("combobox", { name: "Pelanggan" }));
    await user.click(await screen.findByRole("option", { name: "Acme Corp" }));
    const start = screen.getByLabelText("Tanggal mulai");
    const end = screen.getByLabelText("Tanggal akhir");
    await user.type(start, "2026-03-01");
    await user.type(end, "2026-01-01");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    expect(
      await screen.findByText("Tanggal akhir harus setelah tanggal mulai."),
    ).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("creates a project with only required fields", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ProjectFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Proyek baru");
    await user.type(screen.getByLabelText("Nama"), "Website Revamp");
    await user.click(screen.getByRole("combobox", { name: "Pelanggan" }));
    await user.click(await screen.findByRole("option", { name: "Acme Corp" }));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("creates a project with manager, accounts, order, and dates", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ProjectFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Proyek baru");
    await user.type(screen.getByLabelText("Nama"), "Mobile App");
    await user.click(screen.getByRole("combobox", { name: "Pelanggan" }));
    await user.click(await screen.findByRole("option", { name: "Globex" }));
    await user.click(screen.getByRole("combobox", { name: "Manager" }));
    await user.click(await screen.findByRole("option", { name: "Acme Corp" }));
    await user.click(screen.getByRole("combobox", { name: "Jenis penagihan" }));
    await user.click(await screen.findByRole("option", { name: "Harga tetap" }));
    await user.click(screen.getByRole("combobox", { name: "Akun dimensi" }));
    await user.click(await screen.findByRole("option", { name: "Operating costs" }));
    await user.click(screen.getByRole("combobox", { name: "Pesanan penjualan" }));
    await user.click(await screen.findByRole("option", { name: "SO-001" }));
    await user.type(screen.getByLabelText("Tanggal mulai"), "2026-01-05");
    await user.type(screen.getByLabelText("Tanggal akhir"), "2026-02-05");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("does not save when creation fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate(500);
    renderWithProviders(
      <ProjectFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Proyek baru");
    await user.type(screen.getByLabelText("Nama"), "Website Revamp");
    await user.click(screen.getByRole("combobox", { name: "Pelanggan" }));
    await user.click(await screen.findByRole("option", { name: "Acme Corp" }));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(screen.getByText("Proyek baru")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("updates a project with empty optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <ProjectFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        onSave={onSave}
      />,
    );

    await screen.findByText("Ubah proyek");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("updates a project with filled optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <ProjectFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFilled}
        onSave={onSave}
      />,
    );

    await screen.findByText("Ubah proyek");
    expect(screen.getByDisplayValue("2026-02-05")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("does not save when update fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate(500);
    renderWithProviders(
      <ProjectFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        onSave={onSave}
      />,
    );

    await screen.findByText("Ubah proyek");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(screen.getByText("Ubah proyek")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("closes without saving on cancel", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <ProjectFormDialog open={true} onOpenChange={onOpenChange} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Proyek baru");
    await user.click(screen.getByRole("button", { name: "Batal" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
