import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ActivitiesSection } from "../activities-section";

const activities = [
  {
    id: 1,
    prospect_id: 1,
    contact_id: null,
    type: "call",
    summary: "Intro call with Acme",
    note: "Discussed pricing tiers",
    due_date: "2026-03-10",
    done: false,
    done_at: null,
    user_id: null,
    created_at: "2026-03-01T00:00:00Z",
    updated_at: "2026-03-01T00:00:00Z",
  },
  {
    id: 2,
    prospect_id: 2,
    contact_id: null,
    type: "meeting",
    summary: "Onsite demo at Nusantara",
    note: null,
    due_date: null,
    done: true,
    done_at: "2026-03-12T00:00:00Z",
    user_id: null,
    created_at: "2026-03-02T00:00:00Z",
    updated_at: "2026-03-12T00:00:00Z",
  },
];

function useActivitiesHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/crm/activities", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { activities } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/leads", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { leads: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/opportunities", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { opportunities: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
    ),
  );
}

beforeEach(() => {
  useActivitiesHandlers();
});

describe("ActivitiesSection remainder", () => {
  it("shows the error state with retry when loading fails", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/crm/activities", () =>
        HttpResponse.json({ success: false, message: "Activity load failed." }, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<ActivitiesSection orgId="1" />);

    expect(await screen.findByText("Activity load failed.")).toBeInTheDocument();

    server.resetHandlers();
    useActivitiesHandlers();
    await user.click(screen.getByRole("button", { name: /coba lagi|retry|try again/i }));

    expect(await screen.findAllByText("Intro call with Acme")).toHaveLength(2);
  });

  it("shows the empty timeline message when no activities exist", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/crm/activities", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { activities: [] } }),
      ),
    );
    renderWithProviders(<ActivitiesSection orgId="1" />);

    const timelineCard = screen
      .getByText("Linimasa aktivitas")
      .closest('[data-slot="card"]') as HTMLElement;
    expect(await within(timelineCard).findByText("Belum ada aktivitas")).toBeInTheDocument();
  });

  it("marks an undone activity as done", async () => {
    let doneId: number | null = null;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/crm/activities/:id/done", ({ params }) => {
        doneId = Number(params.id);
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ActivitiesSection orgId="1" />);

    await screen.findAllByText("Intro call with Acme");
    const table = screen.getByRole("table");
    await user.click(within(table).getByRole("button", { name: "Tandai selesai" }));

    await waitFor(() => expect(doneId).toBe(1));
  });

  it("does not offer mark-done for completed activities", async () => {
    renderWithProviders(<ActivitiesSection orgId="1" />);

    await screen.findAllByText("Intro call with Acme");
    expect(screen.getAllByRole("button", { name: "Tandai selesai" })).toHaveLength(1);
  });

  it("deletes an activity and shows a success toast", async () => {
    let deletedId: number | null = null;
    server.use(
      http.delete("*/api/v1/organizations/:organizationId/crm/activities/:id", ({ params }) => {
        deletedId = Number(params.id);
        return HttpResponse.json({ success: true, message: "Deleted." });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<ActivitiesSection orgId="1" />);

    await screen.findAllByText("Intro call with Acme");
    const table = screen.getByRole("table");
    const deleteTriggers = within(table)
      .getAllByRole("button", { name: "Hapus aktivitas" })
      .filter((el) => el.getAttribute("aria-haspopup") === "dialog");
    await user.click(deleteTriggers[0]!);

    const confirm = await screen.findByRole("alertdialog");
    await user.click(within(confirm).getByRole("button", { name: "Hapus" }));

    await waitFor(() => expect(deletedId).toBe(1));
  });

  it("opens the edit dialog prefilled with the activity", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ActivitiesSection orgId="1" />);

    await screen.findAllByText("Intro call with Acme");
    const table = screen.getByRole("table");
    await user.click(within(table).getAllByRole("button", { name: "Ubah aktivitas" })[0]!);

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });
});
