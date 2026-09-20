import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { ContactAddresses } from "../contact-addresses";

describe("ContactAddresses", () => {
  it("renders empty state", () => {
    renderWithProviders(<ContactAddresses orgId="1" contactId="1" onRefetch={() => {}} />);

    expect(screen.getByText("No addresses")).toBeInTheDocument();
  });
});
