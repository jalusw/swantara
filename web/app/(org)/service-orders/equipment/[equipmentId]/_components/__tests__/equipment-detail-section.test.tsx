import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { EquipmentDetail } from "../equipment-detail-section";

function useLocalEquipment() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/equipments/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          equipment: {
            id: 1,
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
  useLocalEquipment();
});

describe("EquipmentDetail", () => {
  it("renders the equipment name with location", async () => {
    renderWithProviders(<EquipmentDetail orgId="1" equipmentId="1" />);

    expect(await screen.findByRole("heading", { name: "Excavator ZX350" })).toBeInTheDocument();
    expect(screen.getByText("Site A")).toBeInTheDocument();
  });

  it("shows overview details on tab select", async () => {
    const user = userEvent.setup();
    renderWithProviders(<EquipmentDetail orgId="1" equipmentId="1" />);

    await screen.findByRole("heading", { name: "Excavator ZX350" });
    await user.click(screen.getByRole("tab", { name: "Ringkasan" }));

    expect(await screen.findByText("Heavy")).toBeInTheDocument();
  });
});
