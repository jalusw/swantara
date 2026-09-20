import { describe, expect, it } from "vitest";
import { paymentStateTone } from "../payment-utils";

describe("paymentStateTone", () => {
  it("returns correct tone for each payment state", () => {
    expect(paymentStateTone("draft")).toBe("neutral");
    expect(paymentStateTone("posted")).toBe("success");
    expect(paymentStateTone("reconciled")).toBe("info");
    expect(paymentStateTone("cancelled")).toBe("danger");
  });
});
