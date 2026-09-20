import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { JobPositionFormDialog } from "../job-position-form-dialog";

const departments = [
  {
    id: 2,
    organization_id: 1,
    name: "Engineering",
    description: null,
    parent_id: null,
    manager_id: null,
    dimension_id: null,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/departments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { departments } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/job-positions", () =>
      HttpResponse.json(
        {
          success: true,
          message: "Created.",
          data: { job_position: { id: 4, name: "Backend Engineer" } },
        },
        { status: 201 },
      ),
    ),
  );
});

function renderDialog(onSave: () => void) {
  renderWithProviders(
    <JobPositionFormDialog open onOpenChange={() => {}} orgId="1" initial={null} onSave={onSave} />,
  );
}

describe("JobPositionFormDialog", () => {
  it("renders the create form", async () => {
    renderDialog(() => {});

    expect(await screen.findByRole("heading", { name: "New position" })).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
  });

  it("creates a job position on submit", async () => {
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderDialog(onSave);

    await screen.findByRole("heading", { name: "New position" });
    await user.type(screen.getByLabelText("Name"), "Backend Engineer");
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(1));
  });
});
