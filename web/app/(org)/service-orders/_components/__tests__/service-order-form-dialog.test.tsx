import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ServiceOrderFormDialog } from "../service-order-form-dialog";

function useLocalOptions() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/equipments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { equipments: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/service-contracts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { service_contracts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/employees", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { employees: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [] } }),
    ),
  );
}

beforeEach(() => {
  useLocalOptions();
});

describe("ServiceOrderFormDialog", () => {
  it("renders the create form", async () => {
    renderWithProviders(
      <ServiceOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(await screen.findByText("Create service order")).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
  });

  it("adds another service line", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <ServiceOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Create service order");
    await user.click(screen.getByRole("button", { name: "Add line" }));

    expect(screen.getAllByLabelText("Qty", { exact: false }).length).toBeGreaterThan(1);
  });
});
