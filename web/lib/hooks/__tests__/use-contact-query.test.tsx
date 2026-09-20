import { describe, expect, it } from "vitest";
import { renderHookWithProviders } from "@/lib/tests";
import {
  useContactAddressesQuery,
  useContactBankAccountsQuery,
  useContactQuery,
} from "../use-contact-query";

describe("useContactQuery", () => {
  it("returns a query result with correct shape", () => {
    const { result } = renderHookWithProviders(() => useContactQuery("1", "10"));
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });

  it("is disabled when orgId is empty", () => {
    const { result } = renderHookWithProviders(() => useContactQuery("", "10"));
    expect(result.current.isFetching).toBe(false);
    expect(result.current.isSuccess).toBe(false);
  });
});

describe("useContactAddressesQuery", () => {
  it("returns a query result with correct shape", () => {
    const { result } = renderHookWithProviders(() => useContactAddressesQuery("1", "10"));
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });
});

describe("useContactBankAccountsQuery", () => {
  it("returns a query result with correct shape", () => {
    const { result } = renderHookWithProviders(() => useContactBankAccountsQuery("1", "10"));
    expect(result.current).toHaveProperty("data");
    expect(result.current).toHaveProperty("isSuccess");
  });
});
