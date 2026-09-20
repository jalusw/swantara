import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { EquipmentsSection } from "../equipments-section";

function useLocalEquipments() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/equipments", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          equipments: [
            {
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
            {
              id: 2,
              organization_id: 1,
              name: "Forklift FD30",
              item_id: null,
              serial_batch_id: null,
              owner_contact_id: null,
              fixed_asset_id: null,
              location: "Warehouse 1",
              install_date: null,
              warranty_end: null,
              category: "Light",
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalEquipments();
});

describe("EquipmentsSection", () => {
  it("renders seeded equipments", async () => {
    renderWithProviders(<EquipmentsSection orgId="1" />);

    expect(await screen.findByText("Excavator ZX350")).toBeInTheDocument();
    expect(screen.getByText("Forklift FD30")).toBeInTheDocument();
  });

  it("filters equipments by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<EquipmentsSection orgId="1" />);

    await screen.findByText("Excavator ZX350");
    await user.type(screen.getByPlaceholderText("Search equipments…"), "Forklift");

    expect(await screen.findByText("Forklift FD30")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Excavator ZX350")).not.toBeInTheDocument());
  });
});
