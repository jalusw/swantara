import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { navigationMock, renderWithProviders, server } from "@/lib/tests";
import { ApprovalRequestDetail } from "../approval-request-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

const approvalRequest = {
  id: 5,
  owner_type: "sale_order",
  owner_id: 21,
  requested_by: 9,
  state: "pending",
  created_at: STAMP,
  updated_at: STAMP,
  steps: [
    {
      id: 12,
      approver_id: 9,
      sequence: 2,
      decision: "approved",
      decided_at: STAMP,
      comment: "Looks good",
    },
    {
      id: 11,
      approver_id: 10,
      sequence: 1,
      decision: "pending",
      decided_at: null,
      comment: null,
    },
  ],
};

let decideCalled = false;

beforeEach(() => {
  decideCalled = false;
  server.use(
    http.get("*/api/v1/organizations/:organizationId/approval-requests/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { approval_request: approvalRequest },
      }),
    ),
    http.post("*/api/v1/organizations/:organizationId/approval-requests/:id/decide", () => {
      decideCalled = true;
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { approval_request: approvalRequest },
      });
    }),
  );
});

describe("ApprovalRequestDetail", () => {
  it("renders the request with steps sorted by sequence", async () => {
    renderWithProviders(<ApprovalRequestDetail orgId="1" requestId="5" />);

    expect(await screen.findByText("AR-5")).toBeInTheDocument();
    expect(screen.getAllByText("Looks good")).toHaveLength(2);

    const stepsCard = screen
      .getByText("Approval steps")
      .closest('[data-slot="card"]') as HTMLElement | null;
    if (!stepsCard) throw new Error("expected steps card");
    const steps = within(stepsCard).getAllByText(/Step \d/);
    expect(steps).toHaveLength(2);
    expect(steps[0]).toHaveTextContent("Step 1");
    expect(steps[1]).toHaveTextContent("Step 2");
  });

  it("navigates back from the back button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ApprovalRequestDetail orgId="1" requestId="5" />);

    await screen.findByText("AR-5");
    await user.click(screen.getAllByRole("button")[0]!);

    expect(navigationMock.back).toHaveBeenCalled();
  });

  it("decides the pending step from the approve action", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ApprovalRequestDetail orgId="1" requestId="5" />);

    await screen.findByText("AR-5");
    await user.click(screen.getAllByRole("button", { name: "Approve" })[0]!);

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("Approve step")).toBeInTheDocument();

    await user.click(within(dialog).getByRole("button", { name: "Approve" }));
    await waitFor(() => expect(decideCalled).toBe(true));
  });

  it("shows empty steps for a refused request", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/approval-requests/:id", () =>
        HttpResponse.json({
          success: true,
          message: "OK.",
          data: {
            approval_request: { ...approvalRequest, id: 6, state: "refused", steps: [] },
          },
        }),
      ),
    );
    renderWithProviders(<ApprovalRequestDetail orgId="1" requestId="6" />);

    expect(await screen.findByText("AR-6")).toBeInTheDocument();
    expect(screen.getByText("No approval steps.")).toBeInTheDocument();
    expect(screen.getAllByText("Refused").length).toBeGreaterThan(0);
  });
});
