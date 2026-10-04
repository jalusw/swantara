import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SubscriptionFormDialog } from "../subscription-form-dialog";

const plans = [{ id: 1, name: "Monthly Basic", recurring_interval: "monthly", recurring_count: 1 }];

const contacts = [{ id: 1, organization_id: 1, name: "Acme Corp" }];

const products = [{ id: 9, organization_id: 1, name: "Widget", type: "stockable" }];

function useSubscriptionOptions() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/subscription-plans", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { plans } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
  );
}

beforeEach(() => {
  useSubscriptionOptions();
});

describe("SubscriptionFormDialog", () => {
  it("renders the create form", async () => {
    renderWithProviders(
      <SubscriptionFormDialog open onOpenChange={() => {}} orgId="1" onSave={() => {}} />,
    );

    expect(await screen.findByText("Langganan baru")).toBeInTheDocument();
    expect(screen.getByLabelText("Nama")).toBeInTheDocument();
  });

  it("appends another subscription line", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <SubscriptionFormDialog open onOpenChange={() => {}} orgId="1" onSave={() => {}} />,
    );

    await screen.findByText("Langganan baru");
    await user.click(screen.getByRole("button", { name: "Tambah baris" }));

    expect(screen.getAllByLabelText("Jml").length).toBe(2);
  });
});
