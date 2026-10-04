import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PaymentDetailSection } from "../payment-detail-section";

function useLocalPayment() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/payments/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          payment: {
            id: 3,
            organization_id: 1,
            name: "PAY-003",
            contact_id: 10,
            type: "inbound",
            journal_id: 1,
            payment_method: null,
            amount: 2500,
            currency_code: null,
            date: "2026-02-10T00:00:00Z",
            reference: null,
            entry_id: null,
            contact_bank_account_id: null,
            state: "posted",
            created_at: "2026-02-10T00:00:00Z",
            updated_at: "2026-02-10T00:00:00Z",
          },
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalPayment();
});

describe("PaymentDetailSection", () => {
  it("renders payment header with amount", async () => {
    renderWithProviders(<PaymentDetailSection orgId="1" paymentId="3" />);

    expect(await screen.findByText("PAY-003")).toBeInTheDocument();
    expect(screen.getByText("2.500,00")).toBeInTheDocument();
  });

  it("renders payment direction", async () => {
    renderWithProviders(<PaymentDetailSection orgId="1" paymentId="3" />);

    await screen.findByText("PAY-003");
    expect(screen.getByText("Masuk")).toBeInTheDocument();
  });
});
