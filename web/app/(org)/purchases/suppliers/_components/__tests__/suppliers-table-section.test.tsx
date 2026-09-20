import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SuppliersTable } from "../suppliers-table-section";

const STAMP = "2026-01-05T12:00:00Z";

const contacts = [
  {
    id: 1,
    organization_id: 1,
    name: "Acme",
    display_name: "Acme Corp",
    email: "buy@acme.test",
    industry: "manufacturing",
    active: true,
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 2,
    organization_id: 1,
    name: "Globex",
    display_name: null,
    email: null,
    industry: null,
    active: false,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts } }),
    ),
  );
});

describe("SuppliersTable", () => {
  it("renders seeded suppliers with mapped statuses", async () => {
    renderWithProviders(<SuppliersTable />);

    expect(await screen.findByText("Acme Corp")).toBeInTheDocument();
    expect(screen.getByText("buy@acme.test")).toBeInTheDocument();
    expect(screen.getByText("Globex")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByText("Inactive")).toBeInTheDocument();
  });

  it("filters suppliers by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SuppliersTable />);

    await screen.findByText("Acme Corp");
    await user.type(screen.getByPlaceholderText("Search suppliers…"), "globex");

    expect(await screen.findByText("Globex")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Acme Corp")).not.toBeInTheDocument());
  });

  it("shows the empty state when there are no suppliers", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/contacts", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
      ),
    );
    renderWithProviders(<SuppliersTable />);

    expect(await screen.findByText("No results")).toBeInTheDocument();
  });

  it("shows the error state with retry and refetches", async () => {
    let calls = 0;
    server.use(
      http.get("*/api/v1/organizations/:organizationId/contacts", () => {
        calls += 1;
        if (calls === 1) {
          return HttpResponse.json({ success: false, message: "supplier boom" }, { status: 500 });
        }
        return HttpResponse.json({ success: true, message: "OK.", data: { contacts } });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SuppliersTable />);

    expect(await screen.findByText("supplier boom")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Retry" }));

    expect(await screen.findByText("Acme Corp")).toBeInTheDocument();
    expect(calls).toBe(2);
  });
});
