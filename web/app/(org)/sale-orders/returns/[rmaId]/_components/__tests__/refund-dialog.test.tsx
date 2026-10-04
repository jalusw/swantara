import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Journal } from "@/lib/services/swantara";
import { renderWithProviders } from "@/lib/tests";
import { RefundDialog } from "../refund-dialog";

const STAMP = "2026-01-01T00:00:00Z";

const journals: Journal[] = [
  {
    id: 3,
    organizationId: 1,
    name: "General Journal",
    code: "GJ",
    type: "general",
    defaultAccountId: null,
    currencyCode: "USD",
    bankAccountId: null,
    sequenceId: null,
    createdAt: new Date(STAMP),
    updatedAt: new Date(STAMP),
  },
];

beforeEach(() => {});

describe("RefundDialog", () => {
  it("renders the refund form when open", () => {
    renderWithProviders(
      <RefundDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        rmaId="9"
        journals={journals}
        onSave={vi.fn()}
      />,
    );

    expect(screen.getByRole("heading", { name: "Refund Retur" })).toBeInTheDocument();
    expect(
      screen.getByText("Buat catatan kredit untuk barang yang dikembalikan."),
    ).toBeInTheDocument();
  });

  it("accepts a reference through typing", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <RefundDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        rmaId="9"
        journals={journals}
        onSave={vi.fn()}
      />,
    );

    await user.type(screen.getByLabelText("Reference"), "REF-2026-01");

    expect(screen.getByDisplayValue("REF-2026-01")).toBeInTheDocument();
  });
});
