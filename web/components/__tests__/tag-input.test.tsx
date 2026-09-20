import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TagInput } from "@/components/tag-input";
import { renderWithProviders } from "@/lib/tests";

describe("TagInput", () => {
  it("adds a tag on Enter", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<TagInput value={[]} onChange={onChange} />);
    await user.type(screen.getByRole("textbox"), "design{Enter}");
    expect(onChange).toHaveBeenCalledWith(["design"]);
  });

  it("does not add duplicate tags", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<TagInput value={["design"]} onChange={onChange} />);
    await user.type(screen.getByRole("textbox"), "design{Enter}");
    expect(onChange).not.toHaveBeenCalled();
  });

  it("removes a tag via its dismiss button", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<TagInput value={["design"]} onChange={onChange} />);
    await user.click(screen.getByRole("button", { name: "Remove design" }));
    expect(onChange).toHaveBeenCalledWith([]);
  });
});
