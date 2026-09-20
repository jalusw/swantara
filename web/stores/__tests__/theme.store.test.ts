import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { mediaQueryMock } from "@/lib/tests";
import { DARK, LIGHT, listenToSystemTheme, useThemeStore } from "../theme.store";

describe("theme store", () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.className = "";
    useThemeStore.setState({ theme: DARK, resolvedTheme: DARK });
  });

  afterEach(() => {
    useThemeStore.setState({ theme: DARK, resolvedTheme: DARK });
  });

  it("starts dark by default", () => {
    expect(useThemeStore.getState().theme).toBe(DARK);
    expect(useThemeStore.getState().resolvedTheme).toBe(DARK);
  });

  it("setTheme applies an explicit theme and persists it", () => {
    useThemeStore.getState().setTheme(LIGHT);

    expect(useThemeStore.getState().theme).toBe(LIGHT);
    expect(useThemeStore.getState().resolvedTheme).toBe(LIGHT);
    expect(document.documentElement.classList.contains(LIGHT)).toBe(true);
    expect(document.documentElement.style.colorScheme).toBe(LIGHT);
    expect(localStorage.getItem("theme")).toBe(LIGHT);
  });

  it("setTheme resolves system to the current media query", () => {
    mediaQueryMock.setMatches(true);

    useThemeStore.getState().setTheme("system");

    expect(useThemeStore.getState().resolvedTheme).toBe(DARK);
    expect(document.documentElement.classList.contains(DARK)).toBe(true);
  });

  it("hydrateTheme restores a stored theme", () => {
    localStorage.setItem("theme", LIGHT);

    useThemeStore.getState().hydrateTheme();

    expect(useThemeStore.getState().theme).toBe(LIGHT);
    expect(document.documentElement.classList.contains(LIGHT)).toBe(true);
  });

  it("hydrateTheme falls back to dark for invalid stored values", () => {
    localStorage.setItem("theme", "neon-pink");

    useThemeStore.getState().hydrateTheme();

    expect(useThemeStore.getState().theme).toBe(DARK);
  });

  it("hydrateTheme tolerates an unavailable localStorage", () => {
    const getItem = vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new Error("denied");
    });

    useThemeStore.getState().hydrateTheme();

    expect(useThemeStore.getState().theme).toBe(DARK);
    getItem.mockRestore();
  });

  it("setTheme tolerates an unavailable localStorage", () => {
    const setItem = vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("denied");
    });

    expect(() => useThemeStore.getState().setTheme(LIGHT)).not.toThrow();
    expect(useThemeStore.getState().theme).toBe(LIGHT);
    setItem.mockRestore();
  });

  it("toggleTheme switches from dark to light", () => {
    useThemeStore.setState({ theme: DARK, resolvedTheme: DARK });

    useThemeStore.getState().toggleTheme();

    expect(useThemeStore.getState().theme).toBe(LIGHT);
    expect(useThemeStore.getState().resolvedTheme).toBe(LIGHT);
    expect(document.documentElement.classList.contains(LIGHT)).toBe(true);
    expect(localStorage.getItem("theme")).toBe(LIGHT);
  });

  it("toggleTheme switches from light to dark", () => {
    useThemeStore.setState({ theme: LIGHT, resolvedTheme: LIGHT });

    useThemeStore.getState().toggleTheme();

    expect(useThemeStore.getState().theme).toBe(DARK);
    expect(useThemeStore.getState().resolvedTheme).toBe(DARK);
    expect(document.documentElement.classList.contains(DARK)).toBe(true);
    expect(localStorage.getItem("theme")).toBe(DARK);
  });

  it("toggleTheme resolves system theme before toggling", () => {
    mediaQueryMock.setMatches(true);
    useThemeStore.setState({ theme: "system", resolvedTheme: DARK });

    useThemeStore.getState().toggleTheme();

    expect(useThemeStore.getState().theme).toBe(LIGHT);
    expect(useThemeStore.getState().resolvedTheme).toBe(LIGHT);
  });

  it("toggleTheme tolerates an unavailable localStorage", () => {
    const setItem = vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("denied");
    });

    expect(() => useThemeStore.getState().toggleTheme()).not.toThrow();
    expect(useThemeStore.getState().theme).toBe(LIGHT);
    setItem.mockRestore();
  });
});

describe("listenToSystemTheme", () => {
  it("re-resolves the theme when the system scheme changes while on system", () => {
    useThemeStore.setState({ theme: "system", resolvedTheme: LIGHT });
    const unsubscribe = listenToSystemTheme();

    mediaQueryMock.setMatches(true);

    expect(useThemeStore.getState().resolvedTheme).toBe(DARK);
    unsubscribe();
  });

  it("leaves a fixed theme untouched when the system scheme changes", () => {
    useThemeStore.setState({ theme: LIGHT, resolvedTheme: LIGHT });
    const unsubscribe = listenToSystemTheme();

    mediaQueryMock.setMatches(true);

    expect(useThemeStore.getState().resolvedTheme).toBe(LIGHT);
    unsubscribe();
  });

  it("stops reacting after being unsubscribed", () => {
    useThemeStore.setState({ theme: "system", resolvedTheme: LIGHT });
    const unsubscribe = listenToSystemTheme();
    unsubscribe();

    mediaQueryMock.setMatches(true);

    expect(useThemeStore.getState().resolvedTheme).toBe(LIGHT);
  });
});
