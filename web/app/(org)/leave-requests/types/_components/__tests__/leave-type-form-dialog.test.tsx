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
          message: "Created.",
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

    expect(await screen.findByRole("heading", { name: "New leave type" })).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
  });

  it("creates a leave type on submit", async () => {
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderDialog(onSave);

    await screen.findByRole("heading", { name: "New leave type" });
    await user.type(screen.getByLabelText("Name"), "Study Leave");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(1));
  });
});
