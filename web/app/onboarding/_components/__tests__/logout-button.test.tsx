import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";
import { navigationMock, renderWithProviders, server } from "@/lib/tests";
import { LogoutButton } from "../logout-button";

describe("LogoutButton", () => {
  it("posts a logout request and navigates to login", async () => {
    const user = userEvent.setup();
    renderWithProviders(<LogoutButton />);

    await user.click(screen.getByRole("button", { name: /Logout/i }));

    await waitFor(() => {
      expect(navigationMock.push).toHaveBeenCalledWith("/login");
    });
  });

  it("shows an error toast when logout fails", async () => {
    const toastSpy = vi.spyOn(toast, "error");
    server.use(http.post("*/api/v1/auth/logout", () => HttpResponse.error()));
    const user = userEvent.setup();
    renderWithProviders(<LogoutButton />);

    await user.click(screen.getByRole("button", { name: /Logout/i }));

    await waitFor(() => {
      expect(toastSpy).toHaveBeenCalled();
    });
  });
});
