import { isAxiosError } from "axios";
import { StatusCodes } from "http-status-codes";

export function isServerError(error: unknown): boolean {
  return isAxiosError(error) && (error.response?.status ?? 0) >= StatusCodes.INTERNAL_SERVER_ERROR;
}

export function isNetworkError(error: unknown): boolean {
  return !isAxiosError(error);
}
