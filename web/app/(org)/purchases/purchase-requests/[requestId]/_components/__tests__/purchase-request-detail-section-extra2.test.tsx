import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PurchaseRequestDetail } from "../purchase-request-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

function request(patch: Record<string, unknown> = {}) {
  return {
    id: 3,
    organization_id: 1,
    name: "PR-0003",
    requester_id: 9,
    department_id: 2,
    state: "draft",
    needed_by: "2026-04-01",
    lines: [
      {
        id: 31,
        request_id: 3,
        item_id: 5,
        description: "Widget",
        qty: 4,
        unit_id: null,
        needed_by: null,
      },
    ],
    ...patch,
  };
}

const contacts = [
  {
    id: 9,
    organization_id: 1,
    name: "Rina Requester",
    display_name: "Rina Requester",
    is_organization: false,
    email: null,
    active: true,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

const departments = [
  { id: 2, organization_id: 1, name: "Pengadaan", created_at: STAMP, updated_at: STAMP },
];

function seedDetail(req: unknown, approvals: unknown[] = []) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/purchase-requests/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { request: req } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/departments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { departments } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/approval-requests", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { approvalRequests: approvals } }),
    ),
  );
}

beforeEach(() => {
  seedDetail(request());
});

describe("PurchaseRequestDetail extra2", () => {
  it("shows the not-found state for a missing request", async () => {
    seedDetail(null);
    renderWithProviders(<PurchaseRequestDetail orgId="1" requestId="3" />);

    expect(await screen.findByText("Permintaan pembelian tidak ditemukan.")).toBeInTheDocument();
  });

  it("confirms a draft request", async () => {
    let confirmCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/purchase-requests/:id/confirm", () => {
        confirmCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseRequestDetail orgId="1" requestId="3" />);

    await screen.findAllByText("PR-0003");
    await user.click(screen.getByRole("button", { name: "Konfirmasi" }));

    await waitFor(() => expect(confirmCalls).toBe(1));
  });

  it("approves a confirmed request", async () => {
    let approveCalls = 0;
    seedDetail(request({ state: "confirmed" }));
    server.use(
      http.post("*/api/v1/organizations/:organizationId/purchase-requests/:id/approve", () => {
        approveCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseRequestDetail orgId="1" requestId="3" />);

    await screen.findAllByText("PR-0003");
    expect(screen.queryByRole("button", { name: "Konfirmasi" })).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Setujui" }));

    await waitFor(() => expect(approveCalls).toBe(1));
  });

  it("cancels a draft request", async () => {
    let cancelCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/purchase-requests/:id/cancel", () => {
        cancelCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseRequestDetail orgId="1" requestId="3" />);

    await screen.findAllByText("PR-0003");
    await user.click(screen.getByRole("button", { name: "Batal" }));

    await waitFor(() => expect(cancelCalls).toBe(1));
  });

  it("creates an QuoteRequest from an approved request", async () => {
    let quoteRequestCalls = 0;
    seedDetail(request({ state: "approved" }));
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/supplier-quote-requests/from-request",
        () => {
          quoteRequestCalls += 1;
          return HttpResponse.json({ success: true, message: "OK.", data: {} }, { status: 201 });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<PurchaseRequestDetail orgId="1" requestId="3" />);

    await screen.findAllByText("PR-0003");
    await user.click(screen.getByRole("button", { name: "Buat permintaan penawaran" }));
    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Buat permintaan penawaran" }));

    await waitFor(() => expect(quoteRequestCalls).toBe(1));
  });

  it("falls back for unknown requester and missing dates", async () => {
    seedDetail(request({ requester_id: 4242, department_id: null, needed_by: null }));
    renderWithProviders(<PurchaseRequestDetail orgId="1" requestId="3" />);

    await screen.findAllByText("PR-0003");
    expect(screen.getByText("#4242")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("shows the empty lines state", async () => {
    seedDetail(request({ lines: [] }));
    renderWithProviders(<PurchaseRequestDetail orgId="1" requestId="3" />);

    await screen.findAllByText("PR-0003");
    expect(await screen.findByText("Belum ada baris")).toBeInTheDocument();
  });
});
