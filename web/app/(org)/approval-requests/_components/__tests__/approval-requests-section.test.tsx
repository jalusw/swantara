import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { navigationMock, renderWithProviders, server } from "@/lib/tests";
import { ApprovalRequestsSection } from "../approval-requests-section";

const STAMP = "2026-01-01T00:00:00Z";

const approvalRequests = [
  {
    id: 1,
    owner_type: "sale_order",
    owner_id: 21,
    requested_by: 9,
    state: "pending",
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 2,
    owner_type: "purchase_order",
    owner_id: 33,
    requested_by: 10,
    state: "refused",
    created_at: STAMP,
    updated_at: STAMP,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/approval-requests", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { approval_requests: approvalRequests },
      }),
    ),
  );
});

describe("ApprovalRequestsSection", () => {
  it("renders seeded approval requests", async () => {
    renderWithProviders(<ApprovalRequestsSection />);

    expect(await screen.findByText("AR-1")).toBeInTheDocument();
    expect(screen.getByText("sale_order")).toBeInTheDocument();
    expect(screen.getByText("Pending")).toBeInTheDocument();
    expect(screen.getByText("Refused")).toBeInTheDocument();
  });

  it("navigates to the detail page from the view action", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ApprovalRequestsSection />);

    await screen.findByText("AR-1");
    await user.click(screen.getAllByRole("button", { name: "View" })[0]!);

    expect(navigationMock.push).toHaveBeenCalledWith("/approval-requests/1");
  });

  it("filters requests by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ApprovalRequestsSection />);

    await screen.findByText("AR-1");
    await user.type(screen.getByPlaceholderText("Search approvals…"), "purchase_order");

    expect(await screen.findByText("AR-2")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("AR-1")).not.toBeInTheDocument());
  });

  it("shows the empty state when there are no requests", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/approval-requests", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { approval_requests: [] } }),
      ),
    );
    renderWithProviders(<ApprovalRequestsSection />);

    expect(await screen.findByText("No approval requests")).toBeInTheDocument();
  });

  it("shows the error state with retry and refetches", async () => {
    let calls = 0;
    server.use(
      http.get("*/api/v1/organizations/:organizationId/approval-requests", () => {
        calls += 1;
        if (calls === 1) {
          return HttpResponse.json({ success: false, message: "approval boom" }, { status: 500 });
        }
        return HttpResponse.json({
          success: true,
          message: "OK.",
          data: { approval_requests: approvalRequests },
        });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ApprovalRequestsSection />);

    expect(await screen.findByText("approval boom")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Retry" }));

    expect(await screen.findByText("AR-1")).toBeInTheDocument();
    expect(calls).toBe(2);
  });
});
