import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { ContactBankAccounts } from "../contact-bank-accounts";

describe("ContactBankAccounts", () => {
  it("renders empty state", () => {
    renderWithProviders(<ContactBankAccounts orgId="1" contactId="1" onRefetch={() => {}} />);

    expect(screen.getByText("No bank accounts")).toBeInTheDocument();
  });
});
