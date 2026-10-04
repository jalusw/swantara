import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ContactAddresses } from "../contact-addresses";

const STAMP = "2026-01-01T00:00:00Z";

function address(id: number, patch: Record<string, unknown> = {}) {
  return {
    id,
    contact_id: 1,
    type: "billing",
    line1: `Line ${id}`,
    line2: null,
    city: "Jakarta",
    state: null,
    postal_code: "12190",
    country_code: "ID",
    is_default: false,
    created_at: STAMP,
    updated_at: STAMP,
    ...patch,
  };
}

function seedAddresses(addresses: unknown[]) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts/:contactId/addresses", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { addresses } }),
    ),
  );
}

beforeEach(() => {
  seedAddresses([address(1, { is_default: true }), address(2, { type: "shipping" })]);
});

describe("ContactAddresses extra2", () => {
  it("renders the empty state when there are no addresses", async () => {
    seedAddresses([]);
    renderWithProviders(<ContactAddresses orgId="1" contactId="1" onRefetch={() => {}} />);

    expect(await screen.findByText("Tidak ada alamat")).toBeInTheDocument();
  });

  it("hides the default action for the default address", async () => {
    renderWithProviders(<ContactAddresses orgId="1" contactId="1" onRefetch={() => {}} />);

    await screen.findByText("Line 1");
    expect(screen.getAllByRole("button", { name: "Jadikan bawaan" })).toHaveLength(1);
    expect(screen.getByText("Bawaan")).toBeInTheDocument();
  });

  it("sets an address as default and refetches", async () => {
    let defaultCalls = 0;
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/contacts/:contactId/addresses/:addressId/default",
        () => {
          defaultCalls += 1;
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const onRefetch = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(<ContactAddresses orgId="1" contactId="1" onRefetch={onRefetch} />);

    await screen.findByText("Line 2");
    await user.click(screen.getByRole("button", { name: "Jadikan bawaan" }));

    await waitFor(() => expect(defaultCalls).toBe(1));
    await waitFor(() => expect(onRefetch).toHaveBeenCalled());
  });

  it("deletes an address after confirmation", async () => {
    let deleteCalls = 0;
    server.use(
      http.delete(
        "*/api/v1/organizations/:organizationId/contacts/:contactId/addresses/:addressId",
        () => {
          deleteCalls += 1;
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const onRefetch = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(<ContactAddresses orgId="1" contactId="1" onRefetch={onRefetch} />);

    await screen.findByText("Line 1");
    await user.click(screen.getAllByRole("button", { name: "Hapus" })[0]!);
    await user.click(await screen.findByRole("button", { name: "Hapus" }));

    await waitFor(() => expect(deleteCalls).toBe(1));
    await waitFor(() => expect(onRefetch).toHaveBeenCalled());
  });

  it("creates an address from the dialog", async () => {
    let createCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/contacts/:contactId/addresses", () => {
        createCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} }, { status: 201 });
      }),
    );
    const onRefetch = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(<ContactAddresses orgId="1" contactId="1" onRefetch={onRefetch} />);

    await screen.findByText("Line 1");
    await user.click(screen.getByRole("button", { name: "Tambah alamat" }));
    const dialog = await screen.findByRole("dialog");

    await user.type(within(dialog).getByPlaceholderText("Baris alamat 1"), "Jl. Baru 5");
    await user.click(within(dialog).getByRole("button", { name: "Simpan alamat" }));

    await waitFor(() => expect(createCalls).toBe(1));
    await waitFor(() => expect(onRefetch).toHaveBeenCalled());
  });

  it("updates an existing address from the edit dialog", async () => {
    let updateCalls = 0;
    server.use(
      http.put(
        "*/api/v1/organizations/:organizationId/contacts/:contactId/addresses/:addressId",
        () => {
          updateCalls += 1;
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<ContactAddresses orgId="1" contactId="1" onRefetch={() => {}} />);

    await screen.findByText("Line 1");
    await user.click(screen.getAllByRole("button", { name: "Ubah" })[0]!);

    expect(await screen.findByText("Ubah alamat")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Simpan alamat" }));

    await waitFor(() => expect(updateCalls).toBe(1));
  });

  it("renders line2 and country suffix when present", async () => {
    seedAddresses([address(1, { line2: "Floor 3", country_code: "SG" })]);
    renderWithProviders(<ContactAddresses orgId="1" contactId="1" onRefetch={() => {}} />);

    expect(await screen.findByText("Floor 3")).toBeInTheDocument();
    expect(screen.getByText(/· SG/)).toBeInTheDocument();
  });
});
