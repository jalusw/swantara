import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ServiceContractFormDialog } from "../service-contract-form-dialog";

function useContractOptions() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/equipments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { equipments: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/subscriptions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { subscriptions: [] } }),
    ),
  );
}

beforeEach(() => {
  useContractOptions();
});

describe("ServiceContractFormDialog", () => {
  it("renders the create form", async () => {
    renderWithProviders(
      <ServiceContractFormDialog open onOpenChange={() => {}} orgId="1" onSave={() => {}} />,
    );

    expect(await screen.findByText("Create service contract")).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
  });

  it("closes without saving from the cancel button", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <ServiceContractFormDialog open onOpenChange={onOpenChange} orgId="1" onSave={() => {}} />,
    );

    await screen.findByText("Create service contract");
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
