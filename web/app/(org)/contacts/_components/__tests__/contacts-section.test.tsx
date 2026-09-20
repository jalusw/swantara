import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { ContactsSection } from "../contacts-section";

beforeEach(() => {});

describe("ContactsSection", () => {
  it("renders contacts with type and status badges", async () => {
    renderWithProviders(<ContactsSection orgId="1" />);

    expect(await screen.findByText("Bluebird Trading")).toBeInTheDocument();
    expect(screen.getByText("Nusantara Logistics")).toBeInTheDocument();
    expect(screen.getByText("Klima Foods")).toBeInTheDocument();
    expect(screen.getAllByText("Organization").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Individual").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Active").length).toBeGreaterThan(0);
    expect(screen.getByText("Inactive")).toBeInTheDocument();
  });

  it("filters by status", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ContactsSection orgId="1" />);

    await screen.findByText("Bluebird Trading");
    await user.click(screen.getByRole("button", { name: "Filters" }));
    await user.selectOptions(screen.getByLabelText("Filter by status"), "false");

    expect(await screen.findByText("Klima Foods")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Bluebird Trading")).not.toBeInTheDocument());
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ContactsSection orgId="1" />);

    await screen.findByText("Bluebird Trading");
    await user.click(screen.getByRole("button", { name: "Add contact" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("New contact")).toBeInTheDocument();
  });
});
