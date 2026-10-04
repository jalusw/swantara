import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ServiceContractFormDialog } from "../service-contract-form-dialog";

function seed() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { contacts: [{ id: 2, name: "Acme" }] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/equipments", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { equipments: [{ id: 3, name: "Boiler" }] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/subscriptions", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { subscriptions: [{ id: 4, name: "Paket" }] },
      }),
    ),
    http.post("*/api/v1/organizations/:organizationId/service-contracts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { service_contract: { id: 9 } },
      }),
    ),
  );
}

beforeEach(() => {
  seed();
});

describe("ServiceContractFormDialog branches", () => {
  it("renders the create title", async () => {
    renderWithProviders(
      <ServiceContractFormDialog
        open
        onOpenChange={() => undefined}
        orgId="1"
        onSave={() => undefined}
      />,
    );

    expect(await screen.findByText("Kontrak baru")).toBeInTheDocument();
  });

  it("creates with minimal fields", async () => {
    const user = userEvent.setup();
    let saved: string | null = null;
    renderWithProviders(
      <ServiceContractFormDialog
        open
        onOpenChange={() => undefined}
        orgId="1"
        onSave={(id) => {
          saved = id;
        }}
      />,
    );

    await screen.findByText("Kontrak baru");
    await user.type(screen.getByPlaceholderText("Nama"), "Night support");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(saved).toBe("9"));
  });

  it("creates with all optional fields filled", async () => {
    const user = userEvent.setup();
    let saved: string | null = null;
    renderWithProviders(
      <ServiceContractFormDialog
        open
        onOpenChange={() => undefined}
        orgId="1"
        onSave={(id) => {
          saved = id;
        }}
      />,
    );

    await screen.findByText("Kontrak baru");
    await user.type(screen.getByPlaceholderText("Nama"), "Full cover");
    await user.type(screen.getByPlaceholderText("Cakupan"), "24/7 on-site");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(saved).toBe("9"));
  });

  it("shows an error toast when creation fails", async () => {
    const user = userEvent.setup();
    server.use(
      http.post("*/api/v1/organizations/:organizationId/service-contracts", () =>
        HttpResponse.json({ success: false, message: "Boom." }, { status: 500 }),
      ),
    );
    renderWithProviders(
      <ServiceContractFormDialog
        open
        onOpenChange={() => undefined}
        orgId="1"
        onSave={() => undefined}
      />,
    );

    await screen.findByText("Kontrak baru");
    await user.type(screen.getByPlaceholderText("Nama"), "Failing");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    expect(await screen.findByText("Kontrak baru")).toBeInTheDocument();
  });
});
