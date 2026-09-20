import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { Checklist } from "@/components/checklist";
import { renderWithProviders } from "@/lib/tests";

const ITEMS = [
  { id: "a", label: "Create a company profile" },
  { id: "b", label: "Add your first employee", description: "Invites follow" },
];

describe("Checklist", () => {
  it("renders the progress count and item labels", () => {
    renderWithProviders(<Checklist items={ITEMS} values={{ a: true }} readOnly />);

    expect(screen.getByText("1 of 2")).toBeInTheDocument();
    expect(screen.getByText("Create a company profile")).toBeInTheDocument();
    expect(screen.getByText("Add your first employee")).toBeInTheDocument();
  });

  it("is read-only when no change handler is provided", () => {
    renderWithProviders(<Checklist items={ITEMS} readOnly />);

    for (const checkbox of screen.getAllByRole("checkbox")) {
      expect(checkbox).toHaveAttribute("aria-disabled", "true");
    }
  });

  it("reports toggles to the parent", async () => {
    const user = userEvent.setup();
    const onCheckedChange = vi.fn();
    renderWithProviders(<Checklist items={ITEMS} onCheckedChange={onCheckedChange} />);

    await user.click(screen.getAllByRole("checkbox")[0]!);
    expect(onCheckedChange).toHaveBeenCalledWith("a", true);
  });

  it("shows the full progress when everything is checked", () => {
    renderWithProviders(<Checklist items={ITEMS} values={{ a: true, b: true }} readOnly />);

    expect(screen.getByText("2 of 2")).toBeInTheDocument();
    expect(screen.getByRole("progressbar")).toHaveAttribute("aria-valuenow", "100");
  });
});
