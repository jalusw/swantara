import { describe, expect, it } from "vitest";
import { SwantaraError, SwantaraUnauthorizedError, SwantaraUnprocessableError } from "../errors";

describe("SwantaraError", () => {
  it("sets name, message, code, and status", () => {
    const err = new SwantaraError("fail", "E001", 400);
    expect(err.name).toBe("SwantaraError");
    expect(err.message).toBe("fail");
    expect(err.code).toBe("E001");
    expect(err.status).toBe(400);
  });

  it("defaults code to null and status to 0", () => {
    const err = new SwantaraError("oops");
    expect(err.code).toBeNull();
    expect(err.status).toBe(0);
  });
});

describe("SwantaraUnauthorizedError", () => {
  it("sets code ERR_UNAUTHORIZED and status 401", () => {
    const err = new SwantaraUnauthorizedError("unauthorized");
    expect(err.code).toBe("ERR_UNAUTHORIZED");
    expect(err.status).toBe(401);
    expect(err).toBeInstanceOf(SwantaraError);
  });
});

describe("SwantaraUnprocessableError", () => {
  it("sets code ERR_UNPROCESSABLE and status 422", () => {
    const err = new SwantaraUnprocessableError("invalid");
    expect(err.code).toBe("ERR_UNPROCESSABLE");
    expect(err.status).toBe(422);
    expect(err.fieldErrors).toBeNull();
  });

  it("accepts field errors", () => {
    const fields = [{ field: "name", code: "required", message: "Required" }];
    const err = new SwantaraUnprocessableError("invalid", fields);
    expect(err.fieldErrors).toEqual(fields);
  });
});
