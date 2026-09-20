// @vitest-environment node
import { StatusCodes } from "http-status-codes";
import { describe, expect, it } from "vitest";
import { createBadRequestResponse, createErrorResponse, ErrorCodes } from "../response";

class SwantaraBadRequestError extends Error {}
class SwantaraUnauthorizedError extends Error {}
class SwantaraForbiddenError extends Error {}
class SwantaraNotFoundError extends Error {}
class SwantaraUnprocessableError extends Error {}

async function readJson(response: Response) {
  return { status: response.status, body: await response.json() };
}

describe("createBadRequestResponse", () => {
  it("builds a 400 response with defaults", async () => {
    const { status, body } = await readJson(createBadRequestResponse());
    expect(status).toBe(StatusCodes.BAD_REQUEST);
    expect(body).toEqual({
      success: false,
      message: "Bad request.",
      errorCode: ErrorCodes.BAD_REQUEST,
    });
  });

  it("uses a custom message and error code", async () => {
    const { status, body } = await readJson(
      createBadRequestResponse("Invalid email.", "ERR_CUSTOM"),
    );
    expect(status).toBe(StatusCodes.BAD_REQUEST);
    expect(body.message).toBe("Invalid email.");
    expect(body.errorCode).toBe("ERR_CUSTOM");
  });
});

describe("createErrorResponse mappings", () => {
  it("maps SwantaraBadRequestError to 400 with the error message", async () => {
    const { status, body } = await readJson(
      createErrorResponse(new SwantaraBadRequestError("Bad input.")),
    );
    expect(status).toBe(StatusCodes.BAD_REQUEST);
    expect(body.message).toBe("Bad input.");
    expect(body.errorCode).toBe(ErrorCodes.BAD_REQUEST);
  });

  it("maps SwantaraUnauthorizedError to 401", async () => {
    const { status, body } = await readJson(
      createErrorResponse(new SwantaraUnauthorizedError("Nope.")),
    );
    expect(status).toBe(StatusCodes.UNAUTHORIZED);
    expect(body.errorCode).toBe(ErrorCodes.UNAUTHORIZED);
  });

  it("maps SwantaraForbiddenError to 403 with the error message", async () => {
    const { status, body } = await readJson(
      createErrorResponse(new SwantaraForbiddenError("Denied.")),
    );
    expect(status).toBe(StatusCodes.FORBIDDEN);
    expect(body.message).toBe("Denied.");
    expect(body.errorCode).toBe(ErrorCodes.FORBIDDEN);
  });

  it("maps SwantaraNotFoundError to 404 with the error message", async () => {
    const { status, body } = await readJson(
      createErrorResponse(new SwantaraNotFoundError("Missing.")),
    );
    expect(status).toBe(StatusCodes.NOT_FOUND);
    expect(body.message).toBe("Missing.");
    expect(body.errorCode).toBe(ErrorCodes.NOT_FOUND);
  });

  it("maps SwantaraUnprocessableError to 422 with the error message", async () => {
    const { status, body } = await readJson(
      createErrorResponse(new SwantaraUnprocessableError("Invalid.")),
    );
    expect(status).toBe(StatusCodes.UNPROCESSABLE_ENTITY);
    expect(body.message).toBe("Invalid.");
    expect(body.errorCode).toBe(ErrorCodes.UNPROCESSABLE);
  });

  it("maps unknown errors to 500", async () => {
    const { status, body } = await readJson(createErrorResponse(new TypeError("Unexpected.")));
    expect(status).toBe(StatusCodes.INTERNAL_SERVER_ERROR);
    expect(body.errorCode).toBe(ErrorCodes.SERVICE);
  });

  it("maps non-Error values to 500", async () => {
    const first = await readJson(createErrorResponse("boom"));
    expect(first.status).toBe(StatusCodes.INTERNAL_SERVER_ERROR);
    const second = await readJson(createErrorResponse(null));
    expect(second.status).toBe(StatusCodes.INTERNAL_SERVER_ERROR);
    expect(second.body.errorCode).toBe(ErrorCodes.SERVICE);
  });
});
