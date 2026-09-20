import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ProviderContainer } from "../container";

vi.mock("../theme", () => ({
  ThemeProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));
vi.mock("../digital-fingerprint", () => ({
  DigitalFingerPrintProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));
vi.mock("../tanstack-query", () => ({
  TanstackQueryProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));
vi.mock("../session", () => ({
  SessionProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));
vi.mock("@/components/toast", () => ({
  Toaster: () => <div data-testid="toaster" />,
}));

describe("ProviderContainer", () => {
  it("renders children wrapped in every provider", async () => {
    const element = await ProviderContainer({ children: <p>root child</p> });
    render(element);

    expect(screen.getByText("root child")).toBeInTheDocument();
    expect(screen.getByTestId("toaster")).toBeInTheDocument();
  });
});
