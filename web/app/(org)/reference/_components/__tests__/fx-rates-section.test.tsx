import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { FxRatesSection } from "../fx-rates-section";

beforeEach(() => {});

describe("FxRatesSection", () => {
  it("renders the seeded exchange rates", async () => {
    renderWithProviders(<FxRatesSection orgId="1" />);

    expect(await screen.findByText("IDR")).toBeInTheDocument();
    expect(screen.getByText("EUR")).toBeInTheDocument();
    expect(screen.getAllByText("Spot").length).toBeGreaterThan(0);
    expect(screen.getByText("Average")).toBeInTheDocument();
  });

  it("requires a rate value when creating", async () => {
    const user = userEvent.setup();
    renderWithProviders(<FxRatesSection orgId="1" />);

    await user.click(screen.getByRole("button", { name: "Add rate" }));
    await user.type(screen.getByLabelText("Valid from"), "2026-09-01");
    await user.click(screen.getByRole("button", { name: "Save rate" }));

    expect(screen.getByText("Enter a rate.")).toBeInTheDocument();
  });

  it("adds a new exchange rate", async () => {
    const user = userEvent.setup();
    renderWithProviders(<FxRatesSection orgId="1" />);

    await screen.findByText("IDR");
    await user.click(screen.getByRole("button", { name: "Add rate" }));
    await user.click(screen.getByLabelText("Currency"));
    await user.click(await screen.findByRole("option", { name: "JPY" }));
    await waitFor(() =>
      expect(screen.queryByRole("option", { name: "JPY" })).not.toBeInTheDocument(),
    );
    await user.type(screen.getByLabelText("Rate"), "0.35");
    await user.type(screen.getByLabelText("Valid from"), "2026-09-01");
    await user.click(screen.getByRole("button", { name: "Save rate" }));

    expect(await screen.findByText("JPY")).toBeInTheDocument();
  });

  it("deletes an exchange rate after confirmation", async () => {
    const user = userEvent.setup();
    renderWithProviders(<FxRatesSection orgId="1" />);

    await screen.findByText("IDR");
    await user.click(screen.getAllByRole("button", { name: "Delete" })[0]!);
    const dialog = screen.getByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "Delete" }));

    await waitFor(() => expect(screen.queryByText("IDR")).not.toBeInTheDocument());
  });
});
