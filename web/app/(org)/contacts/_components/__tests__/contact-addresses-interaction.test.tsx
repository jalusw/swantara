import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ContactAddresses } from "../contact-addresses";

const addresses = [
  {
    id: 1,
    contact_id: 1,
    type: "billing",
    line1: "Jl. Sudirman 1",
    line2: null,
    city: "Jakarta",
    state: null,
    postal_code: "12190",
    country_code: "ID",
    is_default: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: 2,
    contact_id: 1,
    type: "shipping",
    line1: "Jl. Thamrin 9",
    line2: null,
    city: "Jakarta",
    state: null,
    postal_code: "10350",
    country_code: "ID",
    is_default: false,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

function useAddressHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts/:contactId/addresses", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { addresses } }),
    ),
  );
}

beforeEach(() => {
  useAddressHandlers();
});

describe("ContactAddresses", () => {
  it("renders addresses with type and default badges", async () => {
    renderWithProviders(<ContactAddresses orgId="1" contactId="1" onRefetch={() => {}} />);

    expect(await screen.findByText("Jl. Sudirman 1")).toBeInTheDocument();
    expect(screen.getByText("Jl. Thamrin 9")).toBeInTheDocument();
    expect(screen.getByText("Penagihan")).toBeInTheDocument();
    expect(screen.getByText("Pengiriman")).toBeInTheDocument();
    expect(screen.getByText("Bawaan")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ContactAddresses orgId="1" contactId="1" onRefetch={() => {}} />);

    await screen.findByText("Jl. Sudirman 1");
    await user.click(screen.getByRole("button", { name: "Tambah alamat" }));

    expect(await screen.findByText("Alamat baru")).toBeInTheDocument();
  });
});
