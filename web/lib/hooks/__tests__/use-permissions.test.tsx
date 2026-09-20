import { waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderHookWithoutOrg, renderHookWithProviders } from "@/lib/tests";
import { usePermissions } from "../use-permissions";

describe("usePermissions", () => {
  it("resolves the granted permission codes for the active organization", async () => {
    const { result } = renderHookWithProviders(() => usePermissions());

    await waitFor(() => {
      expect(result.current.permissions).toContain("contact.view");
    });

    expect(result.current.has("contact.view")).toBe(true);
    expect(result.current.has("missing.view")).toBe(false);
    expect(result.current.has(["contact.view", "employee.view"])).toBe(true);
    expect(result.current.has(["contact.view", "missing.view"])).toBe(false);
    expect(result.current.has(["contact.view", "missing.view"], false)).toBe(true);
  });

  it("reports no permissions without an active organization", () => {
    const { result } = renderHookWithoutOrg(() => usePermissions());

    expect(result.current.permissions).toEqual([]);
    expect(result.current.has("contact.view")).toBe(false);
  });
});
