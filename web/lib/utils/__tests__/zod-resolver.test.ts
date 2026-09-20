import { describe, expect, it } from "vitest";
import { z } from "zod";
import { zodResolver } from "../zod-resolver";

const schema = z.object({ name: z.string().min(1) });

describe("zodResolver", () => {
  it("passes valid values through", async () => {
    const resolver = zodResolver(schema);
    const result = await resolver({ name: "ok" }, undefined, {
      fields: {},
      shouldUseNativeValidation: false,
    });
    expect(result.errors).toEqual({});
    expect(result.values).toMatchObject({ name: "ok" });
  });

  it("reports validation errors for invalid values", async () => {
    const resolver = zodResolver(schema);
    const result = await resolver({ name: "" }, undefined, {
      fields: {},
      shouldUseNativeValidation: false,
    });
    expect(result.values).toEqual({});
    expect(result.errors.name).toBeDefined();
  });
});
