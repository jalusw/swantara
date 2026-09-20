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

    expect(await screen.findByText("New payroll run")).toBeInTheDocument();
    expect(screen.getByLabelText("Period start")).toBeInTheDocument();
    expect(screen.getByLabelText("Period end")).toBeInTheDocument();
  });

  it("accepts a custom period end", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <PayrollRunFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    const periodEnd = await screen.findByLabelText("Period end");
    await user.clear(periodEnd);
    await user.type(periodEnd, "2026-01-31");

    expect(periodEnd).toHaveValue("2026-01-31");
  });
});
