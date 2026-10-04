import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { PayrollRunFormDialog } from "../payroll-run-form-dialog";

beforeEach(() => {});

describe("PayrollRunFormDialog", () => {
  it("renders the period fields", async () => {
    renderWithProviders(
      <PayrollRunFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(await screen.findByText("Proses penggajian baru")).toBeInTheDocument();
    expect(screen.getByLabelText("Awal periode")).toBeInTheDocument();
    expect(screen.getByLabelText("Akhir periode")).toBeInTheDocument();
  });

  it("accepts a custom period end", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <PayrollRunFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    const periodEnd = await screen.findByLabelText("Akhir periode");
    await user.clear(periodEnd);
    await user.type(periodEnd, "2026-01-31");

    expect(periodEnd).toHaveValue("2026-01-31");
  });
});
