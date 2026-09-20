import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PromoteDialog } from "../promote-section";

const stages = [
  { id: 2, name: "Qualified", sequence: 2, is_won: false, probability: 30 },
  { id: 3, name: "Proposal", sequence: 3, is_won: false, probability: 60 },
];

function seedStages() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/crm/stages", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { stages } }),
    ),
  );
}

beforeEach(() => {
  seedStages();
});

describe("PromoteDialog", () => {
  it("renders the promote form with stages", async () => {
    renderWithProviders(
      <PromoteDialog open onOpenChange={() => {}} orgId="1" prospectId="5" onPromoted={() => {}} />,
    );

    expect(await screen.findByText("Promote lead to opportunity")).toBeInTheDocument();
    expect(screen.getByText("Expected revenue")).toBeInTheDocument();
  });

  it("requires a stage before promoting", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <PromoteDialog open onOpenChange={() => {}} orgId="1" prospectId="5" onPromoted={() => {}} />,
    );

    await screen.findByText("Promote lead to opportunity");
    await user.click(screen.getByRole("button", { name: "Promote" }));

    expect(await screen.findByText("Select a stage.")).toBeInTheDocument();
  });

  it("promotes with optional fields left blank", async () => {
    let payload: unknown = null;
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/crm/leads/:id/promote",
        async ({ request }) => {
          payload = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: { opportunity: {} } });
        },
      ),
    );
    const onPromoted = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <PromoteDialog
        open
        onOpenChange={() => {}}
        orgId="1"
        prospectId="5"
        onPromoted={onPromoted}
      />,
    );

    await user.click(screen.getByRole("combobox", { name: "Stage" }));
    await user.click(await screen.findByRole("option", { name: "Qualified — 30%" }));
    await user.click(screen.getByRole("button", { name: "Promote" }));

    await waitFor(() => expect(onPromoted).toHaveBeenCalled());
    expect(payload).toMatchObject({ stage_id: 2, probability: null, expected_close: null });
  });

  it("sends probability and expected close when provided", async () => {
    let payload: unknown = null;
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/crm/leads/:id/promote",
        async ({ request }) => {
          payload = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: { opportunity: {} } });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(
      <PromoteDialog open onOpenChange={() => {}} orgId="1" prospectId="5" onPromoted={() => {}} />,
    );

    await user.click(screen.getByRole("combobox", { name: "Stage" }));
    await user.click(await screen.findByRole("option", { name: "Proposal — 60%" }));
    await user.type(screen.getByLabelText("Probability %"), "70");
    await user.click(screen.getByRole("button", { name: "Promote" }));

    await waitFor(() => expect(payload).toMatchObject({ stage_id: 3, probability: 70 }));
  });

  it("closes from the cancel button without promoting", async () => {
    let promoteCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/crm/leads/:id/promote", () => {
        promoteCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const onOpenChange = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <PromoteDialog
        open
        onOpenChange={onOpenChange}
        orgId="1"
        prospectId="5"
        onPromoted={() => {}}
      />,
    );

    await screen.findByText("Promote lead to opportunity");
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(promoteCalls).toBe(0);
  });
});
