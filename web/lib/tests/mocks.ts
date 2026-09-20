import { vi } from "vitest";

const routerActions = vi.hoisted(() => ({
  push: vi.fn(),
  replace: vi.fn(),
  prefetch: vi.fn(),
  back: vi.fn(),
  refresh: vi.fn(),
}));

const pathnameGetter = vi.hoisted(() => vi.fn(() => "/"));

const cookieStore = vi.hoisted(() => new Map<string, string>());

const mediaQueryState = vi.hoisted(() => ({ matches: false }));
const mediaQueryListeners = vi.hoisted(() => new Set<(event: Event) => void>());

const intersectionState = vi.hoisted(() => ({
  instances: [] as Array<{
    callback: IntersectionObserverCallback;
    targets: Set<Element>;
  }>,
}));

export const navigationMock = {
  push: routerActions.push,
  replace: routerActions.replace,
  prefetch: routerActions.prefetch,
  back: routerActions.back,
  refresh: routerActions.refresh,
  usePathname: pathnameGetter,
  setPathname: (path: string) => {
    pathnameGetter.mockReturnValue(path);
  },
  reset: () => {
    routerActions.push.mockReset();
    routerActions.replace.mockReset();
    routerActions.prefetch.mockReset();
    routerActions.back.mockReset();
    routerActions.refresh.mockReset();
    pathnameGetter.mockReset();
    pathnameGetter.mockReturnValue("/");
  },
};

export const cookieJar = cookieStore;

export const mediaQueryMock = {
  setMatches: (matches: boolean) => {
    mediaQueryState.matches = matches;
    for (const listener of mediaQueryListeners) {
      listener(new Event("change"));
    }
  },
  reset: () => {
    mediaQueryState.matches = false;
    mediaQueryListeners.clear();
  },
};

export const intersectionObserverMock = {
  /** Fire an intersection change on every observed target. */
  triggerAll: (isIntersecting: boolean) => {
    for (const instance of intersectionState.instances) {
      const entries = [...instance.targets].map((target) => ({
        isIntersecting,
        intersectionRatio: isIntersecting ? 1 : 0,
        target,
        boundingClientRect: null,
        intersectionRect: null,
        rootBounds: null,
        time: 0,
      }));
      instance.callback(
        entries as unknown as IntersectionObserverEntry[],
        instance as unknown as IntersectionObserver,
      );
    }
  },
  reset: () => {
    intersectionState.instances.length = 0;
  },
};

export function installBrowserMocks() {
  if (typeof window !== "undefined" && typeof window.localStorage === "undefined") {
    installStorageMock();
  }
  Object.defineProperty(window, "matchMedia", {
    writable: true,
    value: (query: string): MediaQueryList => ({
      matches: mediaQueryState.matches,
      media: query,
      onchange: null,
      addEventListener: (type: string, listener: EventListenerOrEventListenerObject) => {
        if (type === "change") {
          mediaQueryListeners.add(listener as (event: Event) => void);
        }
      },
      removeEventListener: (type: string, listener: EventListenerOrEventListenerObject) => {
        if (type === "change") {
          mediaQueryListeners.delete(listener as (event: Event) => void);
        }
      },
      addListener: () => {},
      removeListener: () => {},
      dispatchEvent: () => true,
    }),
  });

  class ResizeObserverMock {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
  Object.defineProperty(window, "ResizeObserver", {
    writable: true,
    value: ResizeObserverMock,
  });

  class IntersectionObserverMock {
    callback: IntersectionObserverCallback;
    targets = new Set<Element>();

    constructor(callback: IntersectionObserverCallback) {
      this.callback = callback;
      intersectionState.instances.push({ callback, targets: this.targets });
    }

    observe(target: Element) {
      this.targets.add(target);
    }

    unobserve(target: Element) {
      this.targets.delete(target);
    }

    disconnect() {
      this.targets.clear();
    }
  }
  Object.defineProperty(window, "IntersectionObserver", {
    writable: true,
    value: IntersectionObserverMock,
  });

  if (typeof Element.prototype.scrollIntoView !== "function") {
    Element.prototype.scrollIntoView = () => {};
  }

  if (typeof Element.prototype.setPointerCapture !== "function") {
    Element.prototype.setPointerCapture = () => {};
  }
  if (typeof Element.prototype.releasePointerCapture !== "function") {
    Element.prototype.releasePointerCapture = () => {};
  }

  if (typeof Element.prototype.getAnimations !== "function") {
    Element.prototype.getAnimations = () => [];
  }
}

installStorageMockAtModuleScope();

function installStorageMockAtModuleScope() {
  if (typeof globalThis === "undefined" || typeof globalThis.localStorage !== "undefined") {
    return;
  }
  if (typeof window === "undefined") {
    return;
  }
  installStorageMock();
}

function installStorageMock() {
  const data = new Map<string, string>();
  const storage: Storage = {
    get length() {
      return data.size;
    },
    clear: () => data.clear(),
    getItem: (key: string) => data.get(key) ?? null,
    key: (index: number) => [...data.keys()][index] ?? null,
    removeItem: (key: string) => {
      data.delete(key);
    },
    setItem: (key: string, value: string) => {
      data.set(key, String(value));
    },
  };
  Object.defineProperty(window, "localStorage", {
    writable: true,
    value: storage,
  });
  Object.defineProperty(globalThis, "localStorage", {
    writable: true,
    value: storage,
  });
}

vi.mock("next/navigation", () => {
  return {
    useRouter: () => navigationMock,
    usePathname: () => pathnameGetter(),
    useSearchParams: () => new URLSearchParams(),
    redirect: (path: string) => {
      navigationMock.push(path);
    },
  };
});

vi.mock("next/link", async () => {
  const React = await import("react");
  return {
    default: ({
      href,
      children,
      ...props
    }: {
      href: string;
      children: React.ReactNode;
      className?: string;
      "aria-label"?: string;
      onClick?: () => void;
    }) => React.createElement("a", { href, ...props }, children),
  };
});

vi.mock("next/headers", () => ({
  cookies: async () => ({
    get: (name: string) => {
      const value = cookieStore.get(name);
      return value === undefined ? undefined : { name, value };
    },
    getAll: () => [...cookieStore.entries()].map(([name, value]) => ({ name, value })),
    has: (name: string) => cookieStore.has(name),
    set: (name: string, value: string) => {
      cookieStore.set(name, value);
    },
    delete: (name: string) => {
      cookieStore.delete(name);
    },
  }),
  headers: async () => new Headers(),
}));
