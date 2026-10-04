import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { GiftCardFormDialog } from "../gift-card-form-dialog";

function useContactOptions() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
    ),
  );
}

beforeEach(() => {
  useContactOptions();
});

describe("GiftCardFormDialog", () => {
  it("renders the issue form", async () => {
    renderWithProviders(
      <GiftCardFormDialog open onOpenChange={() => {}} orgId="1" onSave={() => {}} />,
    );

    expect(await screen.findByText("Terbitkan kartu hadiah")).toBeInTheDocument();
    expect(screen.getByLabelText("Jumlah")).toBeInTheDocument();
  });

  it("closes without saving from the cancel button", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <GiftCardFormDialog open onOpenChange={onOpenChange} orgId="1" onSave={() => {}} />,
    );

    await screen.findByText("Terbitkan kartu hadiah");
    await user.click(screen.getByRole("button", { name: "Batal" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
