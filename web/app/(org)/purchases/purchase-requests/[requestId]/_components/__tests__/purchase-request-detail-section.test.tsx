import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PurchaseRequestDetail } from "../purchase-request-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

const request = {
  id: 3,
  organization_id: 1,
  name: "PR-0003",
  requester_id: 9,
  department_id: 2,
  state: "approved",
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
};

const requester = {
  id: 9,
  organization_id: 1,
  name: "Rina Requester",
  display_name: "Rina Requester",
  is_organization: false,
  email: null,
  active: true,
  created_at: STAMP,
  updated_at: STAMP,
};

const department = {
  id: 2,
  organization_id: 1,
  name: "Pengadaan",
  created_at: STAMP,
  updated_at: STAMP,
};

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/purchase-requests/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { request } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [requester] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/departments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { departments: [department] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/approval-requests", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { approvalRequests: [] } }),
    ),
  );
});

describe("PurchaseRequestDetail", () => {
  it("renders the request header with requester and department", async () => {
    renderWithProviders(<PurchaseRequestDetail orgId="1" requestId="3" />);

    expect((await screen.findAllByText("PR-0003")).length).toBeGreaterThan(0);
    expect(screen.getByText("Rina Requester")).toBeInTheDocument();
    expect(screen.getByText("Pengadaan")).toBeInTheDocument();
  });

  it("opens the create QuoteRequest dialog from the action button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PurchaseRequestDetail orgId="1" requestId="3" />);

    await screen.findAllByText("PR-0003");
    await user.click(screen.getByRole("button", { name: "Buat permintaan penawaran" }));

    expect(
      await screen.findByText(
        "Ubah permintaan yang sudah disetujui ini menjadi permintaan penawaran.",
      ),
    ).toBeInTheDocument();
  });
});
