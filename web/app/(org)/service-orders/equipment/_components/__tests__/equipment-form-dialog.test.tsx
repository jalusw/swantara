import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { EquipmentFormDialog } from "../equipment-form-dialog";

function useLocalOptions() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [] } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/equipments", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          equipment: {
            id: 9,
            organization_id: 1,
            name: "Excavator ZX350",
            item_id: null,
            serial_batch_id: null,
            owner_contact_id: null,
            fixed_asset_id: null,
            location: "Site A",
            install_date: null,
            warranty_end: null,
            category: "Heavy",
          },
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalOptions();
});

describe("EquipmentFormDialog", () => {
  it("renders the create form", async () => {
    renderWithProviders(
      <EquipmentFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(await screen.findByText("Create equipment")).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
  });

  it("creates a new equipment", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <EquipmentFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await user.type(await screen.findByLabelText("Name"), "Excavator ZX350");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("9"));
  });
});
