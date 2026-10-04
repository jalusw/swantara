import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { MilestoneFormDialog } from "../milestone-form-dialog";

beforeEach(() => {});

describe("MilestoneFormDialog", () => {
  it("renders create title", () => {
    renderWithProviders(
      <MilestoneFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        projectId={1}
        onSave={vi.fn()}
      />,
    );

    expect(screen.getByText("Tonggak baru")).toBeInTheDocument();
  });

  it("renders name field", () => {
    renderWithProviders(
      <MilestoneFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        projectId={1}
        onSave={vi.fn()}
      />,
    );

    expect(screen.getByLabelText("Nama tonggak")).toBeInTheDocument();
  });
});
