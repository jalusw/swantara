import { describe, expect, it } from "vitest";
import { renderHookWithoutOrg, renderHookWithProviders } from "@/lib/tests";
import {
  useAgingReport,
  useBalanceSheet,
  useProfitAndLoss,
  useTrialBalance,
} from "../use-report-queries";

describe("useReportQueries", () => {
  it("returns useTrialBalance hook with correct shape", () => {
    const { result } = renderHookWithProviders(() => useTrialBalance());
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });

  it("returns useAgingReport hook with correct shape", () => {
    const { result } = renderHookWithProviders(() => useAgingReport());
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });

  it("returns useProfitAndLoss hook with correct shape", () => {
    const { result } = renderHookWithProviders(() => useProfitAndLoss());
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });

  it("does not fetch without an active organization", () => {
    const { result } = renderHookWithoutOrg(() => useBalanceSheet());
    expect(result.current.isFetching).toBe(false);
    expect(result.current.isSuccess).toBe(false);
  });
});
