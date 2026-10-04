import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ContactFormDialog } from "../contact-form-dialog";

function seedCreate(status = 200) {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/contacts", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { contact: { id: 9 } },
      });
    }),
  );
}

function seedUpdate(status = 200) {
  server.use(
    http.put("*/api/v1/organizations/:organizationId/contacts/:id", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { contact: { id: 4 } },
      });
    }),
  );
}

const editInitial = {
  id: 4,
  organizationId: 1,
  name: "Bluebird Trading",
  displayName: null,
  isOrganization: true,
  parentId: null,
  email: null,
  phone: null,
  mobile: null,
  website: null,
  taxId: null,
  industry: null,
  currencyCode: null,
  lang: "en",
  active: true,
} as never;

const editInitialFilled = {
  id: 5,
  organizationId: 1,
  name: "PT Nusantara",
  displayName: "Nusantara",
  isOrganization: false,
  parentId: null,
  email: "finance@nusantara.id",
  phone: "+62 21",
  mobile: "+62 811",
  website: "https://nusantara.id",
  taxId: "02.345",
  industry: "logistics",
  currencyCode: "IDR",
  lang: "id",
  active: false,
} as never;

beforeEach(() => {});

describe("ContactFormDialog branches", () => {
  it("blocks submit when name is empty", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ContactFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Kontak baru");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    expect(await screen.findByText("Masukkan nama.")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("blocks submit with an invalid email", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ContactFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Kontak baru");
    await user.type(screen.getByPlaceholderText("Nama"), "Acme");
    await user.type(screen.getByPlaceholderText("Email"), "not-an-email");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    expect(await screen.findByText("Masukkan alamat email yang valid.")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("creates a contact with only a name", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ContactFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Kontak baru");
    await user.type(screen.getByPlaceholderText("Nama"), "Acme");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("creates a contact with all fields filled", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ContactFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Kontak baru");
    await user.type(screen.getByPlaceholderText("Nama"), "PT Nusantara");
    await user.type(screen.getByPlaceholderText("Nama tampilan"), "Nusantara");
    await user.type(screen.getByPlaceholderText("Email"), "finance@nusantara.id");
    await user.type(screen.getByPlaceholderText("Telepon"), "+62 21");
    await user.type(screen.getByPlaceholderText("Seluler"), "+62 811");
    await user.type(screen.getByPlaceholderText("Situs web"), "https://nusantara.id");
    await user.type(screen.getByPlaceholderText("NPWP"), "02.345");
    await user.type(screen.getByPlaceholderText("Industri"), "logistics");
    await user.click(screen.getByRole("switch", { name: "Ini adalah organisasi" }));
    await user.click(screen.getByRole("switch", { name: "Aktif" }));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("creates a contact with Indonesian language", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ContactFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Kontak baru");
    await user.type(screen.getByPlaceholderText("Nama"), "PT Nusantara");
    await user.click(screen.getByRole("combobox", { name: "Bahasa" }));
    await user.click(await screen.findByRole("option", { name: "Bahasa Indonesia" }));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("does not save when creation fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate(500);
    renderWithProviders(
      <ContactFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Kontak baru");
    await user.type(screen.getByPlaceholderText("Nama"), "Acme");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(screen.getByText("Kontak baru")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("updates a contact with empty optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <ContactFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        onSave={onSave}
      />,
    );

    await screen.findByText("Ubah kontak");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("updates a contact with filled optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <ContactFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFilled}
        onSave={onSave}
      />,
    );

    await screen.findByText("Ubah kontak");
    expect(screen.getByDisplayValue("Nusantara")).toBeInTheDocument();
    await user.clear(screen.getByDisplayValue("Nusantara"));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("does not save when update fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate(500);
    renderWithProviders(
      <ContactFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        onSave={onSave}
      />,
    );

    await screen.findByText("Ubah kontak");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(screen.getByText("Ubah kontak")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("closes without saving on cancel", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <ContactFormDialog open={true} onOpenChange={onOpenChange} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Kontak baru");
    await user.click(screen.getByRole("button", { name: "Batal" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
