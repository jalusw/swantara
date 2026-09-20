import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { PaymentTermsSection } from "../payment-terms-section";

beforeEach(() => {});

async function addPercentTerm(
  user: ReturnType<typeof userEvent.setup>,
  name: string,
  values: string[],
) {
  await user.click(screen.getByRole("button", { name: "Add term" }));
  await user.type(screen.getByLabelText("Name"), name);
  for (const [index, value] of values.entries()) {
    if (index > 0) {
      await user.click(screen.getByRole("button", { name: "Add line" }));
    }
    await user.type(screen.getAllByLabelText("Value")[index]!, value);
    await user.type(screen.getAllByLabelText("Due days")[index]!, "30");
  }
  await user.click(screen.getByRole("button", { name: "Save term" }));
}

describe("PaymentTermsSection", () => {
  it("renders the seeded payment terms", async () => {
    renderWithProviders(<PaymentTermsSection orgId="1" />);

    expect(await screen.findByText("Net 30")).toBeInTheDocument();
    expect(screen.getByText("50/50 split")).toBeInTheDocument();
  });

  it("requires a name when creating a term", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PaymentTermsSection orgId="1" />);

    await user.click(screen.getByRole("button", { name: "Add term" }));
    await user.click(screen.getByRole("button", { name: "Save term" }));

    expect(screen.getByText("Enter a name.")).toBeInTheDocument();
  });

  it("rejects percent lines that do not add up to 100", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PaymentTermsSection orgId="1" />);

    await addPercentTerm(user, "60/60", ["60", "60"]);

    expect(
      screen.getByText("Percent lines must add up to 100%. Current total: 120%."),
    ).toBeInTheDocument();
  });

  it("saves a term whose percent lines add up to 100", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PaymentTermsSection orgId="1" />);

    await screen.findByText("Net 30");
    await addPercentTerm(user, "60/40", ["60", "40"]);

    expect(await screen.findByText("60/40")).toBeInTheDocument();
  });

  it("deletes a payment term after confirmation", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PaymentTermsSection orgId="1" />);

    await screen.findByText("Net 30");
    await user.click(screen.getAllByRole("button", { name: "Delete" })[0]!);
    const dialog = screen.getByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "Delete" }));

    await waitFor(() => expect(screen.queryByText("Net 30")).not.toBeInTheDocument());
  });
});
