import { describe, expect, it } from "vitest";
import { journalTypeTone } from "../journal-utils";

describe("journalTypeTone", () => {
  it("returns correct tone for each journal type", () => {
    expect(journalTypeTone("sale")).toBe("success");
    expect(journalTypeTone("purchase")).toBe("info");
    expect(journalTypeTone("bank")).toBe("warning");
    expect(journalTypeTone("cash")).toBe("warning");
    expect(journalTypeTone("general")).toBe("neutral");
  });
});
