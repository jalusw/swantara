import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ExpenseCategoriesSection } from "../expense-categories-section";

function useLocalCategories() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/expense-categories", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          categories: [
            {
              id: 1,
              organization_id: 1,
              name: "Travel",
              expense_account_id: null,
              default_tax_ids: [],
            },
            {
              id: 2,
              organization_id: 1,
              name: "Meals",
              expense_account_id: null,
              default_tax_ids: [],
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalCategories();
});

describe("ExpenseCategoriesSection", () => {
  it("renders seeded expense categories", async () => {
    renderWithProviders(<ExpenseCategoriesSection orgId="1" />);

    expect(await screen.findByText("Travel")).toBeInTheDocument();
    expect(screen.getByText("Meals")).toBeInTheDocument();
  });

  it("filters categories by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ExpenseCategoriesSection orgId="1" />);

    await screen.findByText("Travel");
    await user.type(screen.getByPlaceholderText("Cari kategori"), "Meals");

    expect((await screen.findAllByText("Meals")).length).toBeGreaterThan(0);
    await waitFor(() => expect(screen.queryByText("Travel")).not.toBeInTheDocument());
  });
});
