import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { EntitySection } from "@/components/entity-section";
import { renderWithProviders } from "@/lib/tests";

type Row = { id: string; name: string };

const columns = [{ accessorKey: "name", header: "Name" }] as never;

function renderSection(overrides: Record<string, unknown> = {}) {
  const onRetry = vi.fn();
  const onDialogOpenChange = vi.fn();
  const onAddClick = vi.fn();
  const result = renderWithProviders(
    <EntitySection<Row>
      orgId="org-1"
      isLoading={false}
      error={null}
      onRetry={onRetry}
      columns={columns}
      data={[{ id: "1", name: "Acme" }]}
      getRowId={(row) => row.id}
      searchKeys={["name"]}
      searchPlaceholder="Search customers"
      ariaLabel="Customers"
      dialogOpen={false}
      onDialogOpenChange={onDialogOpenChange}
      dialog={<div>Customer dialog</div>}
      onAddClick={onAddClick}
      addLabel="Add customer"
      {...(overrides as object)}
    />,
  );
  return { onRetry, onDialogOpenChange, onAddClick, ...result };
}

describe("EntitySection", () => {
  it("renders rows, search, add action, and dialog", () => {
    renderSection();
    expect(screen.getByText("Acme")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Search customers")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Add customer" })).toBeInTheDocument();
    expect(screen.getByText("Customer dialog")).toBeInTheDocument();
  });

  it("notifies when the add button is clicked", async () => {
    const { onAddClick } = renderSection();
    await userEvent.setup().click(screen.getByRole("button", { name: "Add customer" }));
    expect(onAddClick).toHaveBeenCalledTimes(1);
  });

  it("shows a loading skeleton while fetching", () => {
    const { container } = renderSection({ isLoading: true });
    expect(screen.queryByText("Acme")).not.toBeInTheDocument();
    expect(container.querySelector('[data-slot="data-table"]')).toBeInTheDocument();
  });

  it("shows an error status with retry", async () => {
    const { onRetry } = renderSection({ error: new Error("offline") });
    expect(screen.getByText("offline")).toBeInTheDocument();
    await userEvent.setup().click(screen.getByRole("button", { name: "Retry" }));
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it("renders extra actions alongside the add button", () => {
    renderSection({ actions: <button type="button">Archive</button> });
    expect(screen.getByRole("button", { name: "Archive" })).toBeInTheDocument();
  });
});
