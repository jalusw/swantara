import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { JournalsSection } from "../journals-section";

const journals = [
  {
    id: 1,
    organization_id: 1,
    name: "Sales Journal",
    code: "SAJ",
    type: "sale",
    default_account_id: 1,
    bank_account_id: null,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: 2,
    organization_id: 1,
    name: "Misc Journal",
    code: null,
    type: "general",
    default_account_id: null,
    bank_account_id: null,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

const accounts = [
  {
    id: 1,
    organization_id: 1,
    code: "1000",
    name: "Kas",
    type: "cash",
    reconcilable: false,
    currency_code: null,
    parent_id: null,
    active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

function useJournalsHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/accounts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { accounts } }),
    ),
  );
}

beforeEach(() => {
  useJournalsHandlers();
});

describe("JournalsSection remainder", () => {
  it("renders a dash for journals without code or default account", async () => {
    renderWithProviders(<JournalsSection orgId="1" />);

    await screen.findByText("Sales Journal");
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("shows the error state with retry when loading fails", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/journals", () =>
        HttpResponse.json({ success: false, message: "Journal load failed." }, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<JournalsSection orgId="1" />);

    expect(await screen.findByText("Journal load failed.")).toBeInTheDocument();

    useJournalsHandlers();
    await user.click(screen.getByRole("button", { name: "Coba lagi" }));

    expect(await screen.findByText("Sales Journal")).toBeInTheDocument();
  });

  it("opens the create dialog with an empty name and disabled save", async () => {
    const user = userEvent.setup();
    renderWithProviders(<JournalsSection orgId="1" />);

    await screen.findByText("Sales Journal");
    await user.click(screen.getByRole("button", { name: /tambah jurnal/i }));

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("Buat jurnal")).toBeInTheDocument();
    expect(within(dialog).getByRole("button", { name: "Simpan" })).toBeDisabled();
  });

  it("creates a journal from the dialog", async () => {
    let created: unknown = null;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/journals", async ({ request }) => {
        created = await request.json();
        return HttpResponse.json(
          { success: true, message: "Dibuat.", data: { journal: journals[0] } },
          { status: 201 },
        );
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<JournalsSection orgId="1" />);

    await screen.findByText("Sales Journal");
    await user.click(screen.getByRole("button", { name: /tambah jurnal/i }));

    const dialog = await screen.findByRole("dialog");
    const nameInput = within(dialog).getAllByRole("textbox")[0]!;
    await user.type(nameInput, "Petty Cash");

    await user.click(within(dialog).getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(created).toMatchObject({ name: "Petty Cash" }));
  });

  it("opens the edit dialog prefilled with the journal values", async () => {
    const user = userEvent.setup();
    renderWithProviders(<JournalsSection orgId="1" />);

    await screen.findByText("Sales Journal");
    const table = screen.getByRole("table");
    await user.click(within(table).getAllByRole("button", { name: "Ubah" })[0]!);

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("Ubah jurnal")).toBeInTheDocument();
    expect(within(dialog).getByDisplayValue("Sales Journal")).toBeInTheDocument();
  });

  it("deletes a journal through the row actions", async () => {
    let deletedId: number | null = null;
    server.use(
      http.delete("*/api/v1/organizations/:organizationId/journals/:id", ({ params }) => {
        deletedId = Number(params.id);
        return HttpResponse.json({ success: true, message: "Deleted." });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<JournalsSection orgId="1" />);

    await screen.findByText("Sales Journal");
    const table = screen.getByRole("table");
    await user.click(within(table).getAllByRole("button", { name: "Hapus" })[0]!);

    const confirm = await screen.findByRole("alertdialog");
    await user.click(within(confirm).getByRole("button", { name: "Hapus" }));

    await waitFor(() => expect(deletedId).toBe(1));
  });
});
