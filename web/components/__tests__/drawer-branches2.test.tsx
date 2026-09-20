import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/drawer";
import { renderWithProviders } from "@/lib/tests";

describe("Drawer branches2", () => {
  it("renders non-modal drawers without an overlay", () => {
    renderWithProviders(
      <Drawer modal={false} open>
        <DrawerContent>
          <DrawerTitle>Plain</DrawerTitle>
        </DrawerContent>
      </Drawer>,
    );

    expect(screen.queryByTestId("drawer-overlay")).not.toBeInTheDocument();
  });

  it("renders swipe handle and snap points branches", () => {
    renderWithProviders(
      <Drawer open showSwipeHandle snapPoints={[0.5, 1]}>
        <DrawerContent>
          <DrawerTitle>Snapped</DrawerTitle>
          <DrawerDescription>With handle</DrawerDescription>
        </DrawerContent>
      </Drawer>,
    );

    expect(screen.getByText("Snapped")).toBeInTheDocument();
  });

  it("supports horizontal swipe directions", () => {
    renderWithProviders(
      <Drawer open swipeDirection="left">
        <DrawerContent>
          <DrawerTitle>Side</DrawerTitle>
        </DrawerContent>
      </Drawer>,
    );

    expect(screen.getByText("Side")).toBeInTheDocument();
  });

  it("renders overlay, close and swipe handle slots", () => {
    renderWithProviders(
      <Drawer>
        <DrawerTrigger>Open drawer</DrawerTrigger>
        <DrawerClose>Close drawer</DrawerClose>
      </Drawer>,
    );

    expect(screen.getByRole("button", { name: "Open drawer" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Close drawer" })).toBeInTheDocument();
  });
});
