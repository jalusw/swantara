import "./lib/tests/polyfills";
import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterAll, afterEach, beforeAll } from "vitest";
import { resetReferenceFixtures } from "./lib/tests/handlers";
import {
  cookieJar,
  installBrowserMocks,
  intersectionObserverMock,
  mediaQueryMock,
  navigationMock,
} from "./lib/tests/mocks";
import { server } from "./lib/tests/server";
import { resetStores } from "./lib/tests/stores";

// Polyfill DOMMatrix for pdfjs-dist (jsdom does not provide it)
if (typeof globalThis.DOMMatrix === "undefined") {
  class DOMMatrixPolyfill {
    a = 1;
    b = 0;
    c = 0;
    d = 1;
    e = 0;
    f = 0;
    m11 = 1;
    m12 = 0;
    m13 = 0;
    m14 = 0;
    m21 = 0;
    m22 = 1;
    m23 = 0;
    m24 = 0;
    m31 = 0;
    m32 = 0;
    m33 = 1;
    m34 = 0;
    m41 = 0;
    m42 = 0;
    m43 = 0;
    m44 = 1;
    constructor() {}
    static fromMatrix() {
      return new DOMMatrixPolyfill();
    }
    multiplySelf() {
      return this;
    }
    translateSelf() {
      return this;
    }
    scaleSelf() {
      return this;
    }
    invertSelf() {
      return this;
    }
    rotateSelf() {
      return this;
    }
  }
  globalThis.DOMMatrix = DOMMatrixPolyfill as unknown as typeof DOMMatrix;
  globalThis.DOMMatrixReadOnly = DOMMatrixPolyfill as unknown as typeof DOMMatrixReadOnly;
  if (typeof window !== "undefined") {
    window.DOMMatrix = DOMMatrixPolyfill as unknown as typeof DOMMatrix;
  }
}

beforeAll(() => {
  if (typeof window !== "undefined") {
    installBrowserMocks();
  }
  server.listen({ onUnhandledRequest: "error" });
});

afterEach(() => {
  cleanup();
  server.resetHandlers();
  navigationMock.reset();
  cookieJar.clear();
  mediaQueryMock.reset();
  intersectionObserverMock.reset();
  resetStores();
  resetReferenceFixtures();
});

afterAll(() => {
  server.close();
});
