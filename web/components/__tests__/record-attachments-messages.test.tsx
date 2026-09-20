import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { RecordAttachmentsMessages } from "@/components/record-attachments-messages";
import { renderWithProviders } from "@/lib/tests";

vi.mock("@/lib/services/swantara", () => ({
  getSwantaraService: () => ({
    attachments: { list: vi.fn().mockResolvedValue({ attachments: [] }) },
    messages: { list: vi.fn().mockResolvedValue({ messages: [] }) },
  }),
}));

vi.mock("@/lib/hooks/use-org-query", () => ({
  useOrgListQuery: () => ({
    data: undefined,
    isLoading: true,
    refetch: vi.fn(),
  }),
}));

describe("RecordAttachmentsMessages", () => {
  it("renders with data-slot", () => {
    renderWithProviders(<RecordAttachmentsMessages orgId="1" ownerType="invoice" ownerId={1} />);

    expect(document.querySelector("[data-slot='record-attachments-messages']")).toBeInTheDocument();
  });

  it("renders two card sections", () => {
    const { container } = renderWithProviders(
      <RecordAttachmentsMessages orgId="1" ownerType="invoice" ownerId={1} />,
    );

    const cards = container.querySelectorAll("[data-slot='card']");
    expect(cards.length).toBe(2);
  });

  it("shows loading state while data is loading", () => {
    renderWithProviders(<RecordAttachmentsMessages orgId="1" ownerType="invoice" ownerId={1} />);

    const loadingTexts = screen.getAllByText("Loading...");
    expect(loadingTexts.length).toBe(2);
  });
});
