import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { JournalEntryDetailSection } from "../journal-entry-detail-section";

function useLocalMove() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/journal-entries/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          movement: {
            id: 9,
            organization_id: 1,
            journal_id: 1,
            name: "JE-009",
            date: "2026-02-01T00:00:00Z",
            ref: "REF-009",
            state: "posted",
            currency_code: null,
            origin_type: null,
            origin_id: null,
            reversed_entry_id: null,
            posted_at: null,
            posted_by: null,
            created_at: "2026-02-01T00:00:00Z",
            updated_at: "2026-02-01T00:00:00Z",
          },
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
  useLocalMove();
});

describe("JournalEntryDetailSection", () => {
  it("renders move header with journal", async () => {
    renderWithProviders(<JournalEntryDetailSection orgId="1" entryId="9" />);

    expect(await screen.findByText("JE-009")).toBeInTheDocument();
    expect(screen.getByText(/General Journal/)).toBeInTheDocument();
  });

  it("offers reverse action for posted moves", async () => {
    const user = userEvent.setup();
    renderWithProviders(<JournalEntryDetailSection orgId="1" entryId="9" />);

    expect(await screen.findByRole("button", { name: "Balikkan entri" })).toBeInTheDocument();
    await user.hover(screen.getByText("JE-009"));
    expect(screen.getByText("REF-009")).toBeInTheDocument();
  });
});
