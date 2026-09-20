import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import { OrgActiveProvider } from "@/providers/org-active-provider";
import { useActiveOrg, useOrganizationId } from "../use-org-context";

function renderOrgHook<Result>(hook: () => Result, orgId: number | null) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: Number.POSITIVE_INFINITY } },
  });
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {orgId == null ? children : <OrgActiveProvider orgId={orgId}>{children}</OrgActiveProvider>}
      </QueryClientProvider>
    );
  }
  return renderHook(hook, { wrapper: Wrapper });
}

describe("org context hooks", () => {
  it("exposes the active organization from the provider", () => {
    const { result } = renderOrgHook(
      () => ({
        activeOrg: useActiveOrg(),
        organizationId: useOrganizationId(),
      }),
      11,
    );

    expect(result.current).toEqual({
      activeOrg: { id: 11 },
      organizationId: 11,
    });
  });

  it("returns null when no organization is active", () => {
    const { result } = renderOrgHook(
      () => ({
        activeOrg: useActiveOrg(),
        organizationId: useOrganizationId(),
      }),
      null,
    );

    expect(result.current).toEqual({ activeOrg: null, organizationId: null });
  });
});
