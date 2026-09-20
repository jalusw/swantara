import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { type LineItem, LineItemsTable } from "@/components/line-items-table";
import { renderWithProviders } from "@/lib/tests";

const ROWS: LineItem[] = [
  { id: "1", description: "Laptop", quantity: 2, unitPrice: 1200 },
  { id: "2", description: "Mouse", quantity: 3, unitPrice: 25 },
];

describe("LineItemsTable", () => {
  it("renders rows with computed amounts and the total", () => {
    renderWithProviders(<LineItemsTable items={ROWS} />);

    expect(screen.getByText("Laptop")).toBeInTheDocument();
    expect(screen.getByText("$2,400.00")).toBeInTheDocument();
    expect(screen.getByText("$75.00")).toBeInTheDocument();
    expect(screen.getByText("Total")).toBeInTheDocument();
    expect(screen.getByText("$2,475.00")).toBeInTheDocument();
  });

  it("adds a new line via the add button", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<LineItemsTable items={ROWS} onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: "Add line" }));
    expect(onChange).toHaveBeenCalled();
    expect(onChange.mock.calls[0]![0]).toHaveLength(3);
  });

  it("removes a line via its remove button", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<LineItemsTable items={ROWS} onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: /Remove.*Mouse/ }));
    expect(onChange).toHaveBeenCalledWith([
      { id: "1", description: "Laptop", quantity: 2, unitPrice: 1200 },
    ]);
  });

  it("recalculates the amount and total when a quantity changes", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<LineItemsTable items={ROWS} onChange={onChange} />);

    const quantityInputs = screen.getAllByRole("spinbutton");
    expect(quantityInputs[0]).toHaveValue(2);
    await user.clear(quantityInputs[0]!);
    await user.type(quantityInputs[0]!, "1");
    expect(onChange).toHaveBeenCalled();
  });
});
