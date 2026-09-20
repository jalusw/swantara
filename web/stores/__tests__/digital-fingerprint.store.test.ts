import { describe, expect, it } from "vitest";
import { useDigitalFingerprintStore } from "../digital-fingerprint.store";

describe("digital fingerprint store", () => {
  it("starts with no captured values", () => {
    const state = useDigitalFingerprintStore.getState();
    expect(state).toMatchObject({
      deviceName: null,
      fingerprint: null,
      browser: null,
      os: null,
    });
  });

  it("setFingerprint stores the fingerprint", () => {
    useDigitalFingerprintStore.getState().setFingerprint("abc123");

    expect(useDigitalFingerprintStore.getState().fingerprint).toBe("abc123");
  });

  it("setBrowser stores the browser", () => {
    useDigitalFingerprintStore.getState().setBrowser("Chrome");

    expect(useDigitalFingerprintStore.getState().browser).toBe("Chrome");
  });

  it("setOS stores the operating system", () => {
    useDigitalFingerprintStore.getState().setOS("Linux");

    expect(useDigitalFingerprintStore.getState().os).toBe("Linux");
  });

  it("keeps unrelated fields untouched when setting one", () => {
    useDigitalFingerprintStore.getState().setFingerprint("xyz");

    expect(useDigitalFingerprintStore.getState().browser).toBeNull();
    expect(useDigitalFingerprintStore.getState().os).toBeNull();
  });
});
