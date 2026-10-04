import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CrmLead } from "@/lib/services/swantara";
import { renderWithProviders, server } from "@/lib/tests";
import { LeadFormDialog } from "../lead-form-dialog";

const baseLead: CrmLead = {
  id: 3,
  organizationId: 1,
  name: "Acme Website Inquiry",
  type: "lead",
  contactId: null,
  contactName: "Aria",
  email: "aria@example.com",
  phone: null,
  jobPosition: null,
  stageId: null,
  expectedRevenue: 5000,
  probability: 10,
  priority: 0,
  salespersonId: null,
  salesGroupId: null,
  source: null,
  medium: null,
  campaign: null,
  lostReason: null,
  expectedClose: null,
  closedAt: null,
  isWon: false,
  createdAt: new Date("2026-01-01T00:00:00Z"),
  updatedAt: new Date("2026-01-01T00:00:00Z"),
};

function seedLookups() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/stages", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { stages: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/teams", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { teams: [] } }),
    ),
  );
}

beforeEach(() => {
  seedLookups();
});

describe("LeadFormDialog extra", () => {
  it("renders the opportunity create title for opportunity type", async () => {
    renderWithProviders(
      <LeadFormDialog
        open
        onOpenChange={() => {}}
        orgId="1"
        type="opportunity"
        onSave={() => {}}
      />,
    );

    expect(await screen.findByRole("heading", { name: "Peluang baru" })).toBeInTheDocument();
  });

  it("renders the edit title with prefilled values", async () => {
    renderWithProviders(
      <LeadFormDialog
        open
        onOpenChange={() => {}}
        orgId="1"
        initial={baseLead}
        onSave={() => {}}
      />,
    );

    expect(await screen.findByRole("heading", { name: "Ubah prospek" })).toBeInTheDocument();
    expect(screen.getByDisplayValue("Acme Website Inquiry")).toBeInTheDocument();
    expect(screen.getByDisplayValue("aria@example.com")).toBeInTheDocument();
  });

  it("renders the edit opportunity title for opportunity initials", async () => {
    renderWithProviders(
      <LeadFormDialog
        open
        onOpenChange={() => {}}
        orgId="1"
        initial={{ ...baseLead, type: "opportunity" }}
        onSave={() => {}}
      />,
    );

    expect(await screen.findByRole("heading", { name: "Ubah peluang" })).toBeInTheDocument();
  });

  it("requires a name before saving", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <LeadFormDialog open onOpenChange={() => {}} orgId="1" onSave={() => {}} />,
    );

    await screen.findByPlaceholderText("Nama");
    await user.click(screen.getByRole("button", { name: "Buat prospek" }));

    expect(await screen.findByText("Nama wajib diisi.")).toBeInTheDocument();
  });

  it("rejects an invalid email address", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <LeadFormDialog open onOpenChange={() => {}} orgId="1" onSave={() => {}} />,
    );

    await user.type(await screen.findByPlaceholderText("Nama"), "Acme Inquiry");
    const emailInput = screen.getByLabelText("Email");
    await user.type(emailInput, "not-an-email");
    await user.click(screen.getByRole("button", { name: "Buat prospek" }));

    expect(await screen.findByText("Masukkan alamat email yang valid.")).toBeInTheDocument();
  });

  it("requires a stage for opportunities", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <LeadFormDialog
        open
        onOpenChange={() => {}}
        orgId="1"
        type="opportunity"
        onSave={() => {}}
      />,
    );

    await user.type(await screen.findByPlaceholderText("Nama"), "Big deal");
    await user.click(screen.getByRole("button", { name: "Buat peluang" }));

    expect(await screen.findByText("Pilih tahap.")).toBeInTheDocument();
  });

  it("creates a lead and notifies on save", async () => {
    let createCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/crm/leads", () => {
        createCalls += 1;
        return HttpResponse.json(
          { success: true, message: "Dibuat.", data: { lead: { ...baseLead, id: 9 } } },
          { status: 201 },
        );
      }),
    );
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(<LeadFormDialog open onOpenChange={() => {}} orgId="1" onSave={onSave} />);

    await user.type(await screen.findByPlaceholderText("Nama"), "Acme Inquiry");
    await user.click(screen.getByRole("button", { name: "Buat prospek" }));

    await waitFor(() => expect(createCalls).toBe(1));
    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("updates an existing lead instead of creating", async () => {
    let updateCalls = 0;
    let createCalls = 0;
    server.use(
      http.put("*/api/v1/organizations/:organizationId/crm/leads/:id", () => {
        updateCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: { lead: baseLead } });
      }),
      http.post("*/api/v1/organizations/:organizationId/crm/leads", () => {
        createCalls += 1;
        return HttpResponse.json({ success: true, message: "Dibuat.", data: {} });
      }),
    );
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <LeadFormDialog open onOpenChange={() => {}} orgId="1" initial={baseLead} onSave={onSave} />,
    );

    await screen.findByDisplayValue("Acme Website Inquiry");
    await user.click(screen.getByRole("button", { name: "Simpan perubahan" }));

    await waitFor(() => expect(updateCalls).toBe(1));
    expect(createCalls).toBe(0);
    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });
});
