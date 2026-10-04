import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { navigationMock, renderWithProviders, server } from "@/lib/tests";
import CashFlowPage from "../page";

vi.mock("next/navigation", async (importOriginal) => ({
  ...((await importOriginal()) as Record<string, unknown>),
  useParams: () => ({ id: "1" }),
  useRouter: () => navigationMock,
  usePathname: () => "/",
}));

const cashFlow = {
  organization_id: 1,
  start: "2026-01-01",
  end: "2026-01-31",
  operating: 1200,
  investing: -300,
  financing: -100,
  net_change: 800,
  opening_cash: 5000,
  closing_cash: 5800,
};

function useCashFlowHandlers(flow: typeof cashFlow | null) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/reports/cash-flow", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { cashFlow: flow } }),
    ),
  );
}

beforeEach(() => {});

describe("CashFlowPage", () => {
  it("renders loading state while fetching", () => {
    server.use(
      http.get(
        "*/api/v1/organizations/:organizationId/reports/cash-flow",
        () => new Promise(() => {}),
      ),
    );
    renderWithProviders(<CashFlowPage />);

    expect(screen.getByText("Memuat arus kas...")).toBeInTheDocument();
  });

  it("renders empty state when no cash flow is returned", async () => {
    useCashFlowHandlers(null);
    renderWithProviders(<CashFlowPage />);

    expect(await screen.findByText("Belum ada data arus kas")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Ekspor CSV" })).toBeDisabled();
  });

  it("renders positive and negative flows with tone classes", async () => {
    useCashFlowHandlers(cashFlow);
    renderWithProviders(<CashFlowPage />);

    expect(await screen.findByText("Operasi")).toBeInTheDocument();
    expect(screen.getByText("Investasi")).toBeInTheDocument();
    expect(screen.getByText("Perubahan bersih")).toBeInTheDocument();
    const operating = screen.getByText("Rp 1.200,00");
    expect(operating.className).toMatch("text-success");
    const investing = screen.getByText("-Rp 300,00");
    expect(investing.className).toMatch("text-destructive");
    expect(screen.getByRole("button", { name: "Ekspor CSV" })).toBeEnabled();
  });

  it("exports csv from the header action", async () => {
    const user = userEvent.setup();
    const anchor = document.createElement("a");
    const clickSpy = vi.spyOn(anchor, "click").mockImplementation(() => {});
    const originalCreate = document.createElement.bind(document);
    const createSpy = vi.spyOn(document, "createElement").mockImplementation(((tag: string) => {
      if (tag === "a") return anchor;
      return originalCreate(tag);
    }) as typeof document.createElement);
    useCashFlowHandlers(cashFlow);
    renderWithProviders(<CashFlowPage />);

    await user.click(await screen.findByRole("button", { name: "Ekspor CSV" }));

    expect(clickSpy).toHaveBeenCalled();
    createSpy.mockRestore();
  });

  it("renders all-negative flows with destructive tone", async () => {
    useCashFlowHandlers({ ...cashFlow, operating: -50, net_change: -450 });
    renderWithProviders(<CashFlowPage />);

    expect(await screen.findByText("Operasi")).toBeInTheDocument();
    expect(screen.getByText("-Rp 50,00").className).toMatch("text-destructive");
  });
});
