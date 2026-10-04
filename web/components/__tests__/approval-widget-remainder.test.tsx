import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ApprovalWidget } from "@/app/(org)/approval-requests/_components/approval-widget-section";
import type { ApprovalRequest } from "@/lib/services/swantara";
import { renderWithProviders, server } from "@/lib/tests";

function makeRequest(overrides: Partial<ApprovalRequest> = {}): ApprovalRequest {
  return {
    id: 1,
    state: "pending",
    steps: [
      { id: 20, sequence: 2, decision: "pending", decidedAt: null, comment: null },
      { id: 10, sequence: 1, decision: "refused", decidedAt: "2026-01-15", comment: "No budget" },
    ],
    ...overrides,
  } as ApprovalRequest;
}

beforeEach(() => {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/approval-requests/:id/decide", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { approval_request: { id: 1, state: "approved", steps: [] } },
      }),
    ),
  );
});

describe("ApprovalWidget remainder", () => {
  it("sorts steps by sequence and shows refused tone", () => {
    renderWithProviders(<ApprovalWidget approvalRequest={makeRequest()} orgId="1" />);

    const steps = screen.getAllByText(/Langkah \d/);
    expect(steps[0]).toHaveTextContent("Langkah 1");
    expect(steps[1]).toHaveTextContent("Langkah 2");
    expect(screen.getByText("Ditolak")).toBeInTheDocument();
    expect(screen.getByText("No budget")).toBeInTheDocument();
  });

  it("renders decided date when present", () => {
    renderWithProviders(<ApprovalWidget approvalRequest={makeRequest()} orgId="1" />);

    expect(screen.getByText("15 Jan 2026")).toBeInTheDocument();
  });

  it("opens the approve dialog and submits the decision", async () => {
    const user = userEvent.setup();
    const onDecided = vi.fn();
    renderWithProviders(
      <ApprovalWidget approvalRequest={makeRequest()} orgId="1" onDecided={onDecided} />,
    );

    await user.click(screen.getByRole("button", { name: "Setujui" }));

    expect(await screen.findByText("Setujui langkah")).toBeInTheDocument();

    await user.type(screen.getByPlaceholderText("Komentar opsional…"), "Looks good");
    await user.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Setujui" }));

    await waitFor(() => expect(onDecided).toHaveBeenCalled());
  });

  it("opens the refuse dialog and submits", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ApprovalWidget approvalRequest={makeRequest()} orgId="1" />);

    await user.click(screen.getByRole("button", { name: "Tolak" }));

    expect(await screen.findByText("Tolak langkah")).toBeInTheDocument();

    await user.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Tolak" }));

    await waitFor(() => expect(screen.queryByText("Tolak langkah")).not.toBeInTheDocument());
  });

  it("shows an error toast when the decision fails", async () => {
    server.use(
      http.post("*/api/v1/organizations/:organizationId/approval-requests/:id/decide", () =>
        HttpResponse.json({ success: false }, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<ApprovalWidget approvalRequest={makeRequest()} orgId="1" />);

    await user.click(screen.getByRole("button", { name: "Setujui" }));
    expect(await screen.findByText("Setujui langkah")).toBeInTheDocument();
    await user.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Setujui" }));

    await waitFor(() => expect(screen.getByText("Setujui langkah")).toBeInTheDocument());
  });
});
