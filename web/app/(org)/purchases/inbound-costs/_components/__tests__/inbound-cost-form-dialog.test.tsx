import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { InboundCostFormDialog } from "../inbound-cost-form-dialog";

function useProductOptions() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [] } }),
    ),
  );
}

beforeEach(() => {
  useProductOptions();
});

describe("InboundCostFormDialog", () => {
  it("renders the create form", async () => {
    renderWithProviders(
      <InboundCostFormDialog open onOpenChange={() => {}} orgId="1" onSave={() => {}} />,
    );

    expect(await screen.findByText("Create landed cost")).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toBeInTheDocument();
  });

  it("appends another cost line", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <InboundCostFormDialog open onOpenChange={() => {}} orgId="1" onSave={() => {}} />,
    );

    await screen.findByText("Create landed cost");
    await user.click(screen.getByRole("button", { name: "Add line" }));

    expect(screen.getAllByLabelText("Amount").length).toBe(2);
  });
});
