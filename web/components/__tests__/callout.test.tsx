import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { Callout } from "@/components/callout";
import { renderWithProviders } from "@/lib/tests";

describe("Callout", () => {
  it("renders a title and children", () => {
    renderWithProviders(
      <Callout variant="warning" title="Tax not filed">
        Update your settings.
      </Callout>,
    );
    expect(screen.getByText("Tax not filed")).toBeInTheDocument();
    expect(screen.getByText("Update your settings.")).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("is a live region for danger variants", () => {
    renderWithProviders(<Callout variant="danger">Failed to sync.</Callout>);
    expect(screen.getByRole("alert")).toBeInTheDocument();
  });

  it("calls onDismiss when dismissed", async () => {
    const user = userEvent.setup();
    const onDismiss = vi.fn();
    renderWithProviders(
      <Callout variant="info" onDismiss={onDismiss}>
        Notice
      </Callout>,
    );
    await user.click(screen.getByRole("button", { name: "Dismiss" }));
    expect(onDismiss).toHaveBeenCalledOnce();
  });
});
