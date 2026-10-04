import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { UnitsSection } from "../units-section";

beforeEach(() => {});

describe("UnitsSection remainder", () => {
  it("shows the empty categories message when no categories exist", async () => {
    server.use(
      http.get("*/api/v1/unit-categories", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { categories: [] } }),
      ),
    );
    renderWithProviders(<UnitsSection orgId="1" />);

    expect(await screen.findByText("Belum ada kategori")).toBeInTheDocument();
  });

  it("shows the error state with retry when loading fails", async () => {
    server.use(
      http.get("*/api/v1/units", () =>
        HttpResponse.json({ success: false, message: "UoM load failed." }, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<UnitsSection orgId="1" />);

    expect(await screen.findByText("UoM load failed.")).toBeInTheDocument();

    server.resetHandlers();
    await user.click(screen.getByRole("button", { name: "Coba lagi" }));

    expect(await screen.findByText("Meter")).toBeInTheDocument();
  });

  it("opens the create dialog and saves a new unit", async () => {
    let created: unknown = null;
    server.use(
      http.post("*/api/v1/units", async ({ request }) => {
        created = await request.json();
        return HttpResponse.json(
          {
            success: true,
            message: "Dibuat.",
            data: {
              unit: {
                id: 9,
                category_id: 1,
                name: "Centimeter",
                factor: 0.01,
                unit_type: "smaller",
                rounding: 0.001,
                created_at: "2026-01-01T00:00:00Z",
                updated_at: "2026-01-01T00:00:00Z",
              },
            },
          },
          { status: 201 },
        );
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<UnitsSection orgId="1" />);

    await screen.findByText("Meter");
    await user.click(screen.getByRole("button", { name: "Tambah satuan" }));

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("Satuan ukur baru")).toBeInTheDocument();

    await user.click(within(dialog).getByLabelText("Kategori"));
    await user.click(await screen.findByRole("option", { name: "Length" }));
    await waitFor(() =>
      expect(screen.queryByRole("option", { name: "Length" })).not.toBeInTheDocument(),
    );
    await user.type(within(dialog).getByLabelText("Nama"), "Centimeter");
    await user.type(within(dialog).getByLabelText("Faktor"), "0.01");
    await user.click(within(dialog).getByRole("button", { name: "Simpan satuan" }));

    await waitFor(() => expect(created).toMatchObject({ name: "Centimeter" }));
  });

  it("opens the edit dialog prefilled and updates the unit", async () => {
    let updated: unknown = null;
    server.use(
      http.put("*/api/v1/units/:unitId", async ({ request }) => {
        updated = await request.json();
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<UnitsSection orgId="1" />);

    await screen.findByText("Meter");
    const table = screen.getByRole("table");
    await user.click(within(table).getAllByRole("button", { name: "Ubah" })[0]!);

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("Ubah satuan ukur")).toBeInTheDocument();
    expect(within(dialog).getByDisplayValue("Meter")).toBeInTheDocument();

    await user.click(within(dialog).getByRole("button", { name: "Simpan satuan" }));

    await waitFor(() => expect(updated).toMatchObject({ name: "Meter" }));
  });

  it("creates a category from the category dialog", async () => {
    let created: unknown = null;
    server.use(
      http.post("*/api/v1/unit-categories", async ({ request }) => {
        created = await request.json();
        return HttpResponse.json(
          {
            success: true,
            message: "Dibuat.",
            data: { category: { id: 9, name: "Time" } },
          },
          { status: 201 },
        );
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<UnitsSection orgId="1" />);

    await screen.findByText("Meter");
    await user.click(screen.getByRole("button", { name: "Tambah kategori" }));

    const dialog = await screen.findByRole("dialog");
    await user.type(within(dialog).getByRole("textbox"), "Time");
    await user.click(within(dialog).getByRole("button", { name: "Simpan kategori" }));

    await waitFor(() => expect(created).toMatchObject({ name: "Time" }));
  });

  it("deletes a unit through the row actions", async () => {
    let deleted = false;
    server.use(
      http.delete("*/api/v1/units/:unitId", () => {
        deleted = true;
        return HttpResponse.json({ success: true, message: "Deleted." });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<UnitsSection orgId="1" />);

    await screen.findByText("Meter");
    const table = screen.getByRole("table");
    await user.click(within(table).getAllByRole("button", { name: "Hapus" })[0]!);

    const confirm = await screen.findByRole("alertdialog");
    await user.click(within(confirm).getByRole("button", { name: "Hapus" }));

    await waitFor(() => expect(deleted).toBe(true));
  });
});
