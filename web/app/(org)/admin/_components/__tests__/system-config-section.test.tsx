import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SystemConfigSection } from "../system-config-section";

const systemConfigs = [
  {
    id: 1,
    organization_id: 1,
    key: "app.name",
    value: "Swantara",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: 2,
    organization_id: 1,
    key: "smtp.password",
    value: "super-secret",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

function useConfigHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/system-configs", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { system_configs: systemConfigs } }),
    ),
  );
}

beforeEach(() => {
  useConfigHandlers();
});

describe("SystemConfigSection", () => {
  it("renders config keys and masks sensitive values", async () => {
    renderWithProviders(<SystemConfigSection orgId="1" />);

    expect(await screen.findByText("app.name")).toBeInTheDocument();
    expect(screen.getByText("smtp.password")).toBeInTheDocument();
    expect(screen.getByText("Swantara")).toBeInTheDocument();
    expect(screen.getByText("••••••••")).toBeInTheDocument();
  });

  it("opens the editor dialog from a row edit button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SystemConfigSection orgId="1" />);

    await screen.findByText("app.name");
    const editButton = screen.getAllByRole("button", { name: "Edit" })[0];
    if (editButton === undefined) {
      throw new Error("Expected an Edit button");
    }
    await user.click(editButton);

    expect(await screen.findByText("Edit configuration")).toBeInTheDocument();
  });
});
