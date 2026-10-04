import { act, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { mediaQueryMock } from "@/lib/tests";
import GlobalError from "../global-error";

function renderGlobalError(storedTheme: string | null) {
  if (storedTheme === null) {
    window.localStorage.removeItem("theme");
  } else {
    window.localStorage.setItem("theme", storedTheme);
  }
  return render(<GlobalError error={new Error("boom")} reset={vi.fn()} />);
}

describe("GlobalError", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it("renders the error message and reset action", async () => {
    renderGlobalError(null);
    await act(async () => {});

    expect(screen.getByText("boom")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /coba lagi/i })).toBeInTheDocument();
  });

  it("resolves a stored light theme", async () => {
    renderGlobalError("light");
    await act(async () => {});

    expect(document.documentElement.getAttribute("style")).toContain("color-scheme: light");
  });

  it("resolves a stored dark theme", async () => {
    renderGlobalError("dark");
    await act(async () => {});

    expect(document.documentElement.getAttribute("style")).toContain("color-scheme: dark");
  });

  it("resolves a system theme to light when the media query matches", async () => {
    mediaQueryMock.setMatches(true);
    renderGlobalError("system");
    await act(async () => {});

    expect(document.documentElement.getAttribute("style")).toContain("color-scheme: light");
    mediaQueryMock.reset();
  });

  it("resolves an unknown stored value to dark", async () => {
    renderGlobalError("unknown");
    await act(async () => {});

    expect(document.documentElement.getAttribute("style")).toContain("color-scheme: dark");
  });
});
