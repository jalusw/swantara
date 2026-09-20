import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { QualityPointFormDialog } from "../quality-point-form-dialog";

function useLocalOptions() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { products: [] },
      }),
    ),
    http.get("*/api/v1/units", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { units: [] },
      }),
    ),
    http.post("*/api/v1/organizations/:organizationId/quality-points", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          point: {
            id: 1,
            organization_id: 1,
            item_id: null,
            operation: "Final inspection",
            test_type: "pass_fail",
            norm_min: null,
            norm_max: null,
            unit_id: null,
          },
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalOptions();
});

describe("QualityPointFormDialog", () => {
  it("renders the operation field", async () => {
    renderWithProviders(
      <QualityPointFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(await screen.findByLabelText("Operation")).toBeInTheDocument();
  });

  it("saves a new quality point", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <QualityPointFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await user.type(await screen.findByLabelText("Operation"), "Final inspection");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });
});
