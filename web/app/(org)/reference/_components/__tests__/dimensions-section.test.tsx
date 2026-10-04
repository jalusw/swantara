import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { DimensionsSection } from "../dimensions-section";

beforeEach(() => {});

describe("DimensionsSection", () => {
  it("renders the account tree with seeded accounts", async () => {
    renderWithProviders(<DimensionsSection orgId="1" />);

    expect(await screen.findByText("Operating costs")).toBeInTheDocument();
    expect(screen.getByText("Salaries")).toBeInTheDocument();
    expect(screen.getByText("Revenue")).toBeInTheDocument();
  });

  it("requires a name when adding an account", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DimensionsSection orgId="1" />);

    await user.click(screen.getByRole("button", { name: "Tambah akun" }));
    await user.click(screen.getByRole("button", { name: "Simpan akun" }));

    expect(screen.getByText("Masukkan nama.")).toBeInTheDocument();
  });

  it("adds a new account to the tree", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DimensionsSection orgId="1" />);

    await screen.findByText("Operating costs");
    await user.click(screen.getByRole("button", { name: "Tambah akun" }));
    await user.type(screen.getByLabelText("Nama"), "Travel");
    await user.type(screen.getByLabelText("Kode"), "OPEX-TRV");
    await user.click(screen.getByRole("button", { name: "Simpan akun" }));

    expect(await screen.findByText("Travel")).toBeInTheDocument();
  });

  it("deletes an account after confirmation", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DimensionsSection orgId="1" />);

    await screen.findByText("Operating costs");
    await user.click(screen.getAllByRole("button", { name: "Hapus" })[0]!);
    const dialog = screen.getByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "Hapus" }));

    await waitFor(() => expect(screen.queryByText("Operating costs")).not.toBeInTheDocument());
  });
});
