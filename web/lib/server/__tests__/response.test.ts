// @vitest-environment node
import { StatusCodes } from "http-status-codes";
import { describe, expect, it } from "vitest";
import {
  createErrorResponse,
  createForbiddenResponse,
  createInternalServerErrorResponse,
  createNoContentResponse,
  createServiceUnavailableResponse,
  ErrorCodes,
} from "../response";

async function readJson(response: Response) {
  return { status: response.status, body: await response.json() };
}

describe("response utilities", () => {
  it("builds a 204 no-content response", () => {
    const response = createNoContentResponse();
    expect(response.status).toBe(StatusCodes.NO_CONTENT);
  });

  it("builds a 403 forbidden response", async () => {
    const { status, body } = await readJson(createForbiddenResponse());
    expect(status).toBe(StatusCodes.FORBIDDEN);
    expect(body.success).toBe(false);
    expect(body.errorCode).toBe(ErrorCodes.FORBIDDEN);
  });

  it("builds a 503 service unavailable response", async () => {
    const { status, body } = await readJson(createServiceUnavailableResponse());
    expect(status).toBe(StatusCodes.SERVICE_UNAVAILABLE);
    expect(body.success).toBe(false);
    expect(body.errorCode).toBe(ErrorCodes.SERVICE_UNAVAILABLE);
  });

  it("builds a 500 internal server error response", async () => {
    const { status, body } = await readJson(createInternalServerErrorResponse());
    expect(status).toBe(StatusCodes.INTERNAL_SERVER_ERROR);
    expect(body.success).toBe(false);
    expect(body.errorCode).toBe(ErrorCodes.SERVICE);
  });

  it("createErrorResponse always yields a 500", async () => {
    const { status, body } = await readJson(createErrorResponse(new Error("x")));
    expect(status).toBe(StatusCodes.INTERNAL_SERVER_ERROR);
    expect(body.errorCode).toBe(ErrorCodes.SERVICE);
  });
});
