import type { NextRequest } from "next/server";
import { csrfKey } from "@/lib/constants/cookies";

const CSRF_HEADER = "x-csrf-token";
const MUTATING_METHODS = new Set(["POST", "PUT", "PATCH", "DELETE"]);

const CSRF_EXEMPT_PATHS = new Set([
  "/api/v1/auth/login",
  "/api/v1/auth/register",
  "/api/v1/auth/email/check",
  "/api/v1/auth/email-verification/request",
  "/api/v1/auth/email-verification/verify",
  "/api/v1/auth/password-reset/request",
  "/api/v1/auth/password-reset",
]);

export function isCsrfValid(request: NextRequest): boolean {
  if (!MUTATING_METHODS.has(request.method)) return true;
  if (CSRF_EXEMPT_PATHS.has(request.nextUrl.pathname)) return true;

  const headerToken = request.headers.get(CSRF_HEADER);
  const cookieToken = request.cookies.get(csrfKey)?.value;

  if (headerToken && cookieToken && headerToken === cookieToken) return true;

  return false;
}
