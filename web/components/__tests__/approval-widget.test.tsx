import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ApprovalWidget } from "@/app/(org)/approval-requests/_components/approval-widget-section";
import type { ApprovalRequest } from "@/lib/services/swantara";
import { renderWithProviders } from "@/lib/tests";

vi.mock("@/lib/services/swantara", () => ({
  getSwantaraService: () => ({
    approvalRequests: { decide: vi.fn() },
  }),
}));

function makeRequest(overrides: Partial<ApprovalRequest> = {}): ApprovalRequest {
  return {
    id: 1,
    state: "pending",
    steps: [
      { id: 10, sequence: 1, decision: "approved", decidedAt: "2026-01-15", comment: null },
      { id: 11, sequence: 2, decision: "pending", decidedAt: null, comment: null },
    ],
    ...overrides,
  } as ApprovalRequest;
}

describe("ApprovalWidget", () => {
  it("returns null when approvalRequest is null", () => {
    const { container } = renderWithProviders(<ApprovalWidget approvalRequest={null} orgId="1" />);

    expect(container.firstChild).toBeNull();
  });

  it("renders steps with their decision badges", () => {
    renderWithProviders(<ApprovalWidget approvalRequest={makeRequest()} orgId="1" />);

    expect(screen.getByText("Langkah 1")).toBeInTheDocument();
    expect(screen.getByText("Langkah 2")).toBeInTheDocument();
    expect(screen.getByText("Disetujui")).toBeInTheDocument();
    expect(screen.getAllByText("Menunggu").length).toBeGreaterThanOrEqual(1);
  });

  it("shows no-steps message when steps array is empty", () => {
    renderWithProviders(<ApprovalWidget approvalRequest={makeRequest({ steps: [] })} orgId="1" />);

    expect(screen.getByText("Tidak ada langkah")).toBeInTheDocument();
  });
});
