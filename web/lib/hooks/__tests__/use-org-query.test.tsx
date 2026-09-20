import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import { OrgActiveProvider } from "@/providers/org-active-provider";
import { useOrgListQuery, useOrgQuery } from "../use-org-query";

function renderOrgHook<Result>(hook: () => Result, org: { current: number | null }) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: Number.POSITIVE_INFINITY } },
  });
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {org.current == null ? (
          children
        ) : (
          <OrgActiveProvider orgId={org.current}>{children}</OrgActiveProvider>
        )}
      </QueryClientProvider>
    );
  }
  return renderHook(hook, { wrapper: Wrapper });
}

describe("org-scoped query hooks", () => {
  it("runs the list fetcher with the active organization and params", async () => {
    const fetcher = vi.fn(async (organizationId: number, params: Record<string, unknown>) => ({
      organizationId,
      params,
    }));

    const { result } = renderOrgHook(() => useOrgListQuery("members", fetcher, { page: 1 }), {
      current: 42,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(fetcher).toHaveBeenCalledWith(42, { page: 1 });
    expect(result.current.data).toEqual({
      organizationId: 42,
      params: { page: 1 },
    });
  });

  it("runs the detail fetcher scoped to the active organization", async () => {
    const fetcher = vi.fn(async (organizationId: number) => ({
      organizationId,
    }));

    const { result } = renderOrgHook(() => useOrgQuery("members", 5, fetcher), { current: 7 });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(fetcher).toHaveBeenCalledWith(7);
  });

  it("does not run the fetcher without an active organization", () => {
    const fetcher = vi.fn(async () => [{ id: 1 }]);

    const { result } = renderOrgHook(() => useOrgListQuery("members", fetcher), { current: null });

    expect(result.current.isFetching).toBe(false);
    expect(fetcher).not.toHaveBeenCalled();
  });

  it("refetches under the new organization when the active org changes", async () => {
    const fetcher = vi.fn(async (organizationId: number) => ({
      organizationId,
    }));
    const org = { current: 1 as number | null };

    const { result, rerender } = renderOrgHook(() => useOrgListQuery("contacts", fetcher), org);

    await waitFor(() => expect(fetcher).toHaveBeenCalledWith(1, {}));

    org.current = 2;
    rerender();

    await waitFor(() => expect(fetcher).toHaveBeenCalledWith(2, {}));
    expect(result.current.data).toEqual({ organizationId: 2 });
  });
});
