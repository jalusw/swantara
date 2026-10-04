import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { AuditLogsSection } from "../audit-logs-section";

const STAMP = "2026-01-01T00:00:00Z";

const auditLogs = [
  {
    id: 1,
    table_name: "invoices",
    record_id: 41,
    action: "insert",
    changed_by: 9,
    changed_at: STAMP,
    diff: {},
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 2,
    table_name: "contacts",
    record_id: 7,
    action: "update",
    changed_by: 9,
    changed_at: STAMP,
    diff: {},
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 3,
    table_name: "invoices",
    record_id: 42,
    action: "delete",
    changed_by: 10,
    changed_at: STAMP,
    diff: {},
    created_at: STAMP,
    updated_at: STAMP,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/audit-logs", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { audit_logs: auditLogs } }),
    ),
  );
});

describe("AuditLogsSection", () => {
  it("renders seeded audit logs with action tones", async () => {
    renderWithProviders(<AuditLogsSection orgId="1" />);

    expect(await screen.findByText("AL-1")).toBeInTheDocument();
    expect(screen.getAllByText("invoices").length).toBeGreaterThan(0);
    expect(screen.getByText("contacts")).toBeInTheDocument();
    expect(screen.getAllByText("Dibuat").length).toBeGreaterThan(0);
    expect(screen.getByText("Diperbarui")).toBeInTheDocument();
    expect(screen.getByText("Dihapus")).toBeInTheDocument();
  });

  it("filters logs by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<AuditLogsSection orgId="1" />);

    await screen.findByText("AL-1");
    await user.type(screen.getByPlaceholderText("Cari log audit…"), "contacts");

    expect(await screen.findByText("AL-2")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("AL-1")).not.toBeInTheDocument());
  });

  it("shows the empty state when there are no logs", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/audit-logs", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { audit_logs: [] } }),
      ),
    );
    renderWithProviders(<AuditLogsSection orgId="1" />);

    expect(await screen.findByText("Belum ada log audit")).toBeInTheDocument();
  });

  it("shows the error state with retry and refetches", async () => {
    let calls = 0;
    server.use(
      http.get("*/api/v1/organizations/:organizationId/audit-logs", () => {
        calls += 1;
        if (calls === 1) {
          return HttpResponse.json({ success: false, message: "audit boom" }, { status: 500 });
        }
        return HttpResponse.json({
          success: true,
          message: "OK.",
          data: { audit_logs: auditLogs },
        });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<AuditLogsSection orgId="1" />);

    expect(await screen.findByText("audit boom")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Coba lagi" }));

    expect(await screen.findByText("AL-1")).toBeInTheDocument();
    expect(calls).toBe(2);
  });
});
