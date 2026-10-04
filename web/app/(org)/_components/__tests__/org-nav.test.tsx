import { screen, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";
import { SidebarProvider } from "@/components/sidebar";

import { renderWithProviders, server } from "@/lib/tests";
import { OrgActiveProvider } from "@/providers/org-active-provider";
import { OrgNav } from "../org-nav";

describe("OrgNav", () => {
  it("renders only the items the actor has permission for", async () => {
    server.use(
      http.get("*/api/v1/me/organizations/:organizationId/permissions", () =>
        HttpResponse.json({
          success: true,
          message: "OK.",
          data: {
            permissions: [
              {
                id: 1,
                name: "View contact",
                code: "contact.view",
                description: null,
                resource: "contact",
                action: "view",
                created_at: "2026-01-01T00:00:00Z",
                updated_at: "2026-01-01T00:00:00Z",
              },
            ],
          },
        }),
      ),
    );

    renderWithProviders(
      <OrgActiveProvider orgId={1}>
        <SidebarProvider>
          <OrgNav />
        </SidebarProvider>
      </OrgActiveProvider>,
    );

    expect(await screen.findByRole("link", { name: "Dasbor" })).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByRole("button", { name: "CRM" })).toBeNull());
    expect(screen.queryByRole("link", { name: "Penjualan" })).toBeNull();
    await waitFor(() => expect(screen.queryByRole("link", { name: "Karyawan" })).toBeNull());
    expect(screen.queryByRole("link", { name: "Anggota" })).toBeNull();
    expect(screen.queryByRole("link", { name: "Billing" })).toBeNull();
  });
});
