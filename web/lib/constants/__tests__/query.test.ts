import { describe, expect, it } from "vitest";
import { lookupStaleTime, serverQueryStaleTime } from "@/lib/constants/query";

describe("query constants", () => {
  it("lookupStaleTime is 30 minutes in milliseconds", () => {
    expect(lookupStaleTime).toBe(1000 * 60 * 30);
  });

  it("serverQueryStaleTime is 2 minutes in milliseconds", () => {
    expect(serverQueryStaleTime).toBe(1000 * 60 * 2);
  });
});
