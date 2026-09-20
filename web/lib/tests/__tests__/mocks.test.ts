import { describe, expect, it } from "vitest";
import {
  cookieJar,
  installBrowserMocks,
  intersectionObserverMock,
  mediaQueryMock,
  navigationMock,
} from "../mocks";

describe("navigationMock", () => {
  it("tracks the pathname", () => {
    navigationMock.setPathname("/orgs");
    expect(navigationMock.usePathname()).toBe("/orgs");
  });

  it("records router actions and resets them", () => {
    navigationMock.push("/login");
    navigationMock.replace("/home");
    expect(navigationMock.push).toHaveBeenCalledWith("/login");
    expect(navigationMock.replace).toHaveBeenCalledWith("/home");
    navigationMock.reset();
    expect(navigationMock.usePathname()).toBe("/");
    expect(navigationMock.push).not.toHaveBeenCalled();
  });
});

describe("mediaQueryMock", () => {
  it("notifies listeners on change", () => {
    installBrowserMocks();
    const query = window.matchMedia("(prefers-reduced-motion: reduce)");
    let notified = 0;
    const listener = () => {
      notified += 1;
    };
    query.addEventListener("change", listener);
    mediaQueryMock.setMatches(true);
    expect(window.matchMedia("(prefers-reduced-motion: reduce)").matches).toBe(true);
    expect(notified).toBe(1);
    query.removeEventListener("change", listener);
    mediaQueryMock.reset();
  });
});

describe("intersectionObserverMock", () => {
  it("fires intersection changes on observed targets", () => {
    installBrowserMocks();
    const seen: boolean[] = [];
    const observer = new IntersectionObserver((entries) => {
      for (const entry of entries) seen.push(entry.isIntersecting);
    });
    observer.observe(document.createElement("div"));
    intersectionObserverMock.triggerAll(true);
    expect(seen).toEqual([true]);
    observer.disconnect();
    intersectionObserverMock.reset();
  });
});

describe("installBrowserMocks", () => {
  it("installs resize observer and pointer helpers", () => {
    installBrowserMocks();
    expect(typeof window.ResizeObserver).toBe("function");
    expect(typeof window.IntersectionObserver).toBe("function");
    expect(typeof Element.prototype.scrollIntoView).toBe("function");
  });

  it("supports localStorage round-trips", () => {
    installBrowserMocks();
    window.localStorage.setItem("mock-key", "mock-value");
    expect(window.localStorage.getItem("mock-key")).toBe("mock-value");
    expect(window.localStorage.length).toBeGreaterThan(0);
    window.localStorage.removeItem("mock-key");
    expect(window.localStorage.getItem("mock-key")).toBeNull();
  });
});

describe("cookieJar", () => {
  it("stores and clears cookies", () => {
    cookieJar.set("probe", "1");
    expect(cookieJar.get("probe")).toBe("1");
    cookieJar.delete("probe");
    expect(cookieJar.has("probe")).toBe(false);
  });
});
