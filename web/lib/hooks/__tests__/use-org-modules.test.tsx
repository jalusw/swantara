import { waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderHookWithoutOrg, renderHookWithProviders } from "@/lib/tests";
import {
  orgModulesQueryKey,
  useActiveModuleIds,
  useOrgModules,
  useUpdateOrgModule,
} from "../use-org-modules";
import { orgListQueryKey } from "../use-org-query";

describe("orgModulesQueryKey", () => {
  it("builds on the shared org list key", () => {
    expect(orgModulesQueryKey(5)).toEqual(orgListQueryKey("modules", 5));
  });
});

describe("useOrgModules", () => {
  it("fetches modules for the active organization", async () => {
    const { result } = renderHookWithProviders(() => useOrgModules());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.modules.length).toBeGreaterThan(0);
    expect(result.current.data?.modules[0]).toHaveProperty("moduleId");
  });

  it("stays idle without an active organization", () => {
    const { result } = renderHookWithoutOrg(() => useOrgModules());
    expect(result.current.isFetching).toBe(false);
    expect(result.current.isSuccess).toBe(false);
  });
});

describe("useActiveModuleIds", () => {
  it("returns the active module ids as a set", async () => {
    const { result } = renderHookWithProviders(() => useActiveModuleIds());
    await waitFor(() => expect(result.current).not.toBeNull());
    expect(result.current).toBeInstanceOf(Set);
    expect(result.current?.has("crm")).toBe(true);
  });

  it("returns null without module data", () => {
    const { result } = renderHookWithProviders(() => useActiveModuleIds());
    expect(result.current).toBeNull();
  });
});

describe("useUpdateOrgModule", () => {
  it("exposes a mutation for toggling modules", () => {
    const { result } = renderHookWithProviders(() => useUpdateOrgModule());
    expect(typeof result.current.mutate).toBe("function");
    expect(typeof result.current.mutateAsync).toBe("function");
  });
});
