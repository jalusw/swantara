import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { JournalEntriesSection } from "../journal-entries-section";

const baseMove = {
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
};

const journals = [
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
];

function seedMoves(moves: unknown[] = [baseMove], fail = false) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/journal-entries", () => {
      if (fail) return HttpResponse.json({ success: false, message: "Boom." }, { status: 500 });
      return HttpResponse.json({ success: true, message: "OK.", data: { movements: moves } });
    }),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals } }),
    ),
  );
}

async function openCreateDialog(user: ReturnType<typeof userEvent.setup>) {
  await user.click(await screen.findByRole("button", { name: "New entry" }));
  return screen.findByRole("dialog");
}

beforeEach(() => {});

describe("JournalEntriesSection extra2", () => {
  it("shows the journal name and reference for each entry", async () => {
    seedMoves();
    renderWithProviders(<JournalEntriesSection orgId="1" />);

    expect(await screen.findByText("General Journal")).toBeInTheDocument();
    expect(screen.getByText("REF-1")).toBeInTheDocument();
  });

  it("falls back to the journal id when the journal is unknown", async () => {
    seedMoves([{ ...baseMove, journal_id: 99 }]);
    server.use(
      http.get("*/api/v1/organizations/:organizationId/journals", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { journals: [] } }),
      ),
    );
    renderWithProviders(<JournalEntriesSection orgId="1" />);

    expect(await screen.findByText("#99")).toBeInTheDocument();
  });

  it("shows a placeholder when the entry has no reference", async () => {
    seedMoves([{ ...baseMove, ref: null }]);
    renderWithProviders(<JournalEntriesSection orgId="1" />);

    await screen.findByText("JE-001");
    expect(screen.getByText("—")).toBeInTheDocument();
  });

  it("filters entries by state", async () => {
    seedMoves([baseMove, { ...baseMove, id: 2, name: "JE-002", ref: "REF-2", state: "posted" }]);
    const user = userEvent.setup();
    renderWithProviders(<JournalEntriesSection orgId="1" />);

    await screen.findByText("JE-001");
    await user.click(screen.getByRole("button", { name: /Filters/ }));
    await user.selectOptions(screen.getByRole("combobox", { name: "State" }), "posted");

    expect(await screen.findByText("JE-002")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("JE-001")).not.toBeInTheDocument());
  });

  it("retries after a load error", async () => {
    seedMoves([], true);
    const user = userEvent.setup();
    renderWithProviders(<JournalEntriesSection orgId="1" />);

    expect(await screen.findByText("Boom.")).toBeInTheDocument();
    seedMoves();
    await user.click(screen.getByRole("button", { name: "Retry" }));

    expect(await screen.findByText("JE-001")).toBeInTheDocument();
  });

  it("creates a balanced entry from the dialog", async () => {
    seedMoves();
    const createBodies: unknown[] = [];
    server.use(
      http.post("*/api/v1/organizations/:organizationId/journal-entries", async ({ request }) => {
        createBodies.push(await request.json());
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<JournalEntriesSection orgId="1" />);

    const dialog = await openCreateDialog(user);
    await user.click(within(dialog).getByRole("combobox", { name: "Journal" }));
    await user.click(await screen.findByRole("option", { name: "General Journal" }));

    const accountInputs = within(dialog).getAllByPlaceholderText("Account ID");
    await user.type(accountInputs[0]!, "101");
    await user.type(accountInputs[1]!, "201");
    const debits = within(dialog).getAllByRole("spinbutton");
    await user.type(debits[0]!, "100");
    await user.type(debits[1]!, "100");

    expect(await within(dialog).findByText("Balanced")).toBeInTheDocument();
    await user.click(within(dialog).getByRole("button", { name: "Save" }));

    await waitFor(() => expect(createBodies).toHaveLength(1));
    expect(createBodies[0]).toMatchObject({ journal_id: 1 });
  });

  it("keeps save disabled while the entry is unbalanced", async () => {
    seedMoves();
    const user = userEvent.setup();
    renderWithProviders(<JournalEntriesSection orgId="1" />);

    const dialog = await openCreateDialog(user);
    await user.click(within(dialog).getByRole("combobox", { name: "Journal" }));
    await user.click(await screen.findByRole("option", { name: "General Journal" }));

    const accountInputs = within(dialog).getAllByPlaceholderText("Account ID");
    await user.type(accountInputs[0]!, "101");
    await user.type(accountInputs[1]!, "201");
    const debits = within(dialog).getAllByRole("spinbutton");
    await user.type(debits[0]!, "100");

    expect(await within(dialog).findByText("Unbalanced")).toBeInTheDocument();
    expect(within(dialog).getByRole("button", { name: "Save" })).toBeDisabled();
  });

  it("adds and removes entry lines in the dialog", async () => {
    seedMoves();
    const user = userEvent.setup();
    renderWithProviders(<JournalEntriesSection orgId="1" />);

    const dialog = await openCreateDialog(user);
    expect(within(dialog).getAllByPlaceholderText("Account ID")).toHaveLength(2);
    await user.click(within(dialog).getByRole("button", { name: "Add line" }));
    expect(within(dialog).getAllByPlaceholderText("Account ID")).toHaveLength(3);
    await user.click(within(dialog).getAllByRole("button", { name: "Remove" })[0]!);
    expect(within(dialog).getAllByPlaceholderText("Account ID")).toHaveLength(2);
  });

  it("updates the reference and description fields", async () => {
    seedMoves();
    const user = userEvent.setup();
    renderWithProviders(<JournalEntriesSection orgId="1" />);

    const dialog = await openCreateDialog(user);
    expect(dialog).toBeInTheDocument();
    const textInputs = Array.from(dialog.querySelectorAll("input")).filter(
      (element) =>
        (element as HTMLInputElement).type === "text" &&
        !element.hasAttribute("aria-hidden") &&
        !(element as HTMLInputElement).placeholder,
    );
    const referenceInput = textInputs[0] as HTMLInputElement;
    const descriptionInput = textInputs[1] as HTMLInputElement;
    await user.type(referenceInput, "REF-9");
    await user.type(descriptionInput, "Year-end accrual");
    expect(referenceInput.value).toBe("REF-9");
    expect(descriptionInput.value).toBe("Year-end accrual");
  });
});
