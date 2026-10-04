import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { CustomersAcquisition } from "../customers-acquisition-section";

const contacts = [
  {
    id: 1,
    organization_id: 1,
    name: "Acme Corporation",
    display_name: "Acme Corp",
    is_organization: true,
    parent_id: null,
    email: "hello@acme.co",
    phone: null,
    mobile: null,
    website: null,
    tax_id: null,
    industry: null,
    currency_code: "USD",
    lang: "en",
    active: true,
    created_at: "2026-01-05T00:00:00Z",
    updated_at: "2026-01-05T00:00:00Z",
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts } }),
    ),
  );
});

describe("CustomersAcquisition", () => {
  it("renders the acquisition trend chart", async () => {
    renderWithProviders(<CustomersAcquisition />);

    expect(await screen.findByText("Akuisisi pelanggan")).toBeInTheDocument();
    expect(await screen.findByRole("img", { name: /Akuisisi pelanggan/ })).toBeInTheDocument();
  });

  it("retries loading after a failure", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/contacts", () =>
        HttpResponse.json({ success: false }, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<CustomersAcquisition />);

    expect(await screen.findByText("Gagal memuat")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Coba lagi" }));

    expect(await screen.findByText("Gagal memuat")).toBeInTheDocument();
  });
});
