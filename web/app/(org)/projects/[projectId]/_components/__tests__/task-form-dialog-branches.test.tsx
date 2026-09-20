import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { TaskFormDialog } from "../task-form-dialog";

const CONTACTS = [{ id: 10, organization_id: 1, name: "Acme Corp", display_name: "Acme Corp" }];

function useHandlers(options?: { failSave?: boolean }) {
  const created: unknown[] = [];
  const updated: unknown[] = [];
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: CONTACTS } }),
    ),
    http.post(
      "*/api/v1/organizations/:organizationId/projects/:projectId/tasks",
      async ({ request }) => {
        created.push(await request.json());
        if (options?.failSave) {
          return HttpResponse.json({ success: false, message: "Nope." }, { status: 422 });
        }
        return HttpResponse.json(
          { success: true, message: "Created.", data: { task: { id: 3 } } },
          { status: 201 },
        );
      },
    ),
    http.put(
      "*/api/v1/organizations/:organizationId/projects/:projectId/tasks/:taskId",
      async ({ request }) => {
        updated.push(await request.json());
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      },
    ),
  );
  return { created, updated };
}

const TASK = {
  id: 5,
  name: "Write docs",
  assigneeId: 10,
  stage: "todo",
  plannedHours: 4,
  effectiveHours: 1.5,
  deadline: "2026-03-10T00:00:00Z",
  priority: 2,
  parentTaskId: null,
};

beforeEach(() => {});

describe("TaskFormDialog branches", () => {
  it("creates a task with default values", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    const { created } = useHandlers();
    renderWithProviders(
      <TaskFormDialog open={true} onOpenChange={vi.fn()} orgId="1" projectId={1} onSave={onSave} />,
    );
    await screen.findByText("New task");
    await user.type(screen.getByLabelText("Task"), "Draft release notes");
    await user.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(created.length).toBe(1));
    const body = created[0] as Record<string, unknown>;
    expect(body.name).toBe("Draft release notes");
    expect(body.assignee_id).toBeNull();
    expect(body.deadline).toBeNull();
    expect(onSave).toHaveBeenCalled();
  });

  it("shows validation for an empty name", async () => {
    const user = userEvent.setup();
    const { created } = useHandlers();
    renderWithProviders(
      <TaskFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        projectId={1}
        onSave={vi.fn()}
      />,
    );
    await screen.findByText("New task");
    await user.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText("Task name is required.")).toBeInTheDocument();
    expect(created.length).toBe(0);
  });

  it("keeps the dialog open when creation fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    useHandlers({ failSave: true });
    renderWithProviders(
      <TaskFormDialog open={true} onOpenChange={vi.fn()} orgId="1" projectId={1} onSave={onSave} />,
    );
    await screen.findByText("New task");
    await user.type(screen.getByLabelText("Task"), "Broken task");
    await user.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() =>
      expect(screen.queryByText("Task name is required.")).not.toBeInTheDocument(),
    );
    expect(onSave).not.toHaveBeenCalled();
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  it("edits a task with prefilled values and null deadline", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    const { updated } = useHandlers();
    renderWithProviders(
      <TaskFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        projectId={1}
        initial={{ ...TASK, deadline: null, priority: null, assigneeId: null } as never}
        onSave={onSave}
      />,
    );
    expect(await screen.findByText("Task updated.")).toBeInTheDocument();
    expect(screen.getByLabelText("Task")).toHaveValue("Write docs");
    expect(screen.getByText("Effective hours")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(updated.length).toBe(1));
    expect(onSave).toHaveBeenCalled();
  });

  it("updates a task with mapped payload values", async () => {
    const user = userEvent.setup();
    const { updated } = useHandlers();
    renderWithProviders(
      <TaskFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        projectId={1}
        initial={TASK as never}
        onSave={vi.fn()}
      />,
    );
    await screen.findByText("Task updated.");
    await user.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(updated.length).toBe(1));
    const body = updated[0] as Record<string, unknown>;
    expect(body.assignee_id).toBe(10);
    expect(body.priority).toBe(2);
    expect(String(body.deadline)).toContain("2026-03-10");
  });

  it("selects an assignee through the dropdown", async () => {
    const user = userEvent.setup();
    const { created } = useHandlers();
    renderWithProviders(
      <TaskFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        projectId={1}
        onSave={vi.fn()}
      />,
    );
    await screen.findByText("New task");
    await user.click(screen.getByRole("combobox", { name: "Assignee" }));
    await user.click(await screen.findByRole("option", { name: "Acme Corp" }));
    await user.type(screen.getByLabelText("Task"), "Assigned work");
    await user.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(created.length).toBe(1));
    expect((created[0] as Record<string, unknown>).assignee_id).toBe(10);
  });

  it("selects priority through the dropdown", async () => {
    const user = userEvent.setup();
    const { created } = useHandlers();
    renderWithProviders(
      <TaskFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        projectId={1}
        onSave={vi.fn()}
      />,
    );
    await screen.findByText("New task");
    const triggers = screen.getAllByRole("combobox");
    const lastTrigger = triggers[triggers.length - 1];
    if (lastTrigger === undefined) throw new Error("expected a combobox");
    await user.click(lastTrigger);
    await user.click(await screen.findByRole("option", { name: "High" }));
    await user.type(screen.getByLabelText("Task"), "Urgent work");
    await user.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() => expect(created.length).toBe(1));
    expect((created[0] as Record<string, unknown>).priority).toBe(3);
  });
});
