import { fireEvent, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ProductFormDialog } from "../product-form-dialog";

const categories = [
  { id: "1", name: "Furniture" },
  { id: "2", name: "Electronics" },
];

function renderDialog(props?: {
  initial?: React.ComponentProps<typeof ProductFormDialog>["initial"];
  onSave?: (itemId: string) => void;
}) {
  return renderWithProviders(
    <ProductFormDialog
      open={true}
      onOpenChange={vi.fn()}
      orgId="1"
      categories={categories}
      initial={props?.initial}
      onSave={props?.onSave ?? vi.fn()}
    />,
  );
}

beforeEach(() => {
  server.use(
    http.get("*/api/v1/units", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          units: [
            {
              id: 1,
              category_id: 1,
              name: "Unit",
              factor: 1,
              unit_type: "reference",
              rounding: 0.001,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
    ),
  );
});

describe("ProductFormDialog", () => {
  it("renders the create form with defaults", () => {
    renderDialog();

    expect(screen.getByText("New item")).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
    expect(screen.getByLabelText("Sales price")).toBeInTheDocument();
  });

  it("renders the edit title when initial data is provided", () => {
    renderDialog({
      initial: {
        id: "9",
        name: "Chair",
        categoryId: "1",
        type: "stockable",
        tracking: "none",
        unitId: null,
        purchaseUnitId: null,
        listPrice: 150,
        standardCost: 90,
        isPurchasable: true,
        isSellable: true,
        isManufactured: false,
      },
    });

    expect(screen.getByText("Edit item")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Chair")).toBeInTheDocument();
  });

  it("creates a item with the filled values", async () => {
    let created: unknown = null;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/products", async ({ request }) => {
        created = await request.json();
        return HttpResponse.json(
          { success: true, message: "Created.", data: { item: { id: 42 } } },
          { status: 201 },
        );
      }),
    );
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderDialog({ onSave });

    await user.type(screen.getByLabelText("Name"), "Office Chair");
    await user.click(screen.getByLabelText("Category"));
    await user.click(await screen.findByRole("option", { name: "Furniture" }));
    await user.click(screen.getByRole("button", { name: "Save item" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("42"));
    expect(created).toMatchObject({ name: "Office Chair", category_id: 1 });
  });

  it("maps empty optionals to null in the request", async () => {
    let created: unknown = null;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/products", async ({ request }) => {
        created = await request.json();
        return HttpResponse.json(
          { success: true, message: "Created.", data: { item: { id: 43 } } },
          { status: 201 },
        );
      }),
    );
    const user = userEvent.setup();
    renderDialog();

    await user.type(screen.getByLabelText("Name"), "Simple Service");
    await user.click(screen.getByRole("button", { name: "Save item" }));

    await waitFor(() =>
      expect(created).toMatchObject({ category_id: null, unit_id: null, purchase_unit_id: null }),
    );
  });

  it("toggles the sellable and manufactured switches", async () => {
    const user = userEvent.setup();
    renderDialog();

    const sellable = screen.getByLabelText("Can be sold");
    expect(sellable).toHaveAttribute("aria-checked", "true");

    await user.click(sellable);
    expect(sellable).toHaveAttribute("aria-checked", "false");

    const manufactured = screen.getByLabelText("Is manufactured");
    await user.click(manufactured);
    expect(manufactured).toHaveAttribute("aria-checked", "true");
  });

  it("blocks saving when generated skus collide", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderDialog({ onSave });

    await user.type(screen.getByLabelText("Name"), "Dup");
    await user.click(screen.getByRole("button", { name: "Add attribute" }));
    await user.type(screen.getByLabelText("Attribute"), "Size");
    fireEvent.change(screen.getByPlaceholderText("Values (comma-separated)"), {
      target: { value: "x, x" },
    });
    await user.click(screen.getByRole("button", { name: "Save item" }));

    expect(await screen.findByText(/generated more than once/)).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });
});
