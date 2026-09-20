import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { DeferralFormDialog } from "../deferral-form-dialog";

beforeEach(() => {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/deferrals", () =>
      HttpResponse.json(
        {
          success: true,
          message: "Created.",
          data: { schedule: { id: 3 } },
        },
        { status: 201 },
      ),
    ),
  );
});

function renderDialog(onSave: (id: string) => void) {
  renderWithProviders(
    <DeferralFormDialog open onOpenChange={() => {}} orgId="1" onSave={onSave} />,
  );
}

describe("DeferralFormDialog", () => {
  it("renders the create form", async () => {
    renderDialog(() => {});

    expect(await screen.findByRole("heading", { name: "Create deferral" })).toBeInTheDocument();
    expect(screen.getByLabelText("Source type")).toBeInTheDocument();
  });

  it("creates a deferral on submit", async () => {
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderDialog(onSave);

    await screen.findByRole("heading", { name: "Create deferral" });

    await user.click(screen.getByRole("combobox", { name: "Type" }));
    await user.click((await screen.findAllByRole("option"))[0]!);
    await user.click(screen.getByRole("combobox", { name: "Method" }));
    await user.click((await screen.findAllByRole("option"))[0]!);

    await user.type(screen.getByLabelText("Source type"), "invoice");
    const sourceId = screen.getByLabelText("Source ID");
    await user.clear(sourceId);
    await user.type(sourceId, "7");
    const totalAmount = screen.getByLabelText("Total amount");
    await user.clear(totalAmount);
    await user.type(totalAmount, "1200");
    await user.type(screen.getByLabelText("Start date"), "2026-02-01");

    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("3"));
  });
});
