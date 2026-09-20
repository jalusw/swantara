import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { DepartmentFormDialog } from "../department-form-dialog";

beforeEach(() => {});

describe("DepartmentFormDialog", () => {
  it("renders the create title with name field", async () => {
    renderWithProviders(
      <DepartmentFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(await screen.findByText("Add department")).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
  });

  it("accepts a department name", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <DepartmentFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    const nameInput = await screen.findByLabelText("Name");
    await user.type(nameInput, "Engineering");

    expect(nameInput).toHaveValue("Engineering");
  });
});
