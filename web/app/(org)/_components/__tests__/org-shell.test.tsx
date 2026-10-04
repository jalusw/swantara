import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { CommandPaletteProvider } from "@/components/command-palette";
import { renderWithProviders } from "@/lib/tests";
import { OrgActiveProvider } from "@/providers/org-active-provider";
import { OrgShell } from "../org-shell";

describe("OrgShell", () => {
  it("renders the org brand, navigation, notification bell and page content", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <OrgActiveProvider orgId={1}>
        <CommandPaletteProvider>
          <OrgShell>
            <div>Page content</div>
          </OrgShell>
        </CommandPaletteProvider>
      </OrgActiveProvider>,
    );

    expect(await screen.findByText("Acme Inc")).toBeInTheDocument();
    expect(await screen.findByRole("link", { name: "Dasbor" })).toHaveAttribute(
      "href",
      "/dashboard",
    );
    await user.click(await screen.findByRole("button", { name: "CRM" }));
    expect(await screen.findByRole("link", { name: "Penjualan" })).toHaveAttribute(
      "href",
      "/sale-orders",
    );
    expect(await screen.findByRole("link", { name: "Umum" })).toHaveAttribute("href", "/settings");
    expect(screen.getByRole("button", { name: /Notifikasi/ })).toBeInTheDocument();
    expect(screen.getByText("Page content")).toBeInTheDocument();
  });
});
