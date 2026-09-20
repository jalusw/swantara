import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/dialog";
import { renderWithProviders } from "@/lib/tests";

describe("Dialog", () => {
  it("renders a trigger button", () => {
    renderWithProviders(
      <Dialog>
        <DialogTrigger>Open</DialogTrigger>
      </Dialog>,
    );

    expect(screen.getByRole("button", { name: "Open" })).toHaveAttribute(
      "data-slot",
      "dialog-trigger",
    );
  });

  it("renders content with data-slot when open", () => {
    renderWithProviders(
      <Dialog open>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Title</DialogTitle>
            <DialogDescription>Description</DialogDescription>
          </DialogHeader>
        </DialogContent>
      </Dialog>,
    );

    expect(screen.getByText("Title")).toHaveAttribute("data-slot", "dialog-title");
    expect(screen.getByText("Description")).toHaveAttribute("data-slot", "dialog-description");
  });

  it("renders header and footer slots", () => {
    renderWithProviders(
      <Dialog open>
        <DialogContent>
          <DialogHeader data-slot="dialog-header">Header</DialogHeader>
          <DialogFooter data-slot="dialog-footer">Footer</DialogFooter>
        </DialogContent>
      </Dialog>,
    );

    expect(screen.getByText("Header")).toHaveAttribute("data-slot", "dialog-header");
    expect(screen.getByText("Footer")).toHaveAttribute("data-slot", "dialog-footer");
  });
});
