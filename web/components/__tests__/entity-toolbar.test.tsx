import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { EntityToolbar } from "@/components/entity-toolbar";
import { renderWithProviders } from "@/lib/tests";

describe("EntityToolbar", () => {
  it("surfaces search changes and renders actions", async () => {
    const user = userEvent.setup();
    const onSearchChange = vi.fn();
    renderWithProviders(
      <EntityToolbar searchValue="" onSearchChange={onSearchChange}>
        <button type="button">New</button>
      </EntityToolbar>,
    );

    expect(screen.getByRole("button", { name: "New" })).toBeInTheDocument();
    await user.type(screen.getByRole("searchbox"), "acme");
    expect(onSearchChange).toHaveBeenCalledTimes(4);
  });
});
