import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SessionContext } from "@/providers/session";
import { PosSessionsSection } from "../pos-sessions-section";

const sessions = [
  {
    id: 1,
    config_id: 1,
    cashier_id: 1,
    opened_at: "2026-01-01T08:00:00Z",
    closed_at: null,
    opening_balance: 100,
    closing_balance: null,
    state: "opened",
  },
  {
    id: 2,
    config_id: 2,
    cashier_id: 1,
    opened_at: "2026-01-02T08:00:00Z",
    closed_at: "2026-01-02T18:00:00Z",
    opening_balance: 50,
    closing_balance: 480,
    state: "closed",
  },
];

const configs = [
  {
    id: 1,
    organization_id: 1,
    name: "Main Store",
    warehouse_id: 1,
    journal_id: 1,
    price_book_id: 1,
  },
  {
    id: 2,
    organization_id: 1,
    name: "Outlet Kiosk",
    warehouse_id: 2,
    journal_id: 1,
    price_book_id: null,
  },
];

function useSessionHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/pos/sessions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { sessions } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/pos/configs", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { configs } }),
    ),
  );
}

function renderSessions() {
  return renderWithProviders(
    <SessionContext.Provider
      value={{
        user: { id: 1, email: "alex@acme.com" },
        organizations: [{ id: 1, name: "Acme Inc" }],
        isAuthenticated: true,
        isLoading: false,
      }}
    >
      <PosSessionsSection orgId="1" />
    </SessionContext.Provider>,
  );
}

beforeEach(() => {
  useSessionHandlers();
});

describe("PosSessionsSection", () => {
  it("renders sessions with config names and state badges", async () => {
    renderSessions();

    const links = await screen.findAllByRole("link", { name: /^Sesi \d+$/ });
    expect(links).toHaveLength(2);
    expect(links[0]).toHaveAttribute("href", "/pos/sessions/1");
    expect(links[0]).toHaveTextContent("Sesi 1");
    expect(links[1]).toHaveAttribute("href", "/pos/sessions/2");
    expect(links[1]).toHaveTextContent("Sesi 2");
    expect(screen.getByText("Main Store")).toBeInTheDocument();
    expect(screen.getByText("Outlet Kiosk")).toBeInTheDocument();
  });

  it("filters sessions through the search box", async () => {
    const user = userEvent.setup();
    renderSessions();

    await screen.findAllByRole("link", { name: /^Sesi \d+$/ });
    await user.type(screen.getByPlaceholderText("Cari sesi…"), "2");

    await waitFor(() => {
      const links = screen.getAllByRole("link", { name: /^Sesi \d+$/ });
      expect(links).toHaveLength(1);
      expect(links[0]).toHaveAttribute("href", "/pos/sessions/2");
      expect(links[0]).toHaveTextContent("Sesi 2");
    });
  });

  it("opens the open-session dialog from the action button", async () => {
    const user = userEvent.setup();
    renderSessions();

    await screen.findAllByRole("link", { name: /^Sesi \d+$/ });
    await user.click(screen.getByRole("button", { name: "Buka Sesi" }));

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("Buka Sesi")).toBeInTheDocument();
  });
});
