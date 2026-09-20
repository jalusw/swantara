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

const flow = {
  organization_id: 1,
  start: "2026-01-01",
  end: "2026-01-31",
  operating: 500,
  investing: 200,
  financing: 100,
  net_change: 800,
  opening_cash: 5000,
  closing_cash: 5800,
};

function useFlow(payload: typeof flow | null) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/reports/cash-flow", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { cashFlow: payload } }),
    ),
  );
}

beforeEach(() => {});

describe("CashFlowPage branches2", () => {
  it("renders all-positive tones with success classes", async () => {
    useFlow(flow);
    renderWithProviders(<CashFlowPage />);

    expect(await screen.findByText("Operating")).toBeInTheDocument();
    expect(screen.getByText("IDR 500.00").className).toMatch("text-success");
    expect(screen.getByText("IDR 200.00").className).toMatch("text-success");
    expect(screen.getByText("IDR 100.00").className).toMatch("text-success");
    expect(screen.getByText("IDR 800.00").className).toMatch("text-success");
    expect(screen.getByText("IDR 5,000.00")).toBeInTheDocument();
    expect(screen.getByText("IDR 5,800.00")).toBeInTheDocument();
  });

  it("renders all-negative tones with destructive classes", async () => {
    useFlow({ ...flow, operating: -10, investing: -20, financing: -30, net_change: -60 });
    renderWithProviders(<CashFlowPage />);

    expect(await screen.findByText("Operating")).toBeInTheDocument();
    expect(screen.getByText("-IDR 10.00").className).toMatch("text-destructive");
    expect(screen.getByText("-IDR 20.00").className).toMatch("text-destructive");
    expect(screen.getByText("-IDR 30.00").className).toMatch("text-destructive");
    expect(screen.getByText("-IDR 60.00").className).toMatch("text-destructive");
  });

  it("navigates back to the reports overview", async () => {
    const user = userEvent.setup();
    useFlow(flow);
    renderWithProviders(<CashFlowPage />);

    await screen.findByText("Operating");
    await user.click(screen.getByRole("button", { name: "Back to reports" }));

    expect(navigationMock.push).toHaveBeenCalledWith("/reports");
  });
});
