import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { DepartmentsSection } from "../departments-section";

function useLocalDepartments() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/departments", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          departments: [
            {
              id: 1,
              organization_id: 1,
              name: "Engineering",
              description: "Item engineering",
              parent_id: null,
              manager_id: null,
              dimension_id: null,
            },
            {
              id: 2,
              organization_id: 1,
              name: "Marketing",
              description: null,
              parent_id: null,
              manager_id: null,
              dimension_id: null,
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalDepartments();
});

describe("DepartmentsSection", () => {
  it("renders seeded departments", async () => {
    renderWithProviders(<DepartmentsSection orgId="1" />);

    expect(await screen.findByText("Engineering")).toBeInTheDocument();
    expect(screen.getByText("Marketing")).toBeInTheDocument();
  });

  it("filters departments by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DepartmentsSection orgId="1" />);

    await screen.findByText("Engineering");
    await user.type(screen.getByPlaceholderText("Cari departemen…"), "Marketing");

    expect((await screen.findAllByText("Marketing")).length).toBeGreaterThan(0);
    await waitFor(() => expect(screen.queryByText("Engineering")).not.toBeInTheDocument());
  });
});
