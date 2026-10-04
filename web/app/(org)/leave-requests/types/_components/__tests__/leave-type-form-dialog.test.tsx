import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { LeaveTypeFormDialog } from "../leave-type-form-dialog";

beforeEach(() => {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/leave-types", () =>
      HttpResponse.json(
        {
          success: true,
          message: "Dibuat.",
          data: { leave_type: { id: 3, name: "Study Leave" } },
        },
        { status: 201 },
      ),
    ),
  );
});

function renderDialog(onSave: () => void) {
  renderWithProviders(
    <LeaveTypeFormDialog open onOpenChange={() => {}} orgId="1" initial={null} onSave={onSave} />,
  );
}

describe("LeaveTypeFormDialog", () => {
  it("renders the create form", async () => {
    renderDialog(() => {});

    expect(await screen.findByRole("heading", { name: "Jenis cuti baru" })).toBeInTheDocument();
    expect(screen.getByLabelText("Nama")).toBeInTheDocument();
  });

  it("creates a leave type on submit", async () => {
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderDialog(onSave);

    await screen.findByRole("heading", { name: "Jenis cuti baru" });
    await user.type(screen.getByLabelText("Nama"), "Study Leave");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(1));
  });
});
