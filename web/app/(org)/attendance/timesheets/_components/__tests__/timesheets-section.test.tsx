import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { TimesheetsSection } from "../timesheets-section";

function useLocalTimesheets() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/timesheets", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          timesheets: [
            {
              id: 1,
              employee_id: 1,
              date: "2026-02-10",
              project_id: null,
              task_id: null,
              dimension_id: null,
              hours: 8,
              description: "Backend development",
            },
            {
              id: 2,
              employee_id: 2,
              date: "2026-02-11",
              project_id: null,
              task_id: null,
              dimension_id: null,
              hours: 6,
              description: "Design review",
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/employees", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          employees: [
            {
              id: 1,
              organization_id: 1,
              contact_id: 10,
              employee_number: "EMP-0001",
              hire_date: "2024-01-01",
              active: true,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
            {
              id: 2,
              organization_id: 1,
              contact_id: 11,
              employee_number: "EMP-0002",
              hire_date: "2024-01-01",
              active: true,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          contacts: [
            { id: 10, organization_id: 1, name: "Alex Rivera", display_name: "Alex Rivera" },
            { id: 11, organization_id: 1, name: "June Park", display_name: "June Park" },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalTimesheets();
});

describe("TimesheetsSection", () => {
  it("renders seeded timesheets", async () => {
    renderWithProviders(<TimesheetsSection orgId="1" />);

    expect(await screen.findByText("Alex Rivera")).toBeInTheDocument();
    expect(screen.getByText("Backend development")).toBeInTheDocument();
  });

  it("filters timesheets by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TimesheetsSection orgId="1" />);

    await screen.findByText("Alex Rivera");
    await user.type(screen.getByPlaceholderText("Search timesheets…"), "Design");

    expect(await screen.findByText("Design review")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TimesheetsSection orgId="1" />);

    await screen.findByText("Alex Rivera");
    await user.click(screen.getByRole("button", { name: "Add entry" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });
});
