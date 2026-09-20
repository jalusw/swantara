import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { AttendanceSection } from "../attendance-section";

function useLocalAttendance() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/attendances", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          attendances: [
            {
              id: 1,
              employee_id: 1,
              check_in: "2026-03-01T08:00:00Z",
              check_out: null,
              worked_hours: 0,
            },
            {
              id: 2,
              employee_id: 2,
              check_in: "2026-03-01T08:00:00Z",
              check_out: "2026-03-01T16:00:00Z",
              worked_hours: 8,
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
  );
}

beforeEach(() => {
  useLocalAttendance();
});

describe("AttendanceSection", () => {
  it("renders seeded attendance records", async () => {
    renderWithProviders(<AttendanceSection orgId="1" />);

    expect(await screen.findByText("EMP-0001")).toBeInTheDocument();
    expect(screen.getByText("EMP-0002")).toBeInTheDocument();
  });

  it("opens the check-in dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<AttendanceSection orgId="1" />);

    await screen.findByText("EMP-0001");
    await user.click(screen.getByRole("button", { name: "Check in" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });
});
