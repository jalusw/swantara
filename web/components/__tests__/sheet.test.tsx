import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/sheet";
import { renderWithProviders } from "@/lib/tests";

describe("Sheet", () => {
  it("renders a trigger button", () => {
    renderWithProviders(
      <Sheet>
        <SheetTrigger>Open</SheetTrigger>
      </Sheet>,
    );

    expect(screen.getByRole("button", { name: "Open" })).toHaveAttribute(
      "data-slot",
      "sheet-trigger",
    );
  });

  it("renders content with header, title, description, and footer", () => {
    renderWithProviders(
      <Sheet open>
        <SheetContent>
          <SheetHeader>
            <SheetTitle>Title</SheetTitle>
            <SheetDescription>Description</SheetDescription>
          </SheetHeader>
          <SheetFooter>Footer</SheetFooter>
        </SheetContent>
      </Sheet>,
    );

    expect(screen.getByText("Title")).toHaveAttribute("data-slot", "sheet-title");
    expect(screen.getByText("Description")).toHaveAttribute("data-slot", "sheet-description");
    expect(screen.getByText("Footer")).toHaveAttribute("data-slot", "sheet-footer");
  });

  it("renders header with data-slot", () => {
    renderWithProviders(<SheetHeader data-slot="sheet-header">Header</SheetHeader>);

    expect(screen.getByText("Header")).toHaveAttribute("data-slot", "sheet-header");
  });
});
