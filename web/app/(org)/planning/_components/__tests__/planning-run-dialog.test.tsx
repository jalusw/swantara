import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PlanningRunDialog } from "../planning-run-dialog";

const STAMP = "2026-01-05T00:00:00Z";

beforeEach(() => {
  server.use(
    http.post("*/api/v1/organizations/:organizationId/planning/runs", () =>
      HttpResponse.json(
        {
          success: true,
          message: "Created.",
          data: {
            run: {
              id: 2,
              organization_id: 1,
              run_date: STAMP,
              horizon_days: 30,
              state: "running",
              created_at: STAMP,
              updated_at: STAMP,
            },
          },
        },
        { status: 201 },
      ),
    ),
  );
});

function renderDialog(onSave: () => void) {
  renderWithProviders(<PlanningRunDialog open onOpenChange={() => {}} orgId="1" onSave={onSave} />);
}

describe("PlanningRunDialog", () => {
  it("renders the run form", async () => {
    renderDialog(() => {});

    expect(await screen.findByRole("heading", { name: "Run Planning" })).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Optional name")).toBeInTheDocument();
  });

  it("starts a run on submit", async () => {
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderDialog(onSave);

    await screen.findByRole("heading", { name: "Run Planning" });
    await user.type(screen.getByPlaceholderText("Optional name"), "January run");
    await user.click(screen.getByRole("button", { name: "Run Planning" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(1));
  });
});
