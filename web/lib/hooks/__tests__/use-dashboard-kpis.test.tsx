import { describe, expect, it } from "vitest";
import { renderHookWithoutOrg, renderHookWithProviders } from "@/lib/tests";
import {
  useArApKpi,
  useCashKpi,
  useFinanceKpi,
  useInventoryKpi,
  useInventoryRatioKpi,
  useManufacturingKpi,
  usePayrollKpi,
  usePipelineKpi,
  useProcurementKpi,
  useProjectKpi,
  useSalesKpi,
  useSubscriptionKpi,
} from "../use-dashboard-kpis";

describe("useDashboardKpis", () => {
  it("returns useFinanceKpi hook with correct shape", () => {
    const { result } = renderHookWithProviders(() => useFinanceKpi());
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });

  it("returns useSalesKpi hook with correct shape", () => {
    const { result } = renderHookWithProviders(() => useSalesKpi());
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });

  it("returns useInventoryKpi hook with correct shape", () => {
    const { result } = renderHookWithProviders(() => useInventoryKpi());
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });

  it("returns the remaining kpi hooks with correct shape", () => {
    const hooks = [
      usePayrollKpi,
      useCashKpi,
      usePipelineKpi,
      useSubscriptionKpi,
      useProjectKpi,
      useProcurementKpi,
      useManufacturingKpi,
      useArApKpi,
      useInventoryRatioKpi,
    ];
    for (const useKpi of hooks) {
      const { result, unmount } = renderHookWithProviders(() => useKpi());
      expect(result.current).toHaveProperty("data");
      expect(result.current).toHaveProperty("isSuccess");
      unmount();
    }
  });

  it("does not fetch without an active organization", () => {
    const { result } = renderHookWithoutOrg(() => useFinanceKpi());
    expect(result.current.isFetching).toBe(false);
    expect(result.current.isSuccess).toBe(false);
  });
});
