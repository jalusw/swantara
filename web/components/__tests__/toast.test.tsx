import { render, screen, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { describe, expect, it } from "vitest";
import { Toaster } from "../toast";

describe("Toaster", () => {
  it("renders the sonner toaster section", async () => {
    render(<Toaster />);

    await waitFor(() => expect(screen.getByLabelText("Notifications alt+T")).toBeInTheDocument());
  });

  it("applies the custom class to the toast list when a toast fires", async () => {
    render(<Toaster />);
    toast("Hello world");

    await waitFor(() => expect(document.querySelector(".toaster")).toBeInTheDocument());
  });
});
