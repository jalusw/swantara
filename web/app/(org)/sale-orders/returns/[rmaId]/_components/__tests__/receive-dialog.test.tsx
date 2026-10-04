import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Journal } from "@/lib/services/swantara";
import { renderWithProviders } from "@/lib/tests";
import { ReceiveDialog } from "../receive-dialog";

const STAMP = "2026-01-01T00:00:00Z";

const journals: Journal[] = [
  {
    id: 2,
    organizationId: 1,
    name: "Purchase Journal",
    code: "PJ",
    type: "purchase",
    defaultAccountId: null,
    currencyCode: "USD",
    bankAccountId: null,
    sequenceId: null,
    createdAt: new Date(STAMP),
    updatedAt: new Date(STAMP),
  },
  {
    id: 4,
    organizationId: 1,
    name: "Bank Journal",
    code: "BJ",
    type: "bank",
    defaultAccountId: null,
    currencyCode: "USD",
    bankAccountId: null,
    sequenceId: null,
    createdAt: new Date(STAMP),
    updatedAt: new Date(STAMP),
  },
];

beforeEach(() => {});

describe("ReceiveDialog", () => {
  it("renders the receive form when open", () => {
    renderWithProviders(
      <ReceiveDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        rmaId="9"
        journals={journals}
        onSave={vi.fn()}
      />,
    );

    expect(screen.getByRole("heading", { name: "Terima Retur" })).toBeInTheDocument();
    expect(screen.getByText("Terima barang yang dikembalikan ke persediaan.")).toBeInTheDocument();
  });

  it("closes through the cancel button", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <ReceiveDialog
        open={true}
        onOpenChange={onOpenChange}
        orgId="1"
        rmaId="9"
        journals={journals}
        onSave={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Batal" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
