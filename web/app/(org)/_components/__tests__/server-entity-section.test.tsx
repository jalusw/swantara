import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import type { ListQuery } from "@/lib/services/swantara";
import { renderWithProviders } from "@/lib/tests";
import { ServerEntityTable } from "../server-entity-section";

type Widget = { id: number; name: string };

const widgets: Widget[] = [
  { id: 1, name: "Alpha Widget" },
  { id: 2, name: "Beta Widget" },
];

async function fetchWidgets(
  _organizationId: number,
  params?: ListQuery,
): Promise<{ widgets: Widget[] }> {
  const needle = params?.filter
    ?.map((entry) => entry.split(":").at(-1)?.replaceAll("%", "").toLowerCase() ?? "")
    .find((value) => value.length > 0);
  if (!needle) return { widgets };
  return { widgets: widgets.filter((widget) => widget.name.toLowerCase().includes(needle)) };
}

function renderTable() {
  return renderWithProviders(
    <ServerEntityTable<Widget, unknown, { widgets: Widget[] }>
      resource="testWidgets"
      fetcher={fetchWidgets}
      selectData={(response) => response.widgets}
      columns={[{ accessorKey: "name", header: "Nama" }]}
      getRowId={(row) => String(row.id)}
      searchKeys={["name"]}
      searchPlaceholder="Search widgets..."
      ariaLabel="Widgets"
      emptyTitle="No widgets"
    />,
  );
}

beforeEach(() => {});

describe("ServerEntityTable", () => {
  it("renders rows resolved through selectData", async () => {
    renderTable();

    expect(await screen.findByText("Alpha Widget")).toBeInTheDocument();
    expect(screen.getByText("Beta Widget")).toBeInTheDocument();
  });

  it("filters rows through the search box", async () => {
    const user = userEvent.setup();
    renderTable();

    await screen.findByText("Alpha Widget");
    await user.type(screen.getByPlaceholderText("Search widgets..."), "Beta");

    expect(await screen.findByText("Beta Widget")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Alpha Widget")).not.toBeInTheDocument());
  });

  it("regression: uses backend pagination meta for row and page counts", async () => {
    renderWithProviders(
      <ServerEntityTable<Widget, unknown, { widgets: Widget[] }>
        resource="testWidgetsPaginationMeta"
        fetcher={async () => ({
          widgets,
          meta: { pagination: { page: 1, perPage: 10, total: 25, totalPages: 3 } },
        })}
        selectData={(response) => response.widgets}
        columns={[{ accessorKey: "name", header: "Nama" }]}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder="Search widgets..."
        ariaLabel="Widgets"
      />,
    );

    expect(await screen.findByText("Alpha Widget")).toBeInTheDocument();
    expect(await screen.findByText("1–10 dari 25")).toBeInTheDocument();
    expect(screen.getByText("Halaman 1 dari 3")).toBeInTheDocument();
  });
});
