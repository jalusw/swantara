import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { DashboardOverview } from "../_components/dashboard-overview-section";

describe("DashboardOverview", () => {
  it("renders stats, revenue overview and greeting for the active org", async () => {
    renderWithProviders(<DashboardOverview orgId="1" />);

    expect((await screen.findAllByText("Rp 84.320,00")).length).toBeGreaterThan(0);
    expect(await screen.findByText("1.284")).toBeInTheDocument();
    expect((await screen.findAllByText("96")).length).toBeGreaterThan(0);
    expect((await screen.findAllByText("48")).length).toBeGreaterThan(0);
    expect(await screen.findByText("Halo, Alex!")).toBeInTheDocument();
    expect(screen.getByText("Ringkasan pendapatan")).toBeInTheDocument();
  });

  it("opens the new record menu and shows the export button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DashboardOverview orgId="1" />);

    expect((await screen.findAllByText("Rp 84.320,00")).length).toBeGreaterThan(0);
    await user.click(screen.getByRole("button", { name: "Buat cepat" }));
    expect(await screen.findByRole("menuitem", { name: "Pelanggan baru" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Faktur baru" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ekspor laporan/i })).toBeInTheDocument();
  });

  it("renders service coverage grid with all business flows", async () => {
    renderWithProviders(<DashboardOverview orgId="1" />);

    expect(await screen.findByText("Jelajahi alur bisnis")).toBeInTheDocument();
    expect(await screen.findByText("CRM")).toBeInTheDocument();
    expect(
      screen.getByText("Setiap layanan organisasi Anda — satu ketuk ke modulnya."),
    ).toBeInTheDocument();
    expect(screen.getByText("Point of Sale")).toBeInTheDocument();
    expect(screen.getByText("Aset Tetap")).toBeInTheDocument();
  });

  it("covers all quick-create actions for every service", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DashboardOverview orgId="1" />);

    expect((await screen.findAllByText("Rp 84.320,00")).length).toBeGreaterThan(0);
    await user.click(screen.getByRole("button", { name: "Buat cepat" }));
    expect(await screen.findByRole("menuitem", { name: "Pemasok baru" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Produk baru" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Pesanan penjualan baru" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Pesanan pembelian baru" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Perintah produksi baru" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Proyek baru" })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Karyawan baru" })).toBeInTheDocument();
  });
});
