import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { SessionProvider, useSession } from "../session";

const mockMe: {
  data?: { user: { id: number; email: string } };
  isLoading: boolean;
} = {
  data: { user: { id: 1, email: "me@example.com" } },
  isLoading: false,
};
const mockOrgs: {
  data?: { organizations: { id: number; name: string }[] };
  isLoading: boolean;
} = {
  data: { organizations: [{ id: 2, name: "Acme" }] },
  isLoading: false,
};

vi.mock("@/lib/hooks/use-me-query", () => ({
  useMeQuery: vi.fn(() => mockMe),
  useMeOrganizationsQuery: vi.fn(() => mockOrgs),
}));

function Probe() {
  const { user, organizations, isAuthenticated, isLoading } = useSession();
  return (
    <div>
      <span data-testid="email">{user?.email ?? "none"}</span>
      <span data-testid="orgs">{organizations?.length ?? 0}</span>
      <span data-testid="auth">{String(isAuthenticated)}</span>
      <span data-testid="loading">{String(isLoading)}</span>
    </div>
  );
}

function renderSession() {
  const client = new QueryClient();
  return render(
    <QueryClientProvider client={client}>
      <SessionProvider>
        <Probe />
      </SessionProvider>
    </QueryClientProvider>,
  );
}

afterEach(() => {
  mockMe.data = { user: { id: 1, email: "me@example.com" } };
  mockMe.isLoading = false;
  mockOrgs.data = { organizations: [{ id: 2, name: "Acme" }] };
  mockOrgs.isLoading = false;
});

describe("SessionProvider", () => {
  it("exposes the authenticated user and organizations", () => {
    renderSession();

    expect(screen.getByTestId("email")).toHaveTextContent("me@example.com");
    expect(screen.getByTestId("orgs")).toHaveTextContent("1");
    expect(screen.getByTestId("auth")).toHaveTextContent("true");
    expect(screen.getByTestId("loading")).toHaveTextContent("false");
  });

  it("reports an unauthenticated loading state when queries have no data", () => {
    mockMe.data = undefined;
    mockMe.isLoading = true;
    mockOrgs.data = undefined;
    mockOrgs.isLoading = true;

    renderSession();

    expect(screen.getByTestId("email")).toHaveTextContent("none");
    expect(screen.getByTestId("auth")).toHaveTextContent("false");
    expect(screen.getByTestId("loading")).toHaveTextContent("true");
  });

  it("throws when used outside the provider", () => {
    const spy = vi.spyOn(console, "error").mockImplementation(() => {});
    expect(() => render(<Probe />)).toThrow(/useSession must be used within/);
    spy.mockRestore();
  });
});
