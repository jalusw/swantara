import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { LocationFormDialog } from "../location-form-dialog";

beforeEach(() => {});

describe("LocationFormDialog branches2", () => {
  it("renders the edit title branch with initial values", () => {
    renderWithProviders(
      <LocationFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        warehouseId="1"
        initial={{ id: "4", name: "WH/Stock", code: "STK", usage: "internal" }}
        onSave={vi.fn()}
      />,
    );

    expect(screen.getByRole("heading", { name: "Ubah lokasi" })).toBeInTheDocument();
    expect(screen.getByDisplayValue("WH/Stock")).toBeInTheDocument();
  });

  it("creates a location with null parent branch", async () => {
    let createBody: unknown = null;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/stock-locations", async ({ request }) => {
        createBody = await request.json();
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <LocationFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        warehouseId="1"
        onSave={onSave}
      />,
    );

    await user.type(screen.getByPlaceholderText("Nama"), "WH/New");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
    expect(createBody).toMatchObject({ name: "WH/New" });
  });

  it("updates a location through the edit branch", async () => {
    let updateCalls = 0;
    server.use(
      http.put("*/api/v1/organizations/:organizationId/stock-locations/:id", () => {
        updateCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <LocationFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        warehouseId="1"
        initial={{ id: "4", name: "WH/Stock", code: null, usage: "internal" }}
        onSave={onSave}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Simpan" }));
    await waitFor(() => expect(updateCalls).toBeGreaterThan(0));
  });

  it("shows validation errors for the empty name branch", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <LocationFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        warehouseId="1"
        onSave={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Simpan" }));
    expect(await screen.findByText("Nama lokasi wajib diisi.")).toBeInTheDocument();
  });
});
