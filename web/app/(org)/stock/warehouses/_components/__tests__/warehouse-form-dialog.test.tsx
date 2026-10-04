import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { WarehouseFormDialog } from "../warehouse-form-dialog";

beforeEach(() => {});

describe("WarehouseFormDialog", () => {
  it("renders the create form when open", () => {
    renderWithProviders(
      <WarehouseFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(screen.getByRole("heading", { name: "Gudang baru" })).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Nama")).toBeInTheDocument();
  });

  it("accepts a warehouse name through typing", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <WarehouseFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    await user.type(screen.getByPlaceholderText("Nama"), "East Warehouse");

    expect(screen.getByDisplayValue("East Warehouse")).toBeInTheDocument();
  });
});
