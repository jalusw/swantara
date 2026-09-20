import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ListQuery } from "@/lib/services/swantara";
import { renderWithProviders } from "@/lib/tests";
import { exportCsv } from "@/lib/utils";
import { ServerEntityTable } from "../server-entity-section";

vi.mock("@/lib/utils", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/lib/utils")>()),
  exportCsv: vi.fn(),
}));

type Widget = { id: number; name: string };

const widgets: Widget[] = [
  { id: 1, name: "Alpha Widget" },
  { id: 2, name: "Beta Widget" },
];

async function fetchWidgets(
  _organizationId: number,
  _params?: ListQuery,
): Promise<{ widgets: Widget[] }> {
  return { widgets };
}

function renderTable(props?: {
  statusOptions?: Array<{ value: string; label: string }>;
  exportFileName?: string;
}) {
  return renderWithProviders(
    <ServerEntityTable<Widget, unknown, { widgets: Widget[] }>
      resource="testWidgetsRemainder"
      fetcher={fetchWidgets}
      selectData={(response) => response.widgets}
      columns={[
        { accessorKey: "name", header: "Name" },
        { accessorKey: "id", header: "Id" },
      ]}
      getRowId={(row) => String(row.id)}
      searchKeys={["name"]}
      statusKey="name"
      statusOptions={props?.statusOptions}
      searchPlaceholder="Search widgets..."
      filterLabel="Status"
      allLabel="All"
      ariaLabel="Widgets"
      emptyTitle="No widgets"
      exportFileName={props?.exportFileName}
    />,
  );
}

beforeEach(() => {});

describe("ServerEntityTable remainder", () => {
  it("shows the empty state when selectData resolves no rows", async () => {
    renderWithProviders(
      <ServerEntityTable<Widget, unknown, { widgets: Widget[] }>
        resource="testWidgetsEmpty"
        fetcher={async () => ({ widgets: [] })}
        selectData={(response) => response.widgets}
        columns={[{ accessorKey: "name", header: "Name" }]}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder="Search widgets..."
        ariaLabel="Widgets"
        emptyTitle="Nothing here"
      />,
    );

    expect(await screen.findByText("Nothing here")).toBeInTheDocument();
  });

  it("shows the error state when the fetcher rejects", async () => {
    renderWithProviders(
      <ServerEntityTable<Widget, unknown, { widgets: Widget[] }>
        resource="testWidgetsFailing"
        fetcher={async () => {
          throw new Error("boom");
        }}
        selectData={(response) => response.widgets}
        columns={[{ accessorKey: "name", header: "Name" }]}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder="Search widgets..."
        ariaLabel="Widgets"
      />,
    );

    expect(await screen.findByText("boom")).toBeInTheDocument();
  });

  it("renders a status filter when statusOptions are provided", async () => {
    const user = userEvent.setup();
    renderTable({ statusOptions: [{ value: "Alpha Widget", label: "Alpha" }] });

    expect(await screen.findByText("Alpha Widget")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Filters" }));

    expect(await screen.findByPlaceholderText("Filter id…")).toBeInTheDocument();
  });

  it("exports the visible rows to csv", async () => {
    const user = userEvent.setup();
    renderTable({ exportFileName: "widgets-export" });

    await screen.findByText("Alpha Widget");
    await user.click(screen.getByRole("button", { name: /export/i }));

    await waitFor(() =>
      expect(exportCsv).toHaveBeenCalledWith(
        "widgets-export",
        expect.any(Array),
        expect.any(Array),
      ),
    );
  });

  it("sorts by column header without losing rows", async () => {
    const user = userEvent.setup();
    renderTable();

    await screen.findByText("Alpha Widget");
    await user.click(screen.getByRole("button", { name: /name/i }));

    expect(await screen.findByText("Beta Widget")).toBeInTheDocument();
  });
});
