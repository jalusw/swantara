import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { QualityPoint } from "@/lib/services/swantara";
import { renderWithProviders, server } from "@/lib/tests";
import { QualityPointFormDialog } from "../quality-point-form-dialog";

function useLocalOptions() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { products: [{ id: 5, name: "Widget" }] },
      }),
    ),
    http.get("*/api/v1/units", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { units: [{ id: 3, name: "Gram" }] },
      }),
    ),
    http.post("*/api/v1/organizations/:organizationId/quality-points", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { point: { id: 1 } } }),
    ),
  );
}

beforeEach(() => {
  useLocalOptions();
});

const fullInitial = {
  id: 1,
  organizationId: 1,
  itemId: 5,
  operation: "Final",
  testType: "measure",
  normMin: 1,
  normMax: 2,
  unitId: 3,
} as QualityPoint;

describe("QualityPointFormDialog branches", () => {
  it("renders edit defaults for an existing point", async () => {
    renderWithProviders(
      <QualityPointFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={fullInitial}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByLabelText("Operasi")).toHaveValue("Final");
    expect(screen.getByLabelText("Nilai minimum")).toHaveValue(1);
    expect(screen.getByLabelText("Nilai maksimum")).toHaveValue(2);
  });

  it("saves an edited point", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <QualityPointFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={fullInitial}
        onSave={onSave}
      />,
    );

    await screen.findByLabelText("Operasi");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("selects then clears item and unit", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <QualityPointFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    const dialog = await screen.findByRole("dialog");
    const productBox = within(dialog).getByRole("combobox", { name: "Item" });
    await user.click(productBox);
    await user.click(await screen.findByRole("option", { name: "Widget" }));
    await user.click(productBox);
    await user.click(await screen.findByRole("option", { name: "—" }));

    const uomBox = within(dialog).getByRole("combobox", { name: "Uom" });
    await user.click(uomBox);
    await user.click(await screen.findByRole("option", { name: "Gram" }));
    await user.click(uomBox);
    await user.click(await screen.findByRole("option", { name: "—" }));

    expect(productBox).toBeInTheDocument();
  });

  it("types then clears operation and norms", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <QualityPointFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    const operation = await screen.findByLabelText("Operasi");
    await user.type(operation, "Visual check");
    await user.clear(operation);

    const normMin = screen.getByLabelText("Nilai minimum");
    await user.type(normMin, "1.5");
    await user.clear(normMin);

    const normMax = screen.getByLabelText("Nilai maksimum");
    await user.type(normMax, "9.5");
    await user.clear(normMax);

    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });
});
