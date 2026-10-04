import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { RenderHookOptions, RenderOptions } from "@testing-library/react";
import { render, renderHook } from "@testing-library/react";
import { NextIntlClientProvider } from "next-intl";
import type { ReactNode } from "react";
import { useState } from "react";
import { defaultLocale, type Locale } from "@/lib/i18n/config";
import enMessages from "@/messages/en.json";
import idMessages from "@/messages/id.json";
import { OrgActiveProvider } from "@/providers/org-active-provider";

const messagesByLocale: Record<Locale, typeof enMessages> = {
  en: enMessages,
  id: idMessages,
};

export function IntlTestProvider({
  children,
  locale = defaultLocale,
}: {
  children: ReactNode;
  locale?: Locale;
}) {
  return (
    <NextIntlClientProvider
      locale={locale}
      messages={messagesByLocale[locale]}
      timeZone="Asia/Jakarta"
    >
      {children}
    </NextIntlClientProvider>
  );
}

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
      <IntlTestProvider>
        <OrgActiveProvider orgId={1}>{children}</OrgActiveProvider>
      </IntlTestProvider>
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

  return (
    <QueryClientProvider client={queryClient}>
      <IntlTestProvider>{children}</IntlTestProvider>
    </QueryClientProvider>
  );
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
