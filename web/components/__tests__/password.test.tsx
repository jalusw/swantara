import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { Password } from "@/components/password";

describe("Password", () => {
  it("masks the value by default and reveals it on toggle", async () => {
    const user = userEvent.setup();
    const { container } = render(<Password />);

    const input = container.querySelector("input");
    const toggle = screen.getByRole("button");

    expect(input).toHaveAttribute("type", "password");

    await user.click(toggle);
    expect(input).toHaveAttribute("type", "text");

    await user.click(toggle);
    expect(input).toHaveAttribute("type", "password");
  });

  it("keeps the toggle centered while pressed", async () => {
    const user = userEvent.setup();
    render(<Password />);

    const toggle = screen.getByRole("button");

    await user.click(toggle);

    expect(toggle.className).toContain("-translate-y-1/2");
    expect(toggle.className).toContain("active:not-aria-[haspopup]:translate-y-[calc(-50%+1px)]");
  });
});
