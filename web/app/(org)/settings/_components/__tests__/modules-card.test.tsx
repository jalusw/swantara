import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ModulesCard } from "../modules-card";

function renderCard() {
  renderWithProviders(<ModulesCard />);
}

function permissionsHandler(codes: string[]) {
  return http.get("*/api/v1/me/organizations/:organizationId/permissions", () =>
    HttpResponse.json({
      success: true,
      message: "OK.",
      data: {
        permissions: codes.map((code, index) => {
          const [resource, action] = code.split(".");
          return {
            id: index + 1,
            name: `${action} ${resource}`,
            code,
            description: null,
            resource,
            action,
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z",
          };
        }),
      },
    }),
  );
}

describe("ModulesCard", () => {
  it("sends the toggle state when a module switch changes", async () => {
    const user = userEvent.setup();
    let body: unknown;
    server.use(
      permissionsHandler(["organization.view", "organization.update"]),
      http.put("*/api/v1/organizations/:organizationId/modules", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json({
          success: true,
          message: "Module updated successfully.",
          data: { module: { module_id: "hr", active: false } },
        });
      }),
    );
    renderCard();

    const hrSwitch = await screen.findByRole("switch", { name: "HR" });
    expect(hrSwitch).toHaveAttribute("data-checked");
    await user.click(hrSwitch);

    await waitFor(() => expect(body).toMatchObject({ module_id: "hr", active: false }));
  });

  it("disables switches without the organization update permission", async () => {
    server.use(permissionsHandler(["organization.view"]));
    renderCard();

    const hrSwitch = await screen.findByRole("switch", { name: "HR" });
    expect(hrSwitch).toHaveAttribute("aria-disabled", "true");
  });
});
