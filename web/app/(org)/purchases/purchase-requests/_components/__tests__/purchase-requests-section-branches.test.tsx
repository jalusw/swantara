import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PurchaseRequestsSection } from "../purchase-requests-section";

const STAMP = "2026-01-01T00:00:00Z";

const requester = {
  id: 7,
  organization_id: 1,
  name: "Aria Chen",
  display_name: null,
  is_organization: false,
  active: true,
  created_at: STAMP,
  updated_at: STAMP,
};

const department = {
  id: 3,
  organization_id: 1,
  name: "Engineering",
};

function requisitionFixture(overrides = {}) {
  return {
    id: 1,
    organization_id: 1,
    name: "PR-0001",
    requester_id: 7,
    department_id: 3,
    state: "draft",
    needed_by: "2026-03-01",
    lines: [
      {
        id: 1,
        request_id: 1,
        item_id: 5,
        description: null,
        qty: 4,
        unit_id: null,
        needed_by: null,
      },
    ],
    ...overrides,
  };
}

function seedAll(
  requisitions: unknown[],
  opts?: { failRequisitions?: boolean; failContacts?: boolean; failDepartments?: boolean },
) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/purchase-requests", () => {
      if (opts?.failRequisitions) {
        return HttpResponse.json(
          { success: false, message: "Requisitions down." },
          { status: 500 },
        );
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { requisitions },
      });
    }),
    http.get("*/api/v1/organizations/:organizationId/contacts", () => {
      if (opts?.failContacts) {
        return HttpResponse.json({ success: false, message: "Contacts down." }, { status: 500 });
      }
      return HttpResponse.json({ success: true, message: "OK.", data: { contacts: [requester] } });
    }),
    http.get("*/api/v1/organizations/:organizationId/departments", () => {
      if (opts?.failDepartments) {
        return HttpResponse.json({ success: false, message: "Departments down." }, { status: 500 });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { departments: [department] },
      });
    }),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [] } }),
    ),
  );
}

beforeEach(() => {});

describe("PurchaseRequestsSection branches", () => {
  it("renders state variants with name and map fallbacks", async () => {
    seedAll([
      requisitionFixture({ id: 1, name: "PR-A", state: "approved" }),
      requisitionFixture({ id: 2, name: null, state: "cancelled" }),
      requisitionFixture({
        id: 3,
        name: "PR-C",
        state: "done",
        requester_id: 999,
        department_id: 888,
        needed_by: null,
        lines: null,
      }),
      requisitionFixture({ id: 4, name: "PR-D", state: "confirmed", department_id: null }),
    ]);
    renderWithProviders(<PurchaseRequestsSection orgId="1" />);

    expect(await screen.findByText("PR-A")).toBeInTheDocument();
    expect(screen.getByText("PR-2")).toBeInTheDocument();
    expect(screen.getByText("PR-C")).toBeInTheDocument();
    expect(screen.getAllByText("Aria Chen").length).toBeGreaterThan(0);
    expect(screen.getByText("#999")).toBeInTheDocument();
    expect(screen.getByText("#888")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("shows the requisitions error state", async () => {
    seedAll([], { failRequisitions: true });
    renderWithProviders(<PurchaseRequestsSection orgId="1" />);

    expect(await screen.findByText("Requisitions down.")).toBeInTheDocument();
  });

  it("shows the contacts error state", async () => {
    seedAll([requisitionFixture()], { failContacts: true });
    renderWithProviders(<PurchaseRequestsSection orgId="1" />);

    expect(await screen.findByText("Contacts down.")).toBeInTheDocument();
  });

  it("shows the departments error state", async () => {
    seedAll([requisitionFixture()], { failDepartments: true });
    renderWithProviders(<PurchaseRequestsSection orgId="1" />);

    expect(await screen.findByText("Departments down.")).toBeInTheDocument();
  });
});
