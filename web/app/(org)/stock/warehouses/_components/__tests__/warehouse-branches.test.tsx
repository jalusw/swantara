import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { WarehouseFormDialog } from "../warehouse-form-dialog";

function seedCreate(status = 200) {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/warehouses", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          warehouse: {
            id: 9,
            organization_id: 1,
            name: "East Warehouse",
            code: null,
            line1: null,
            line2: null,
            city: null,
            state: null,
            postal_code: null,
            country_code: null,
          },
        },
      });
    }),
  );
}

function seedUpdate(status = 200) {
  server.use(
    http.put("*/api/v1/organizations/:organizationId/warehouses/:id", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          warehouse: {
            id: 4,
            organization_id: 1,
            name: "Main Warehouse",
            code: "WH-01",
            line1: "Jl. Merdeka 1",
            line2: null,
            city: "Jakarta",
            state: null,
            postal_code: null,
            country_code: null,
          },
        },
      });
    }),
  );
}

const editInitial = {
  id: 4,
  organizationId: 1,
  name: "Main Warehouse",
  code: null,
  line1: null,
  line2: null,
  city: null,
  state: null,
  postalCode: null,
  countryCode: null,
} as never;

const editInitialFilled = {
  id: 5,
  organizationId: 1,
  name: "East Warehouse",
  code: "WH-02",
  line1: "Jl. Merdeka 1",
  line2: "Block B",
  city: "Jakarta",
  state: "DKI",
  postalCode: "10110",
  countryCode: "ID",
} as never;

beforeEach(() => {});

describe("WarehouseFormDialog branches", () => {
  it("shows edit title with prefilled values", async () => {
    renderWithProviders(
      <WarehouseFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFilled}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByRole("heading", { name: "Edit warehouse" })).toBeInTheDocument();
    expect(screen.getByDisplayValue("WH-02")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Block B")).toBeInTheDocument();
  });

  it("blocks submit when name is empty", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <WarehouseFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByRole("heading", { name: "New warehouse" });
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Warehouse name is required.")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("creates a warehouse with only a name", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <WarehouseFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await user.type(await screen.findByPlaceholderText("Name"), "East Warehouse");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("9"));
  });

  it("creates a warehouse with a full address", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <WarehouseFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await user.type(await screen.findByPlaceholderText("Name"), "East Warehouse");
    await user.type(screen.getByPlaceholderText("Code"), "WH-02");
    await user.type(screen.getByPlaceholderText("Address line 1"), "Jl. Merdeka 1");
    await user.type(screen.getByPlaceholderText("Address line 2"), "Block B");
    await user.type(screen.getByPlaceholderText("City"), "Jakarta");
    await user.type(screen.getByPlaceholderText("State"), "DKI");
    await user.type(screen.getByPlaceholderText("Postal code"), "10110");
    await user.type(screen.getByPlaceholderText("Country code"), "ID");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("9"));
  });

  it("does not save when creation fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate(500);
    renderWithProviders(
      <WarehouseFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await user.type(await screen.findByPlaceholderText("Name"), "East Warehouse");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() =>
      expect(screen.getByRole("heading", { name: "New warehouse" })).toBeInTheDocument(),
    );
    expect(onSave).not.toHaveBeenCalled();
  });

  it("updates a warehouse from empty optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <WarehouseFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        onSave={onSave}
      />,
    );

    await screen.findByRole("heading", { name: "Edit warehouse" });
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("4"));
  });

  it("updates a warehouse with filled optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <WarehouseFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFilled}
        onSave={onSave}
      />,
    );

    await screen.findByRole("heading", { name: "Edit warehouse" });
    await user.clear(screen.getByDisplayValue("Block B"));
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("4"));
  });

  it("does not save when update fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate(500);
    renderWithProviders(
      <WarehouseFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        onSave={onSave}
      />,
    );

    await screen.findByRole("heading", { name: "Edit warehouse" });
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() =>
      expect(screen.getByRole("heading", { name: "Edit warehouse" })).toBeInTheDocument(),
    );
    expect(onSave).not.toHaveBeenCalled();
  });

  it("closes without saving on cancel", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <WarehouseFormDialog open={true} onOpenChange={onOpenChange} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByRole("heading", { name: "New warehouse" });
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
