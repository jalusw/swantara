import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { SubscriptionPlanFormDialog } from "../subscription-plan-form-dialog";

beforeEach(() => {});

describe("SubscriptionPlanFormDialog", () => {
  it("renders the create form", async () => {
    renderWithProviders(
      <SubscriptionPlanFormDialog
        open
        onOpenChange={() => {}}
        orgId="1"
        initial={null}
        onSave={() => {}}
      />,
    );

    expect(await screen.findByText("Create subscription plan")).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
  });

  it("closes without saving from the cancel button", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <SubscriptionPlanFormDialog
        open
        onOpenChange={onOpenChange}
        orgId="1"
        initial={null}
        onSave={() => {}}
      />,
    );

    await screen.findByText("Create subscription plan");
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
