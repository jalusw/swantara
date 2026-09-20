import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { LocationFormDialog } from "../location-form-dialog";

beforeEach(() => {});

describe("LocationFormDialog", () => {
  it("renders the create form when open", () => {
    renderWithProviders(
      <LocationFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        warehouseId="1"
        onSave={vi.fn()}
      />,
    );

    expect(screen.getByRole("heading", { name: "New location" })).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Name")).toBeInTheDocument();
  });

  it("accepts a location name through typing", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <LocationFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        warehouseId="1"
        onSave={vi.fn()}
      />,
    );

    await user.type(screen.getByPlaceholderText("Name"), "WH/Overflow");

    expect(screen.getByDisplayValue("WH/Overflow")).toBeInTheDocument();
  });
});
