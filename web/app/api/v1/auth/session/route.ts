import { type NextRequest, NextResponse } from "next/server";
import { csrfKey } from "@/lib/constants/cookies";
import { extractSessionTokens } from "@/lib/server/bff";
import { serverConfig } from "@/lib/server/config";
import { isCsrfValid } from "@/lib/server/csrf";
import { getRefreshToken, setSessionCookies } from "@/lib/server/session";

const CSRF_MAX_AGE = 48 * 60 * 60;
const MAX_TOKEN_LENGTH = 2048;

export async function POST(request: NextRequest): Promise<NextResponse> {
  if (!isCsrfValid(request)) {
    return NextResponse.json({ error: "forbidden" }, { status: 403 });
  }

  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ error: "invalid body" }, { status: 400 });
  }
  const { accessToken, refreshToken } = body as {
    accessToken?: string;
    refreshToken?: string;
  };

  if (!accessToken || !refreshToken) {
    return NextResponse.json({ error: "missing tokens" }, { status: 400 });
  }

  if (
    accessToken.length < 10 ||
    refreshToken.length < 10 ||
    accessToken.length > MAX_TOKEN_LENGTH ||
    refreshToken.length > MAX_TOKEN_LENGTH
  ) {
    return NextResponse.json({ error: "invalid tokens" }, { status: 400 });
  }

  const csrfToken = crypto.randomUUID();

  const response = NextResponse.json({ ok: true });
  setSessionCookies(response, { accessToken, refreshToken });

  response.cookies.set(csrfKey, csrfToken, {
    httpOnly: false,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: CSRF_MAX_AGE,
  });

  return response;
}

const REFRESH_TIMEOUT_MS = 10_000;

function resolveNext(request: NextRequest): string {
  const next = request.nextUrl.searchParams.get("next") ?? "/onboarding";
  return next.startsWith("/") && !next.startsWith("//") ? next : "/onboarding";
}

export async function GET(request: NextRequest): Promise<NextResponse> {
  const destination = resolveNext(request);
  const refreshToken = getRefreshToken(request);
  if (refreshToken) {
    const tokens = await refreshSession(refreshToken);
    if (tokens) {
      const response = NextResponse.redirect(new URL(destination, request.url));
      setSessionCookies(response, tokens);
      return response;
    }
  }
  return NextResponse.redirect(new URL("/login", request.url));
}

async function refreshSession(
  refreshToken: string,
): Promise<{ accessToken: string; refreshToken: string } | null> {
  const target = new URL("/api/v1/auth/refresh", serverConfig.serviceApiUrl);
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), REFRESH_TIMEOUT_MS);
  try {
    const upstream = await fetch(target, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
      signal: controller.signal,
    });
    if (!upstream.ok) return null;
    return extractSessionTokens(await upstream.json().catch(() => null));
  } catch {
    return null;
  } finally {
    clearTimeout(timeout);
  }
}
