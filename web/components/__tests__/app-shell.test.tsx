import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { AppShell } from "@/components/app-shell";
import { renderWithProviders } from "@/lib/tests";

function TestShell() {
  return (
    <AppShell
      brand={<div>Acme Inc</div>}
      nav={<nav>Navigation</nav>}
      search={<div>Search</div>}
      actions={<button type="button">Actions</button>}
    >
      <div>Page content</div>
    </AppShell>
  );
}

describe("AppShell", () => {
  it("renders sidebar, header and content", () => {
    renderWithProviders(<TestShell />);

    expect(screen.getByText("Acme Inc")).toBeInTheDocument();
    expect(screen.getByText("Navigation")).toBeInTheDocument();
    expect(screen.getByText("Search")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Actions" })).toBeInTheDocument();
    expect(screen.getByText("Page content")).toBeInTheDocument();
  });

  it("toggles the desktop sidebar via the trigger", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TestShell />);

    const sidebar = document.querySelector("[data-slot='sidebar']");
    expect(sidebar).toHaveAttribute("data-state", "expanded");

    await user.click(screen.getByRole("button", { name: "Toggle sidebar" }));
    expect(sidebar).toHaveAttribute("data-state", "collapsed");
  });
});
