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
    due_date: "2026-03-12",
    done: true,
    done_at: "2026-03-12T00:00:00Z",
    user_id: null,
    created_at: "2026-03-02T00:00:00Z",
    updated_at: "2026-03-12T00:00:00Z",
  },
];

function useActivityHandlers() {
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
  useActivityHandlers();
});

describe("ActivitiesSection", () => {
  it("renders activities with type badges and timeline", async () => {
    renderWithProviders(<ActivitiesSection orgId="1" />);

    await screen.findAllByText("Intro call with Acme");
    const table = screen.getByRole("table");
    expect(within(table).getByText("Intro call with Acme")).toBeInTheDocument();
    expect(within(table).getByText("Onsite demo at Nusantara")).toBeInTheDocument();
    expect(screen.getByText("Panggilan")).toBeInTheDocument();
    expect(screen.getByText("Linimasa aktivitas")).toBeInTheDocument();
  });

  it("filters activities through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ActivitiesSection orgId="1" />);

    await screen.findAllByText("Intro call with Acme");
    const table = screen.getByRole("table");
    await user.type(screen.getByPlaceholderText("Cari aktivitas…"), "Onsite demo");

    expect(await within(table).findByText("Onsite demo at Nusantara")).toBeInTheDocument();
    await waitFor(() =>
      expect(within(table).queryByText("Intro call with Acme")).not.toBeInTheDocument(),
    );
  });

  it("opens the activity dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ActivitiesSection orgId="1" />);

    await screen.findAllByText("Intro call with Acme");
    await user.click(screen.getByRole("button", { name: "Tambah Aktivitas" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });
});
