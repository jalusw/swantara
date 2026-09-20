import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { LeaveTypesSection } from "../leave-types-section";

const leaveTypes = [
  { id: 1, organization_id: 1, name: "Annual Leave", paid: true, allocation_days: 12 },
  { id: 2, organization_id: 1, name: "Sick Leave", paid: true, allocation_days: 6 },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/leave-types", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { leaveTypes } }),
    ),
  );
});

describe("LeaveTypesSection", () => {
  it("renders seeded leave types", async () => {
    renderWithProviders(<LeaveTypesSection orgId="1" />);

    expect(await screen.findByText("Annual Leave")).toBeInTheDocument();
    expect(screen.getByText("Sick Leave")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<LeaveTypesSection orgId="1" />);

    await screen.findByText("Annual Leave");
    await user.click(screen.getByRole("button", { name: "Add leave type" }));

    expect(await screen.findByRole("heading", { name: "New leave type" })).toBeInTheDocument();
  });
});
