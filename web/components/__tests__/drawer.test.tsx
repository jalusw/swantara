import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Drawer, DrawerFooter, DrawerHeader, DrawerTrigger } from "@/components/drawer";
import { renderWithProviders } from "@/lib/tests";

describe("Drawer", () => {
  it("renders a trigger button", () => {
    renderWithProviders(
      <Drawer>
        <DrawerTrigger>Open</DrawerTrigger>
      </Drawer>,
    );

    expect(screen.getByRole("button", { name: "Open" })).toHaveAttribute(
      "data-slot",
      "drawer-trigger",
    );
  });

  it("renders header and footer with data-slot", () => {
    renderWithProviders(
      <div>
        <DrawerHeader data-slot="drawer-header">Header</DrawerHeader>
        <DrawerFooter data-slot="drawer-footer">Footer</DrawerFooter>
      </div>,
    );

    expect(screen.getByText("Header")).toHaveAttribute("data-slot", "drawer-header");
    expect(screen.getByText("Footer")).toHaveAttribute("data-slot", "drawer-footer");
  });

  it("passes through a custom className on trigger", () => {
    renderWithProviders(
      <Drawer>
        <DrawerTrigger className="my-drawer-trigger">Open</DrawerTrigger>
      </Drawer>,
    );

    expect(screen.getByRole("button", { name: "Open" })).toHaveClass("my-drawer-trigger");
  });
});
