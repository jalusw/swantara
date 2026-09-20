import { describe, expect, it } from "vitest";
import {
  SwantaraError,
  SwantaraService,
  SwantaraUnauthorizedError,
  SwantaraUnprocessableError,
  toSwantaraError,
} from "../index";

describe("swantara barrel exports", () => {
  it("exports SwantaraService", () => {
    expect(SwantaraService).toBeDefined();
  });

  it("exports error classes", () => {
    expect(SwantaraError).toBeDefined();
    expect(SwantaraUnauthorizedError).toBeDefined();
    expect(SwantaraUnprocessableError).toBeDefined();
  });

  it("exports toSwantaraError mapper", () => {
    expect(typeof toSwantaraError).toBe("function");
  });
});
