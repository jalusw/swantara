import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { LeadFormDialog } from "../lead-form-dialog";

function useLeadDialogHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/stages", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { stages: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/teams", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { teams: [] } }),
    ),
  );
}

beforeEach(() => {
  useLeadDialogHandlers();
});

describe("LeadFormDialog", () => {
  it("renders the create form when open", async () => {
    renderWithProviders(
      <LeadFormDialog open onOpenChange={() => {}} orgId="1" onSave={() => {}} />,
    );

    expect(await screen.findByText("New lead")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Name")).toBeInTheDocument();
  });

  it("accepts a lead name through the name field", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <LeadFormDialog open onOpenChange={() => {}} orgId="1" onSave={() => {}} />,
    );

    const nameInput = await screen.findByPlaceholderText("Name");
    await user.type(nameInput, "Acme Website Inquiry");

    expect(nameInput).toHaveValue("Acme Website Inquiry");
  });
});
