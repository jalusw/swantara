import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { OrganizationAdminSection } from "../organization-admin-section";

beforeEach(() => {});

describe("OrganizationAdminSection", () => {
  it("renders the organization profile with the current-org badge", async () => {
    renderWithProviders(<OrganizationAdminSection orgId="1" />);

    expect(await screen.findByText("Profil organisasi")).toBeInTheDocument();
    expect(screen.getByText("Organisasi saat ini")).toBeInTheDocument();
  });

  it("shows the not-found state for an unknown organization", async () => {
    renderWithProviders(<OrganizationAdminSection orgId="999" />);

    expect(await screen.findByText("Organisasi tidak ditemukan.")).toBeInTheDocument();
  });

  it("saves profile changes", async () => {
    const updateCalls: unknown[] = [];
    server.use(
      http.put("*/api/v1/organizations/:id", async ({ request }) => {
        updateCalls.push(await request.json());
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<OrganizationAdminSection orgId="1" />);

    await screen.findByText("Profil organisasi");
    const nameInput = await screen.findByPlaceholderText("Nama");
    await user.clear(nameInput);
    await user.type(nameInput, "Acme Inc");
    await user.type(screen.getByPlaceholderText("Nama badan hukum"), "Acme Inc LLC");
    await user.click(screen.getByRole("combobox", { name: "Organisasi induk" }));
    await user.click(await screen.findByRole("option", { name: "PT Nusantara" }));
    await user.click(screen.getByRole("combobox", { name: "Mata uang dasar" }));
    await user.click(await screen.findByRole("option", { name: "US Dollar (USD)" }));
    await user.click(screen.getByRole("combobox", { name: "Negara" }));
    await user.click(await screen.findByRole("option", { name: "Amerika Serikat" }));
    await user.type(screen.getByPlaceholderText("ID pajak"), "TAX-1");
    await user.click(screen.getByRole("combobox", { name: "Zona waktu" }));
    await user.click(await screen.findByRole("option", { name: "UTC" }));
    await user.click(screen.getByRole("combobox", { name: "Awal tahun pajak" }));
    await user.click(await screen.findByRole("option", { name: "Januari" }));
    await user.click(screen.getByRole("button", { name: "Simpan perubahan" }));

    await waitFor(() => expect(updateCalls).toHaveLength(1));
    expect(updateCalls[0]).toMatchObject({
      name: "Acme Inc",
      legal_name: "Acme Inc LLC",
      parent_id: 2,
      base_currency: "USD",
      country_code: "US",
      tax_id: "TAX-1",
      timezone: "UTC",
      tax_year_start_month: 1,
    });
  });
});
