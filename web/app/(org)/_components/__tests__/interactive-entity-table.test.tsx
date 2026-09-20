import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import * as exportModule from "@/lib/utils";
import { InteractiveEntityTable } from "../interactive-entity-table";

type Row = { id: string; name: string; status: "active" | "inactive" };

const rows: Row[] = [
  { id: "a", name: "Alpha", status: "active" },
  { id: "b", name: "Beta", status: "inactive" },
];

function renderTable() {
  return renderWithProviders(
    <InteractiveEntityTable
      columns={[
        {
          id: "name",
          accessorKey: "name",
          header: "Name",
          cell: ({ row }) => row.original.name,
        },
        {
          id: "status",
          accessorKey: "status",
          header: "Status",
          cell: ({ row }) => row.original.status,
        },
      ]}
      data={rows}
      getRowId={(row) => row.id}
      searchKeys={["name"]}
      statusKey="status"
      statusOptions={[{ value: "active", label: "Active" }]}
      searchPlaceholder="Search"
      filterLabel="Filter by status"
      allLabel="All statuses"
      ariaLabel="Entities"
    />,
  );
}

describe("InteractiveEntityTable", () => {
  it("exports the filtered rows as CSV", async () => {
    const user = userEvent.setup();
    const exportSpy = vi.spyOn(exportModule, "exportCsv").mockImplementation(() => {});
    renderTable();

    await user.type(screen.getByPlaceholderText("Search"), "Alpha");
    await user.click(screen.getByRole("button", { name: "Export" }));

    expect(exportSpy).toHaveBeenCalledWith("Entities", ["Name", "Status"], [["Alpha", "active"]]);
    exportSpy.mockRestore();
  });

  it("exports all rows when no search is applied", async () => {
    const user = userEvent.setup();
    const exportSpy = vi.spyOn(exportModule, "exportCsv").mockImplementation(() => {});
    renderTable();

    await user.click(screen.getByRole("button", { name: "Export" }));

    expect(exportSpy).toHaveBeenCalledWith(
      "Entities",
      ["Name", "Status"],
      [
        ["Alpha", "active"],
        ["Beta", "inactive"],
      ],
    );
    exportSpy.mockRestore();
  });
});
