import { isAxiosError } from "axios";
import {
  SwantaraError,
  type SwantaraFieldError,
  SwantaraUnauthorizedError,
  SwantaraUnprocessableError,
} from "./errors";

type SwantaraErrorBody = {
  message?: string;
  error_code?: string;
  field_errors?: SwantaraFieldError[];
};

const codeToErrorClass: Record<string, new (message: string) => SwantaraError> = {
  ERR_UNAUTHORIZED: SwantaraUnauthorizedError,
  ERR_UNPROCESSABLE: SwantaraUnprocessableError,
};

export function toSwantaraError(error: unknown): SwantaraError {
  if (!isAxiosError(error)) {
    return new SwantaraError("Network error.");
  }

  const status = error.response?.status ?? 0;
  const body = error.response?.data as SwantaraErrorBody | undefined;
  const message = body?.message ?? error.message;

  const errorCode = body?.error_code;
  if (!errorCode) {
    return new SwantaraError(message, null, status);
  }

  const ErrorClass = codeToErrorClass[errorCode];
  if (!ErrorClass) {
    return new SwantaraError(message, errorCode, status);
  }

  const mapped = new ErrorClass(message);
  if (mapped instanceof SwantaraUnprocessableError) {
    mapped.fieldErrors = body.field_errors ?? null;
  }

  return mapped;
}
