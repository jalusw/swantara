import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { JournalEntriesSection } from "../journal-entries-section";

function useLocalMoves() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/journal-entries", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          movements: [
            {
              id: 1,
              organization_id: 1,
              journal_id: 1,
              name: "JE-001",
              date: "2026-02-01T00:00:00Z",
              ref: "REF-1",
              state: "draft",
              currency_code: null,
              origin_type: null,
              origin_id: null,
              reversed_entry_id: null,
              posted_at: null,
              posted_by: null,
              created_at: "2026-02-01T00:00:00Z",
              updated_at: "2026-02-01T00:00:00Z",
            },
            {
              id: 2,
              organization_id: 1,
              journal_id: 1,
              name: "JE-002",
              date: "2026-02-02T00:00:00Z",
              ref: "REF-2",
              state: "posted",
              currency_code: null,
              origin_type: null,
              origin_id: null,
              reversed_entry_id: null,
              posted_at: null,
              posted_by: null,
              created_at: "2026-02-02T00:00:00Z",
              updated_at: "2026-02-02T00:00:00Z",
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          journals: [
            {
              id: 1,
              organization_id: 1,
              name: "General Journal",
              code: "GEN",
              type: "general",
              default_account_id: null,
              bank_account_id: null,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalMoves();
});

describe("JournalEntriesSection", () => {
  it("renders seeded entries", async () => {
    renderWithProviders(<JournalEntriesSection orgId="1" />);

    expect(await screen.findByText("JE-001")).toBeInTheDocument();
    expect(screen.getByText("JE-002")).toBeInTheDocument();
  });

  it("filters entries by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<JournalEntriesSection orgId="1" />);

    await screen.findByText("JE-001");
    await user.type(screen.getByPlaceholderText("Search journal entries..."), "002");

    expect(await screen.findByText("JE-002")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<JournalEntriesSection orgId="1" />);

    await screen.findByText("JE-001");
    await user.click(screen.getByRole("button", { name: "New entry" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Create journal entry")).toBeInTheDocument();
  });
});
