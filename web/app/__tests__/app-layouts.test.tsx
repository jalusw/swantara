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

describe("layouts", () => {
  it("renders the root layout wrapping children", () => {
    render(RootLayout({ children: <p>content</p> }));

    expect(screen.getByText("content")).toBeInTheDocument();
  });

  it("renders the root layout with the skip link", () => {
    render(RootLayout({ children: <p>content</p> }));

    expect(screen.getByText("Skip to main content")).toBeInTheDocument();
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
