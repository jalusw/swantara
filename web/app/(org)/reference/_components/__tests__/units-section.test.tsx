import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { UnitsSection } from "../units-section";

beforeEach(() => {});

describe("UnitsSection", () => {
  it("renders categories and units of measure", async () => {
    renderWithProviders(<UnitsSection orgId="1" />);

    expect(await screen.findByText("Meter")).toBeInTheDocument();
    expect(screen.getByText("Kilogram")).toBeInTheDocument();
    expect(screen.getAllByText("Length").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Weight").length).toBeGreaterThan(0);
  });

  it("adds a category", async () => {
    const user = userEvent.setup();
    renderWithProviders(<UnitsSection orgId="1" />);

    await screen.findByText("Meter");
    await user.click(screen.getByRole("button", { name: "Tambah kategori" }));
    await user.type(screen.getByLabelText("Nama"), "Temperature");
    await user.click(screen.getByRole("button", { name: "Simpan kategori" }));

    expect(await screen.findByText("Temperature")).toBeInTheDocument();
  });

  it("requires a name when adding a UoM", async () => {
    const user = userEvent.setup();
    renderWithProviders(<UnitsSection orgId="1" />);

    await user.click(screen.getByRole("button", { name: "Tambah satuan" }));
    await user.click(screen.getByRole("button", { name: "Simpan satuan" }));

    expect(screen.getByText("Masukkan nama.")).toBeInTheDocument();
  });

  it("adds a new unit of measure", async () => {
    const user = userEvent.setup();
    renderWithProviders(<UnitsSection orgId="1" />);

    await screen.findByText("Meter");
    await user.click(screen.getByRole("button", { name: "Tambah satuan" }));
    await user.click(screen.getByLabelText("Kategori"));
    await user.click(await screen.findByRole("option", { name: "Volume" }));
    await waitFor(() =>
      expect(screen.queryByRole("option", { name: "Volume" })).not.toBeInTheDocument(),
    );
    await user.type(screen.getByLabelText("Nama"), "Gallon");
    await user.type(screen.getByLabelText("Faktor"), "3.785");
    await user.click(screen.getByRole("button", { name: "Simpan satuan" }));

    expect(await within(screen.getByRole("table")).findByText("Gallon")).toBeInTheDocument();
  });

  it("deletes a unit of measure after confirmation", async () => {
    const user = userEvent.setup();
    renderWithProviders(<UnitsSection orgId="1" />);

    await screen.findByText("Meter");
    await user.click(screen.getAllByRole("button", { name: "Hapus" })[0]!);
    const dialog = screen.getByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "Hapus" }));

    await waitFor(() => expect(screen.queryByText("Meter")).not.toBeInTheDocument());
  });
});
