import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { CopyField } from "@/components/copy-field";
import { renderWithProviders } from "@/lib/tests";

describe("CopyField", () => {
  it("shows the value and reports a copy", async () => {
    const user = userEvent.setup();
    const onCopy = vi.fn();
    renderWithProviders(<CopyField value="sk_live_123" onCopy={onCopy} />);
    expect(screen.getByText("sk_live_123")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Copy to clipboard" }));
    expect(onCopy).toHaveBeenCalledWith("sk_live_123");
  });

  it("masks values when masked and not revealed", () => {
    renderWithProviders(<CopyField value="secret-token" masked />);
    expect(screen.getByText("••••••••")).toBeInTheDocument();
    expect(screen.queryByText("secret-token")).not.toBeInTheDocument();
  });
});
