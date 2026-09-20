import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { authMiddleware } from "./lib/middleware/auth";
import {
  applySessionCookiesToResponse,
  refreshSessionInMiddleware,
} from "./lib/middleware/refresh";

export default async function proxy(request: NextRequest) {
  const refresh = await refreshSessionInMiddleware(request);

  const auth = await authMiddleware(request);
  let final: NextResponse;
  try {
    final = auth ?? NextResponse.next();
  } catch {
    const url = request.nextUrl.clone();
    url.pathname = "/login";
    final = NextResponse.redirect(url);
  }

  applySessionCookiesToResponse(final, refresh);

  return final;
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|.*\\.(?:png|jpg|jpeg|gif|webp|avif|svg|ico)$).*)"],
};
