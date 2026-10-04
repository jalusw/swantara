import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { CustomerFormDialog } from "../customer-form-dialog";

beforeEach(() => {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json(
        {
          success: true,
          message: "Dibuat.",
          data: { contact: { id: 10, name: "Acme Corp" } },
        },
        { status: 201 },
      ),
    ),
  );
});

function renderDialog(onSave: () => void) {
  renderWithProviders(
    <CustomerFormDialog open onOpenChange={() => {}} orgId="1" onSave={onSave} />,
  );
}

describe("CustomerFormDialog", () => {
  it("renders the create form", async () => {
    renderDialog(() => {});

    expect(await screen.findByRole("heading", { name: "Pelanggan baru" })).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Nama")).toBeInTheDocument();
  });

  it("creates a customer on submit", async () => {
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderDialog(onSave);

    await screen.findByRole("heading", { name: "Pelanggan baru" });
    await user.type(screen.getByPlaceholderText("Nama"), "Acme Corp");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(1));
  });
});
