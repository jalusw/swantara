import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";
import { setActiveOrg } from "@/lib/server/active-org-actions";
import { navigationMock, renderWithProviders, server } from "@/lib/tests";
import { CreateOrganizationDialog } from "../create-organization-dialog";

vi.mock("@/lib/server/active-org-actions", () => ({
  setActiveOrg: vi.fn(async () => {}),
  clearActiveOrg: vi.fn(async () => {}),
}));

describe("CreateOrganizationDialog", () => {
  it("renders name and country fields when open", async () => {
    renderWithProviders(<CreateOrganizationDialog open={true} onOpenChange={vi.fn()} />);

    expect(await screen.findByText("Buat organisasi")).toBeInTheDocument();
    expect(screen.getByLabelText("Nama perusahaan")).toBeInTheDocument();
    expect(screen.getByLabelText("Negara")).toBeInTheDocument();
  });

  it("creates an organization and navigates to its dashboard", async () => {
    const toastSpy = vi.spyOn(toast, "success");
    server.use(
      http.post("*/api/v1/organizations/quick", () =>
        HttpResponse.json(
          { success: true, data: { organization: { id: 9, name: "Acme Inc" } } },
          { status: 201 },
        ),
      ),
    );
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(<CreateOrganizationDialog open={true} onOpenChange={onOpenChange} />);

    await user.type(await screen.findByLabelText("Nama perusahaan"), "Acme Inc");
    await user.click(screen.getByLabelText("Negara"));
    await user.click(await screen.findByRole("option", { name: "Indonesia" }));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => {
      expect(toastSpy).toHaveBeenCalledWith("Organisasi dibuat");
    });
    expect(setActiveOrg).toHaveBeenCalledWith(9);
    expect(navigationMock.push).toHaveBeenCalledWith("/dashboard");
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
