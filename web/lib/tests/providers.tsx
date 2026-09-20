import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { RenderHookOptions, RenderOptions } from "@testing-library/react";
import { render, renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { useState } from "react";
import { OrgActiveProvider } from "@/providers/org-active-provider";

export function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: Number.POSITIVE_INFINITY,
      },
      mutations: {
        retry: false,
      },
    },
  });
}

export function TestProviders({ children }: { children: ReactNode }) {
  const [queryClient] = useState(createTestQueryClient);

  return (
    <QueryClientProvider client={queryClient}>
      <OrgActiveProvider orgId={1}>{children}</OrgActiveProvider>
    </QueryClientProvider>
  );
}

export function renderWithProviders(
  ui: React.ReactElement,
  options?: Omit<RenderOptions, "wrapper">,
) {
  return render(ui, { wrapper: TestProviders, ...options });
}

export function TestProvidersWithoutOrg({ children }: { children: ReactNode }) {
  const [queryClient] = useState(createTestQueryClient);

  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

export function renderHookWithoutOrg<Result, Props>(
  hook: (initialProps: Props) => Result,
  options?: Omit<RenderHookOptions<Props>, "wrapper">,
) {
  return renderHook(hook, { wrapper: TestProvidersWithoutOrg, ...options });
}

export function renderHookWithProviders<Result, Props>(
  hook: (initialProps: Props) => Result,
  options?: Omit<RenderHookOptions<Props>, "wrapper">,
) {
  return renderHook(hook, { wrapper: TestProviders, ...options });
}
