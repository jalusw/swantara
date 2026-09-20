import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/alert-dialog";
import { renderWithProviders } from "@/lib/tests";

describe("AlertDialog", () => {
  it("renders the trigger button", () => {
    renderWithProviders(
      <AlertDialog>
        <AlertDialogTrigger>Open dialog</AlertDialogTrigger>
      </AlertDialog>,
    );

    expect(screen.getByRole("button", { name: "Open dialog" })).toHaveAttribute(
      "data-slot",
      "alert-dialog-trigger",
    );
  });

  it("renders the content with title and description when opened", () => {
    renderWithProviders(
      <AlertDialog open>
        <AlertDialogContent>
          <AlertDialogTitle>Confirm</AlertDialogTitle>
          <AlertDialogDescription>Are you sure?</AlertDialogDescription>
          <AlertDialogAction>Yes</AlertDialogAction>
          <AlertDialogCancel>No</AlertDialogCancel>
        </AlertDialogContent>
      </AlertDialog>,
    );

    expect(screen.getByRole("alertdialog")).toHaveAttribute("data-slot", "alert-dialog-content");
    expect(screen.getByText("Confirm")).toBeInTheDocument();
    expect(screen.getByText("Are you sure?")).toBeInTheDocument();
  });

  it("renders action and cancel buttons with correct data-slots", () => {
    renderWithProviders(
      <AlertDialog open>
        <AlertDialogContent>
          <AlertDialogTitle>Title</AlertDialogTitle>
          <AlertDialogAction>OK</AlertDialogAction>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
        </AlertDialogContent>
      </AlertDialog>,
    );

    expect(screen.getByRole("button", { name: "OK" })).toHaveAttribute(
      "data-slot",
      "alert-dialog-action",
    );
    expect(screen.getByRole("button", { name: "Cancel" })).toHaveAttribute(
      "data-slot",
      "alert-dialog-cancel",
    );
  });
});
