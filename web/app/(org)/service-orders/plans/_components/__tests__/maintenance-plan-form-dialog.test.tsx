import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { MaintenancePlanFormDialog } from "../maintenance-plan-form-dialog";

const equipments = [
  {
    id: 3,
    organization_id: 1,
    name: "Forklift A",
    item_id: null,
    serial_batch_id: null,
    owner_contact_id: null,
    fixed_asset_id: null,
    location: "Gudang",
    install_date: null,
    warranty_end: null,
    category: "forklift",
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/equipments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { equipments } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/maintenance-plans", () =>
      HttpResponse.json(
        {
          success: true,
          message: "Dibuat.",
          data: { maintenance_plan: { id: 6 } },
        },
        { status: 201 },
      ),
    ),
  );
});

function renderDialog(onSave: (id: string) => void) {
  renderWithProviders(
    <MaintenancePlanFormDialog open onOpenChange={() => {}} orgId="1" onSave={onSave} />,
  );
}

describe("MaintenancePlanFormDialog", () => {
  it("renders the create form", async () => {
    renderDialog(() => {});

    expect(
      await screen.findByRole("heading", { name: "Paket perawatan baru" }),
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Nama")).toBeInTheDocument();
  });

  it("creates a maintenance plan on submit", async () => {
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderDialog(onSave);

    await screen.findByRole("heading", { name: "Paket perawatan baru" });

    await user.click(screen.getByRole("combobox", { name: "Peralatan" }));
    await user.click(await screen.findByRole("option", { name: "Forklift A" }));

    await user.type(screen.getByLabelText("Nama"), "Quarterly Inspection");
    const interval = screen.getByLabelText("Interval (hari)");
    await user.clear(interval);
    await user.type(interval, "90");
    await user.type(screen.getByLabelText("Tanggal jatuh tempo berikutnya"), "2026-04-01");

    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("6"));
  });
});
