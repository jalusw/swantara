import { StatusCodes } from "http-status-codes";

export const ErrorCodes = {
  BAD_REQUEST: "ERR_BAD_REQUEST",
  UNAUTHORIZED: "ERR_UNAUTHORIZED",
  FORBIDDEN: "ERR_FORBIDDEN",
  CONFLICT: "ERR_CONFLICT",
  UNPROCESSABLE: "ERR_UNPROCESSABLE",
  RATE_LIMIT: "ERR_RATE_LIMIT",
  SERVICE: "ERR_SERVICE",
  SERVICE_UNAVAILABLE: "ERR_SERVICE_UNAVAILABLE",
  NOT_FOUND: "ERR_NOT_FOUND",
};

function errorBody(message: string, errorCode?: string) {
  return { success: false, message, errorCode };
}

export function createForbiddenResponse(
  message = "Forbidden.",
  errorCode = ErrorCodes.FORBIDDEN,
): Response {
  return Response.json(errorBody(message, errorCode), {
    status: StatusCodes.FORBIDDEN,
  });
}

export function createBadRequestResponse(
  message = "Bad request.",
  errorCode = ErrorCodes.BAD_REQUEST,
): Response {
  return Response.json(errorBody(message, errorCode), {
    status: StatusCodes.BAD_REQUEST,
  });
}

export function createServiceUnavailableResponse(
  message = "Service unavailable.",
  errorCode = ErrorCodes.SERVICE_UNAVAILABLE,
): Response {
  return Response.json(errorBody(message, errorCode), {
    status: StatusCodes.SERVICE_UNAVAILABLE,
  });
}

export function createNoContentResponse(): Response {
  return new Response(null, { status: StatusCodes.NO_CONTENT });
}

export function createInternalServerErrorResponse(
  message = "Internal server error.",
  errorCode = ErrorCodes.SERVICE,
): Response {
  return Response.json(errorBody(message, errorCode), {
    status: StatusCodes.INTERNAL_SERVER_ERROR,
  });
}

export function createErrorResponse(error: unknown): Response {
  if (error instanceof Error) {
    const name = error.constructor.name;
    if (name === "SwantaraBadRequestError") {
      return createBadRequestResponse(error.message);
    }
    if (name === "SwantaraUnauthorizedError") {
      return Response.json(errorBody("Unauthorized.", ErrorCodes.UNAUTHORIZED), {
        status: StatusCodes.UNAUTHORIZED,
      });
    }
    if (name === "SwantaraForbiddenError") {
      return createForbiddenResponse(error.message);
    }
    if (name === "SwantaraNotFoundError") {
      return Response.json(errorBody(error.message, ErrorCodes.NOT_FOUND), {
        status: StatusCodes.NOT_FOUND,
      });
    }
    if (name === "SwantaraUnprocessableError") {
      return Response.json(errorBody(error.message, ErrorCodes.UNPROCESSABLE), {
        status: StatusCodes.UNPROCESSABLE_ENTITY,
      });
    }
  }
  return createInternalServerErrorResponse();
}
