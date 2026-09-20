import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { OrganizationAdminSection } from "../organization-admin-section";

beforeEach(() => {});

describe("OrganizationAdminSection", () => {
  it("renders the organization profile with the current-org badge", async () => {
    renderWithProviders(<OrganizationAdminSection orgId="1" />);

    expect(await screen.findByText("Organization profile")).toBeInTheDocument();
    expect(screen.getByText("Current organization")).toBeInTheDocument();
    expect(screen.getByText("Danger zone")).toBeInTheDocument();
  });

  it("shows the not-found state for an unknown organization", async () => {
    renderWithProviders(<OrganizationAdminSection orgId="999" />);

    expect(await screen.findByText("Organization not found.")).toBeInTheDocument();
  });

  it("disables the danger action for the active organization", async () => {
    renderWithProviders(<OrganizationAdminSection orgId="1" />);

    await screen.findByText("Organization profile");
    expect(
      screen.getByText("You cannot disable the organization you are currently working in."),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Disable organization" })).toBeDisabled();
  });

  it("enables the danger action for a non-active organization", async () => {
    renderWithProviders(<OrganizationAdminSection orgId="2" />);

    await screen.findByText("Organization profile");
    expect(screen.queryByText("Current organization")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Disable organization" })).toBeEnabled();
  });

  it("saves profile changes", async () => {
    const updateCalls: unknown[] = [];
    server.use(
      http.put("*/api/v1/organizations/:id", async ({ request }) => {
        updateCalls.push(await request.json());
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<OrganizationAdminSection orgId="1" />);

    await screen.findByText("Organization profile");
    await user.type(screen.getByPlaceholderText("Name"), "Acme Inc");
    await user.type(screen.getByPlaceholderText("Legal name"), "Acme Inc LLC");
    await user.click(screen.getByRole("combobox", { name: "Parent organization" }));
    await user.click(await screen.findByRole("option", { name: "PT Nusantara" }));
    await user.click(screen.getByRole("combobox", { name: "Base currency" }));
    await user.click(await screen.findByRole("option", { name: "US Dollar (USD)" }));
    await user.click(screen.getByRole("combobox", { name: "Country" }));
    await user.click(await screen.findByRole("option", { name: "United States" }));
    await user.type(screen.getByPlaceholderText("Tax ID"), "TAX-1");
    await user.click(screen.getByRole("combobox", { name: "Timezone" }));
    await user.click(await screen.findByRole("option", { name: "UTC" }));
    await user.click(screen.getByRole("combobox", { name: "Tax year starts in" }));
    await user.click(await screen.findByRole("option", { name: "January" }));
    await user.click(screen.getByRole("button", { name: "Save changes" }));

    await waitFor(() => expect(updateCalls).toHaveLength(1));
    expect(updateCalls[0]).toMatchObject({ name: "Acme Inc" });
  });
});
