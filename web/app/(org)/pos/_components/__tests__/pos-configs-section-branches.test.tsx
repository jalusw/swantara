import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PosConfigsSection } from "../pos-configs-section";

const CONFIGS = [
  {
    id: 1,
    organization_id: 1,
    name: "",
    warehouse_id: 9,
    journal_id: null,
    price_book_id: null,
  },
  {
    id: 2,
    organization_id: 1,
    name: "Outlet Kiosk",
    warehouse_id: null,
    journal_id: 7,
    price_book_id: 8,
  },
];

function useHandlers(options?: { failConfigs?: boolean }) {
  const created: unknown[] = [];
  server.use(
    http.get("*/api/v1/organizations/:organizationId/pos/configs", () => {
      if (options?.failConfigs) {
        return HttpResponse.json({ success: false, message: "Configs down." }, { status: 500 });
      }
      return HttpResponse.json({ success: true, message: "OK.", data: { configs: CONFIGS } });
    }),
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { warehouses: [{ id: 1, organization_id: 1, name: "Central Warehouse" }] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { journals: [{ id: 1, organization_id: 1, name: "Cash Journal" }] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/price_books", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { price_books: [{ id: 1, organization_id: 1, name: "Retail PriceBook" }] },
      }),
    ),
    http.post("*/api/v1/organizations/:organizationId/pos/configs", async ({ request }) => {
      created.push(await request.json());
      return HttpResponse.json(
        { success: true, message: "Dibuat.", data: { config: { id: 3 } } },
        { status: 201 },
      );
    }),
  );
  return { created };
}

beforeEach(() => {});

describe("PosConfigsSection branches", () => {
  it("falls back to generated names and unknown references", async () => {
    useHandlers();
    renderWithProviders(<PosConfigsSection orgId="1" />);
    expect(await screen.findByText("POS-1")).toBeInTheDocument();
    expect(screen.getByText("#9")).toBeInTheDocument();
    expect(screen.getByText("#7")).toBeInTheDocument();
    expect(screen.getByText("#8")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("shows the error state with retry when configs fail", async () => {
    const user = userEvent.setup();
    useHandlers({ failConfigs: true });
    renderWithProviders(<PosConfigsSection orgId="1" />);
    expect(await screen.findByText("Configs down.")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Coba lagi" }));
    expect(await screen.findByText("Configs down.")).toBeInTheDocument();
  });

  it("cancels the create dialog", async () => {
    const user = userEvent.setup();
    useHandlers();
    renderWithProviders(<PosConfigsSection orgId="1" />);
    await screen.findByText("Outlet Kiosk");
    await user.click(screen.getByRole("button", { name: "Konfigurasi baru" }));
    await screen.findByRole("heading", { name: "Konfigurasi baru" });
    await user.click(screen.getByRole("button", { name: "Batal" }));
    await waitFor(() =>
      expect(screen.queryByRole("heading", { name: "Konfigurasi baru" })).not.toBeInTheDocument(),
    );
  });

  it("keeps save disabled without a name", async () => {
    const user = userEvent.setup();
    const { created } = useHandlers();
    renderWithProviders(<PosConfigsSection orgId="1" />);
    await screen.findByText("Outlet Kiosk");
    await user.click(screen.getByRole("button", { name: "Konfigurasi baru" }));
    await screen.findByRole("heading", { name: "Konfigurasi baru" });
    expect(screen.getByRole("button", { name: "Simpan" })).toBeDisabled();
    expect(created.length).toBe(0);
  });

  it("creates a config with null optionals", async () => {
    const user = userEvent.setup();
    const { created } = useHandlers();
    renderWithProviders(<PosConfigsSection orgId="1" />);
    await screen.findByText("Outlet Kiosk");
    await user.click(screen.getByRole("button", { name: "Konfigurasi baru" }));
    await user.type(screen.getByPlaceholderText("mis. Toko Utama"), "Night Market");
    await user.click(screen.getByRole("button", { name: "Simpan" }));
    await waitFor(() => expect(created.length).toBe(1));
    const body = created[0] as Record<string, unknown>;
    expect(body.name).toBe("Night Market");
    expect(body.warehouse_id).toBeNull();
    expect(body.journal_id).toBeNull();
    expect(body.price_book_id).toBeNull();
  });
});
