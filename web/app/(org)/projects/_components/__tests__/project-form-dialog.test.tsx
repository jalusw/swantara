import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ProjectFormDialog } from "../project-form-dialog";

function useLocalForm() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          contacts: [{ id: 10, organization_id: 1, name: "Acme Corp", display_name: "Acme Corp" }],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/dimensions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { accounts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/sale-orders", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { orders: [] } }),
    ),
  );
}

beforeEach(() => {
  useLocalForm();
});

describe("ProjectFormDialog", () => {
  it("renders create title", async () => {
    renderWithProviders(
      <ProjectFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(await screen.findByText("Proyek baru")).toBeInTheDocument();
  });

  it("renders name field", async () => {
    renderWithProviders(
      <ProjectFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Proyek baru");
    expect(screen.getByLabelText("Nama")).toBeInTheDocument();
  });
});
