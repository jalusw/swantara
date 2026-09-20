import { AxiosError, type AxiosResponse, type InternalAxiosRequestConfig } from "axios";
import { describe, expect, it } from "vitest";
import {
  SwantaraError,
  SwantaraUnauthorizedError,
  SwantaraUnprocessableError,
  toSwantaraError,
} from "@/lib/services/swantara";

function axiosErrorWith(body: unknown, status = 422): AxiosError {
  return new AxiosError(
    `Request failed with status code ${status}`,
    undefined,
    {} as InternalAxiosRequestConfig,
    undefined,
    { data: body, status } as AxiosResponse,
  );
}

describe("toSwantaraError", () => {
  it("maps non-axios errors to a network SwantaraError", () => {
    const error = toSwantaraError(new Error("boom"));
    expect(error).toBeInstanceOf(SwantaraError);
    expect(error.message).toBe("Network error.");
    expect(error.code).toBeNull();
    expect(error.status).toBe(0);
  });

  it("maps ERR_BAD_REQUEST to SwantaraError", () => {
    const error = toSwantaraError(
      axiosErrorWith({ message: "Can't process the data.", error_code: "ERR_BAD_REQUEST" }, 400),
    );
    expect(error).toBeInstanceOf(SwantaraError);
    expect(error.message).toBe("Can't process the data.");
    expect(error.code).toBe("ERR_BAD_REQUEST");
    expect(error.status).toBe(400);
  });

  it("maps ERR_UNAUTHORIZED to SwantaraUnauthorizedError", () => {
    const error = toSwantaraError(
      axiosErrorWith({ message: "Unauthorized.", error_code: "ERR_UNAUTHORIZED" }, 401),
    );
    expect(error).toBeInstanceOf(SwantaraUnauthorizedError);
    expect(error.code).toBe("ERR_UNAUTHORIZED");
    expect(error.status).toBe(401);
  });

  it("maps ERR_FORBIDDEN to SwantaraError", () => {
    const error = toSwantaraError(
      axiosErrorWith({ message: "Forbidden.", error_code: "ERR_FORBIDDEN" }, 403),
    );
    expect(error).toBeInstanceOf(SwantaraError);
    expect(error.code).toBe("ERR_FORBIDDEN");
    expect(error.status).toBe(403);
  });

  it("maps ERR_NOT_FOUND to SwantaraError", () => {
    const error = toSwantaraError(
      axiosErrorWith({ message: "Resource not found.", error_code: "ERR_NOT_FOUND" }, 404),
    );
    expect(error).toBeInstanceOf(SwantaraError);
    expect(error.code).toBe("ERR_NOT_FOUND");
    expect(error.status).toBe(404);
  });

  it("maps ERR_UNPROCESSABLE to SwantaraUnprocessableError with field errors", () => {
    const error = toSwantaraError(
      axiosErrorWith({
        message: "Failed to process request.",
        error_code: "ERR_UNPROCESSABLE",
        field_errors: [{ field: "email", code: "EMAIL", message: "email must be an email" }],
      }),
    );
    expect(error).toBeInstanceOf(SwantaraUnprocessableError);
    expect(error.code).toBe("ERR_UNPROCESSABLE");
    expect(error.status).toBe(422);
    const unprocessable = error as SwantaraUnprocessableError;
    expect(unprocessable.fieldErrors).toEqual([
      { field: "email", code: "EMAIL", message: "email must be an email" },
    ]);
  });

  it("maps ERR_UNPROCESSABLE without field errors to null fieldErrors", () => {
    const error = toSwantaraError(
      axiosErrorWith({
        message: "Failed to process request.",
        error_code: "ERR_UNPROCESSABLE",
      }),
    );
    expect(error).toBeInstanceOf(SwantaraUnprocessableError);
    const unprocessable = error as SwantaraUnprocessableError;
    expect(unprocessable.fieldErrors).toBeNull();
  });

  it("maps ERR_SERVICE to SwantaraError", () => {
    const error = toSwantaraError(
      axiosErrorWith({ message: "Internal server error.", error_code: "ERR_SERVICE" }, 500),
    );
    expect(error).toBeInstanceOf(SwantaraError);
    expect(error.code).toBe("ERR_SERVICE");
    expect(error.status).toBe(500);
  });

  it("maps ERR_SERVICE_UNAVAILABLE to SwantaraError", () => {
    const error = toSwantaraError(
      axiosErrorWith({ message: "Unavailable.", error_code: "ERR_SERVICE_UNAVAILABLE" }, 503),
    );
    expect(error).toBeInstanceOf(SwantaraError);
    expect(error.code).toBe("ERR_SERVICE_UNAVAILABLE");
    expect(error.status).toBe(503);
  });

  it("keeps the raw code and status for unknown error codes", () => {
    const error = toSwantaraError(
      axiosErrorWith({ message: "Unknown.", error_code: "ERR_WHATEVER" }, 418),
    );
    expect(error).toBeInstanceOf(SwantaraError);
    expect(error.code).toBe("ERR_WHATEVER");
    expect(error.status).toBe(418);
  });

  it("falls back to the axios message when the body has no message", () => {
    const error = toSwantaraError(axiosErrorWith({ error_code: "ERR_SERVICE" }, 500));
    expect(error).toBeInstanceOf(SwantaraError);
    expect(error.message).toBe("Request failed with status code 500");
  });

  it("falls back to the axios message when the body has no error code", () => {
    const error = toSwantaraError(axiosErrorWith({ message: "Internal server error." }, 500));
    expect(error).toBeInstanceOf(SwantaraError);
    expect(error.code).toBeNull();
    expect(error.status).toBe(500);
    expect(error.message).toBe("Internal server error.");
  });
});
