import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import AuthLayout from "@/app/(auth)/layout";
import OrganizationLayout from "@/app/(org)/layout";
import RootLayout from "@/app/layout";
import { renderWithProviders } from "@/lib/tests";

vi.mock("@/lib/server/active-org-actions", () => ({
  requireActiveOrgId: async () => 1,
  getActiveOrgId: async () => 1,
  setActiveOrg: async () => {},
  clearActiveOrg: async () => {},
}));

vi.mock("@/lib/server/prefetch", () => ({
  prefetchMeData: async () => ({ queries: [], mutations: [] }),
  prefetchPermissionsData: async () => ({ queries: [], mutations: [] }),
}));

vi.mock("next/font/google", () => ({
  Inter_Tight: () => ({ variable: "--font-sans" }),
}));

vi.mock("@/providers/container", () => ({
  ProviderContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock("@/app/(org)/_components/org-shell", () => ({
  OrgShell: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

// Server components call `getTranslations`/`getLocale`/`getMessages` from
// `next-intl/server`, which throws outside a React Server Components environment
// (vitest resolves the client build). Resolve them from the default test locale (id).
vi.mock("next-intl/server", async () => {
  const loaded = (await import("@/messages/id.json")) as unknown as Record<
    string,
    Record<string, string>
  >;
  const messages: Record<string, Record<string, string>> = (
    loaded as { default?: Record<string, Record<string, string>> }
  ).default ?? loaded;
  return {
    getLocale: async () => "id",
    getMessages: async () => messages,
    getTranslations: async (namespace: string) => {
      const ns = messages[namespace];
      if (!ns) {
        throw new Error(`Missing messages namespace: ${namespace}`);
      }
      return (key: string, values?: Record<string, string | number>) => {
        const message = ns[key];
        if (message === undefined) {
          throw new Error(`Missing message: ${namespace}.${key}`);
        }
        if (!values) return message;
        return message.replace(/\{(\w+)\}/g, (match, name: string) =>
          values[name] !== undefined ? String(values[name]) : match,
        );
      };
    },
  };
});

// Vitest resolves the client build of `next-intl`, where the provider requires an
// explicit `locale` (in a Server Component it is inferred from `getLocale`).
vi.mock("next-intl", async (importOriginal) => {
  const actual = await importOriginal<typeof import("next-intl")>();
  type ProviderProps = Parameters<typeof actual.NextIntlClientProvider>[0];
  return {
    ...actual,
    NextIntlClientProvider: ({ locale = "id", ...rest }: ProviderProps) =>
      actual.NextIntlClientProvider({ locale, ...rest }),
  };
});

describe("layouts", () => {
  it("renders the root layout wrapping children", async () => {
    render(await RootLayout({ children: <p>content</p> }));

    expect(screen.getByText("content")).toBeInTheDocument();
  });

  it("renders the root layout with the skip link", async () => {
    render(await RootLayout({ children: <p>content</p> }));

    expect(screen.getByText("Lewati ke konten utama")).toBeInTheDocument();
  });

  it("renders the auth layout with an accessible main landmark", async () => {
    const element = await AuthLayout({
      children: <p>auth</p>,
    });
    renderWithProviders(element);

    expect(screen.getByText("auth")).toBeInTheDocument();
    expect(screen.getByRole("main")).toBeInTheDocument();
  });

  it("renders the organization layout with the org shell", async () => {
    const element = await OrganizationLayout({
      children: <p>org content</p>,
    });
    renderWithProviders(element);

    expect(screen.getByText("org content")).toBeInTheDocument();
  });
});
