import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ActivityFormDialog } from "../activity-form-dialog";

const leads = [{ id: 51, name: "Big Lead" }];
const opportunities = [{ id: 52, name: "Hot Deal" }];
const contacts = [
  { id: 61, name: "Acme Corp", display_name: "Acme Corp" },
  { id: 62, name: "Globex", display_name: null },
];

function seedLists() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/crm/leads", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { leads } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/opportunities", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { opportunities } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts } }),
    ),
  );
}

function seedCreate(status = 200) {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/crm/activities", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { activity: { id: 9 } },
      });
    }),
  );
}

function seedUpdate(status = 200) {
  server.use(
    http.put("*/api/v1/organizations/:organizationId/crm/activities/:id", () => {
      if (status !== 200) {
        return HttpResponse.json({ success: false, message: "Error." }, { status });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { activity: { id: 4 } },
      });
    }),
  );
}

const editInitial = {
  id: 4,
  prospectId: null,
  contactId: null,
  type: "note",
  summary: "Follow up",
  note: null,
  dueDate: null,
  done: null,
} as never;

const editInitialFilled = {
  id: 5,
  prospectId: 51,
  contactId: 61,
  type: "call",
  summary: "Kickoff call",
  note: "Discuss scope",
  dueDate: "2026-02-01T00:00:00Z",
  done: true,
} as never;

beforeEach(() => {
  seedLists();
});

describe("ActivityFormDialog branches", () => {
  it("shows edit title with prefilled summary", async () => {
    renderWithProviders(
      <ActivityFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFilled}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Aktivitas")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Kickoff call")).toBeInTheDocument();
    expect(screen.getByDisplayValue("2026-02-01")).toBeInTheDocument();
  });

  it("lists leads and opportunities together", async () => {
    const user = userEvent.setup();
    seedCreate();
    renderWithProviders(
      <ActivityFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Aktivitas");
    await user.click(screen.getByRole("combobox", { name: "Prospek atau peluang" }));

    expect(await screen.findByRole("option", { name: "Big Lead" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Hot Deal" })).toBeInTheDocument();
  });

  it("blocks submit when summary is empty", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ActivityFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Aktivitas");
    await user.click(screen.getByRole("button", { name: "Tambah Aktivitas" }));

    await waitFor(() => expect(onSave).not.toHaveBeenCalled());
  });

  it("creates an activity with only a summary", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ActivityFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Aktivitas");
    await user.type(screen.getByLabelText("Ringkasan"), "Quick note");
    await user.click(screen.getByRole("button", { name: "Tambah Aktivitas" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("creates an activity with lead, contact, and due date", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate();
    renderWithProviders(
      <ActivityFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Aktivitas");
    await user.click(screen.getByRole("combobox", { name: "Prospek atau peluang" }));
    await user.click(await screen.findByRole("option", { name: "Hot Deal" }));
    await user.click(screen.getByRole("combobox", { name: "Kontak" }));
    await user.click(await screen.findByRole("option", { name: "Globex" }));
    await user.click(screen.getByRole("combobox", { name: "Jenis" }));
    await user.click(await screen.findByRole("option", { name: "Panggilan" }));
    await user.type(screen.getByLabelText("Ringkasan"), "Kickoff call");
    await user.type(screen.getByLabelText("Catatan", { selector: "textarea" }), "Discuss scope");
    await user.type(screen.getByLabelText("Tanggal jatuh tempo"), "2026-02-01");
    await user.click(screen.getByRole("button", { name: "Tambah Aktivitas" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("does not save when creation fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedCreate(500);
    renderWithProviders(
      <ActivityFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await screen.findByText("Aktivitas");
    await user.type(screen.getByLabelText("Ringkasan"), "Quick note");
    await user.click(screen.getByRole("button", { name: "Tambah Aktivitas" }));

    await waitFor(() => expect(screen.getByText("Aktivitas")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("updates an activity with empty optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <ActivityFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        onSave={onSave}
      />,
    );

    await screen.findByText("Aktivitas");
    await user.click(screen.getByRole("button", { name: "Simpan perubahan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("updates an activity with filled optionals", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate();
    renderWithProviders(
      <ActivityFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitialFilled}
        onSave={onSave}
      />,
    );

    await screen.findByText("Aktivitas");
    await user.clear(screen.getByDisplayValue("Discuss scope"));
    await user.click(screen.getByRole("button", { name: "Simpan perubahan" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("does not save when update fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    seedUpdate(500);
    renderWithProviders(
      <ActivityFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={editInitial}
        onSave={onSave}
      />,
    );

    await screen.findByText("Aktivitas");
    await user.click(screen.getByRole("button", { name: "Simpan perubahan" }));

    await waitFor(() => expect(screen.getByText("Aktivitas")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });

  it("closes without saving on cancel", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <ActivityFormDialog open={true} onOpenChange={onOpenChange} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByText("Aktivitas");
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
