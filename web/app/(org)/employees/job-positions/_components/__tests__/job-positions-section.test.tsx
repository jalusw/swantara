import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { JobPositionsSection } from "../job-positions-section";

const jobPositions = [
  { id: 1, organization_id: 1, name: "Backend Engineer", department_id: 2 },
  { id: 2, organization_id: 1, name: "Designer", department_id: null },
];

const departments = [
  {
    id: 2,
    organization_id: 1,
    name: "Engineering",
    description: null,
    parent_id: null,
    manager_id: null,
    dimension_id: null,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/job-positions", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { jobPositions } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/departments", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { departments } }),
    ),
  );
});

describe("JobPositionsSection", () => {
  it("renders seeded job positions with departments", async () => {
    renderWithProviders(<JobPositionsSection orgId="1" />);

    expect(await screen.findByText("Backend Engineer")).toBeInTheDocument();
    expect(screen.getByText("Designer")).toBeInTheDocument();
    expect(screen.getByText("Engineering")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<JobPositionsSection orgId="1" />);

    await screen.findByText("Backend Engineer");
    await user.click(screen.getByRole("button", { name: "Add position" }));

    expect(await screen.findByRole("heading", { name: "New position" })).toBeInTheDocument();
  });
});
