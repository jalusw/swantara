import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SubscriptionFormDialog } from "../subscription-form-dialog";

const plans = [{ id: 1, name: "Monthly Basic", recurring_interval: "monthly", recurring_count: 1 }];

const contacts = [{ id: 1, organization_id: 1, name: "Acme Corp" }];

const products = [{ id: 9, organization_id: 1, name: "Widget", type: "stockable" }];

function seedOptions() {
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

function renderDialog(onSave: (id: string) => void = () => {}) {
  return renderWithProviders(
    <SubscriptionFormDialog open onOpenChange={() => {}} orgId="1" onSave={onSave} />,
  );
}

beforeEach(() => {
  seedOptions();
});

describe("SubscriptionFormDialog extra", () => {
  it("requires a name before saving", async () => {
    const user = userEvent.setup();
    renderDialog();

    await screen.findByText("Create subscription");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Name is required")).toBeInTheDocument();
  });

  it("keeps a single line removable only when multiple lines exist", async () => {
    const user = userEvent.setup();
    renderDialog();

    await screen.findByText("Create subscription");
    expect(screen.getByRole("button", { name: "×" })).toBeDisabled();

    await user.click(screen.getByRole("button", { name: "Add line" }));
    const removeButtons = screen.getAllByRole("button", { name: "×" });
    expect(removeButtons).toHaveLength(2);
    expect(removeButtons[0]).toBeEnabled();

    await user.click(removeButtons[1]!);
    expect(screen.getAllByLabelText("Qty")).toHaveLength(1);
  });

  it("creates a subscription and notifies on save", async () => {
    const createCalls: unknown[] = [];
    server.use(
      http.post("*/api/v1/organizations/:organizationId/subscriptions", async ({ request }) => {
        createCalls.push(await request.json());
        return HttpResponse.json(
          {
            success: true,
            message: "Created.",
            data: { subscription: { id: 12, name: "Acme Monthly" } },
          },
          { status: 201 },
        );
      }),
    );
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderDialog(onSave);

    await screen.findByText("Create subscription");
    await user.type(screen.getByLabelText("Name"), "Acme Monthly");

    await user.click(screen.getByRole("combobox", { name: "Customer" }));
    await user.click(await screen.findByRole("option", { name: "Acme Corp" }));

    await user.click(screen.getByRole("combobox", { name: "Plan" }));
    await user.click(await screen.findByRole("option", { name: "Monthly Basic" }));

    await user.click(screen.getByRole("combobox", { name: "Item" }));
    await user.click(await screen.findByRole("option", { name: "Widget" }));

    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(createCalls).toHaveLength(1));
    expect(createCalls[0]).toMatchObject({ name: "Acme Monthly" });
    await waitFor(() => expect(onSave).toHaveBeenCalledWith("12"));
  });
});
