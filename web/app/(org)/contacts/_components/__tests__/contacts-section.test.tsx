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
    expect(screen.getAllByText("Organisasi").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Perorangan").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Aktif").length).toBeGreaterThan(0);
    expect(screen.getByText("Nonaktif")).toBeInTheDocument();
  });

  it("filters by status", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ContactsSection orgId="1" />);

    await screen.findByText("Bluebird Trading");
    await user.click(screen.getByRole("button", { name: "Filter" }));
    await user.selectOptions(screen.getByLabelText("Saring berdasar status"), "false");

    expect(await screen.findByText("Klima Foods")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Bluebird Trading")).not.toBeInTheDocument());
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ContactsSection orgId="1" />);

    await screen.findByText("Bluebird Trading");
    await user.click(screen.getByRole("button", { name: "Tambah kontak" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Kontak baru")).toBeInTheDocument();
  });
});
