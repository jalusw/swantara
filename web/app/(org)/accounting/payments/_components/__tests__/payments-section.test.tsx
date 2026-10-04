import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PaymentsSection } from "../payments-section";

function useLocalPayments() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/payments", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          payments: [
            {
              id: 1,
              organization_id: 1,
              name: "PAY-001",
              contact_id: 10,
              type: "inbound",
              journal_id: 1,
              payment_method: null,
              amount: 1500,
              currency_code: null,
              date: "2026-02-10T00:00:00Z",
              reference: null,
              entry_id: null,
              contact_bank_account_id: null,
              state: "posted",
              created_at: "2026-02-10T00:00:00Z",
              updated_at: "2026-02-10T00:00:00Z",
            },
            {
              id: 2,
              organization_id: 1,
              name: "PAY-002",
              contact_id: 11,
              type: "outbound",
              journal_id: 1,
              payment_method: null,
              amount: 800,
              currency_code: null,
              date: "2026-02-11T00:00:00Z",
              reference: null,
              entry_id: null,
              contact_bank_account_id: null,
              state: "draft",
              created_at: "2026-02-11T00:00:00Z",
              updated_at: "2026-02-11T00:00:00Z",
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalPayments();
});

describe("PaymentsSection", () => {
  it("renders seeded payments", async () => {
    renderWithProviders(<PaymentsSection />);

    expect(await screen.findByText("PAY-001")).toBeInTheDocument();
    expect(screen.getByText("PAY-002")).toBeInTheDocument();
  });

  it("filters payments by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PaymentsSection />);

    await screen.findByText("PAY-001");
    await user.type(screen.getByPlaceholderText("Cari pembayaran..."), "002");

    expect(await screen.findByText("PAY-002")).toBeInTheDocument();
  });
});
