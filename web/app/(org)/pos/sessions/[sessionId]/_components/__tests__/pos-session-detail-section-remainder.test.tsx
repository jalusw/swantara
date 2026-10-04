import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PosSessionDetail } from "../pos-session-detail-section";

function sessionFixture(overrides = {}) {
  return {
    id: 1,
    config_id: 5,
    cashier_id: 7,
    opened_at: "2026-01-01T08:00:00Z",
    closed_at: null,
    opening_balance: 100000,
    closing_balance: null,
    state: "opened",
    orders: [],
    ...overrides,
  };
}

function useSessionHandlers(session: Record<string, unknown>) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/pos/sessions/:sessionId", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { session } }),
    ),
  );
}

beforeEach(() => {});

describe("PosSessionDetail remainder", () => {
  it("shows the not-found state when the session is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/pos/sessions/:sessionId", () =>
        HttpResponse.json({ success: false, message: "Not found." }, { status: 404 }),
      ),
    );
    renderWithProviders(<PosSessionDetail orgId="1" sessionId="999" />);

    expect(await screen.findByText("Sesi tidak ditemukan.")).toBeInTheDocument();
  });

  it("shows the empty orders and payments states", async () => {
    useSessionHandlers(sessionFixture());
    const user = userEvent.setup();
    renderWithProviders(<PosSessionDetail orgId="1" sessionId="1" />);

    await screen.findByRole("heading", { name: "Sesi 1" });
    expect(screen.getByText("Tidak ada pembayaran")).toBeInTheDocument();

    await user.click(screen.getByRole("tab", { name: "Pesanan" }));
    expect(await screen.findByText("Tidak ada pesanan sesi")).toBeInTheDocument();
  });

  it("moves an opened session to closing", async () => {
    let closingCalled = false;
    useSessionHandlers(sessionFixture());
    server.use(
      http.post("*/api/v1/organizations/:organizationId/pos/sessions/:id/closing", () => {
        closingCalled = true;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<PosSessionDetail orgId="1" sessionId="1" />);

    await screen.findByRole("heading", { name: "Sesi 1" });
    await user.click(screen.getByRole("button", { name: "Mulai penutupan" }));

    await waitFor(() => expect(closingCalled).toBe(true));
  });

  it("closes an opened session through the dialog with a balanced amount", async () => {
    let closedBalance: unknown = null;
    useSessionHandlers(
      sessionFixture({
        orders: [
          {
            id: 11,
            session_id: 1,
            name: "POS-001",
            amount_total: 50000,
            amount_tax: 5000,
            state: "done",
            order_time: "2026-01-02T10:00:00Z",
            payments: [{ method: "cash", amount: 50000 }],
          },
        ],
      }),
    );
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/pos/sessions/:id/close",
        async ({ request }) => {
          closedBalance = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<PosSessionDetail orgId="1" sessionId="1" />);

    await screen.findByRole("heading", { name: "Sesi 1" });
    expect(screen.getByText("cash")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Tutup Sesi" }));

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("Balanced")).toBeInTheDocument();
    await user.click(within(dialog).getByRole("button", { name: "Tutup Sesi" }));

    await waitFor(() => expect(closedBalance).toMatchObject({ closing_balance: 150000 }));
  });

  it("shows the difference when the closing balance is edited", async () => {
    useSessionHandlers(sessionFixture());
    const user = userEvent.setup();
    renderWithProviders(<PosSessionDetail orgId="1" sessionId="1" />);

    await screen.findByRole("heading", { name: "Sesi 1" });
    await user.click(screen.getByRole("button", { name: "Tutup Sesi" }));

    const dialog = await screen.findByRole("dialog");
    const input = within(dialog).getByRole("spinbutton");
    await user.clear(input);
    await user.type(input, "1");

    expect(await within(dialog).findByText(/Difference/)).toBeInTheDocument();
  });

  it("hides session actions once the session is closed", async () => {
    useSessionHandlers(
      sessionFixture({
        state: "closed",
        closed_at: "2026-01-03T10:00:00Z",
        closing_balance: 150000,
      }),
    );
    renderWithProviders(<PosSessionDetail orgId="1" sessionId="1" />);

    await screen.findByRole("heading", { name: "Sesi 1" });
    expect(screen.queryByRole("button", { name: "Buka Kasir" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Mulai penutupan" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Tutup Sesi" })).not.toBeInTheDocument();
  });

  it("shows an error toast when closing fails", async () => {
    useSessionHandlers(sessionFixture());
    server.use(
      http.post("*/api/v1/organizations/:organizationId/pos/sessions/:id/closing", () =>
        HttpResponse.json({ success: false, message: "Cannot close." }, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<PosSessionDetail orgId="1" sessionId="1" />);

    await screen.findByRole("heading", { name: "Sesi 1" });
    await user.click(screen.getByRole("button", { name: "Mulai penutupan" }));

    expect(await screen.findByRole("heading", { name: "Sesi 1" })).toBeInTheDocument();
  });
});
