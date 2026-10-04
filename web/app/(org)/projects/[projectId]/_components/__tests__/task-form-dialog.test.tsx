import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { TaskFormDialog } from "../task-form-dialog";

function useLocalContacts() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          contacts: [{ id: 10, organization_id: 1, name: "Acme Corp", display_name: "Acme Corp" }],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalContacts();
});

describe("TaskFormDialog", () => {
  it("renders create title", async () => {
    renderWithProviders(
      <TaskFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        projectId={1}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Tugas baru")).toBeInTheDocument();
  });

  it("renders name field", async () => {
    renderWithProviders(
      <TaskFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        projectId={1}
        onSave={vi.fn()}
      />,
    );

    await screen.findByText("Tugas baru");
    expect(screen.getByLabelText("Nama tugas")).toBeInTheDocument();
  });
});
