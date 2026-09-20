import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { Thread } from "@/components/thread";
import { renderWithProviders } from "@/lib/tests";

const messages = [
  { id: "m1", author: "Jane Doe", body: "Ready for review." },
  { id: "m2", author: "John Smith", body: "Approved." },
];

describe("Thread", () => {
  it("renders messages with author initials", () => {
    renderWithProviders(<Thread messages={messages} onSubmit={vi.fn()} />);

    expect(screen.getByText("Jane Doe")).toBeInTheDocument();
    expect(screen.getByText("Ready for review.")).toBeInTheDocument();
    expect(screen.getByText("JD")).toBeInTheDocument();
    expect(screen.getByText("JS")).toBeInTheDocument();
  });

  it("clears the comment box after submitting", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    renderWithProviders(<Thread messages={messages} onSubmit={onSubmit} />);

    const textbox = screen.getByRole("textbox", { name: "Write a comment" });
    await user.type(textbox, "Looks good");
    await user.click(screen.getByRole("button", { name: "Send comment" }));

    expect(onSubmit).toHaveBeenCalledWith("Looks good");
    expect(textbox).toHaveValue("");
  });

  it("disables sending while the comment is empty", () => {
    renderWithProviders(<Thread messages={messages} onSubmit={vi.fn()} />);
    expect(screen.getByRole("button", { name: "Send comment" })).toBeDisabled();
  });
});
