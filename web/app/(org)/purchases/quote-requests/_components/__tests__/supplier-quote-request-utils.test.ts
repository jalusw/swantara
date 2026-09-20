import { describe, expect, it } from "vitest";
import {
  canCancel,
  canCreatePo,
  canEdit,
  canReceiveQuotes,
  canSend,
  quoteRequestStateTone,
} from "../supplier-quote-request-utils";

describe("supplier-quote-request-utils", () => {
  describe("canSend", () => {
    it("returns true for draft", () => {
      expect(canSend("draft")).toBe(true);
    });

    it("returns false for sent", () => {
      expect(canSend("sent")).toBe(false);
    });
  });

  describe("canCancel", () => {
    it("returns true for draft", () => {
      expect(canCancel("draft")).toBe(true);
    });

    it("returns true for sent", () => {
      expect(canCancel("sent")).toBe(true);
    });

    it("returns false for done", () => {
      expect(canCancel("done")).toBe(false);
    });
  });

  describe("canCreatePo", () => {
    it("returns true for done", () => {
      expect(canCreatePo("done")).toBe(true);
    });

    it("returns false for draft", () => {
      expect(canCreatePo("draft")).toBe(false);
    });

    it("returns false for sent", () => {
      expect(canCreatePo("sent")).toBe(false);
    });
  });

  describe("canEdit", () => {
    it("returns true for draft", () => {
      expect(canEdit("draft")).toBe(true);
    });

    it("returns false for sent", () => {
      expect(canEdit("sent")).toBe(false);
    });
  });

  describe("canReceiveQuotes", () => {
    it("returns true for sent", () => {
      expect(canReceiveQuotes("sent")).toBe(true);
    });

    it("returns false for draft", () => {
      expect(canReceiveQuotes("draft")).toBe(false);
    });
  });

  describe("quoteRequestStateTone", () => {
    it("returns neutral for draft", () => {
      expect(quoteRequestStateTone("draft")).toBe("neutral");
    });

    it("returns info for sent", () => {
      expect(quoteRequestStateTone("sent")).toBe("info");
    });

    it("returns success for done", () => {
      expect(quoteRequestStateTone("done")).toBe("success");
    });

    it("returns danger for cancelled", () => {
      expect(quoteRequestStateTone("cancelled")).toBe("danger");
    });
  });
});
