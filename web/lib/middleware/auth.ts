import { type NextRequest, NextResponse } from "next/server";
import { isProtectedPage } from "@/lib/utils";
import { accessTokenKey } from "../constants/cookies";

const guestPaths = ["/login", "/register"];

export async function authMiddleware(request: NextRequest) {
  const url = request.nextUrl.clone();
  const accessToken = request.cookies.get(accessTokenKey);
  const rawPath = url.pathname;
  const path = rawPath !== "/" ? rawPath.replace(/\/$/, "") : rawPath;

  if (path === "/") {
    url.pathname = accessToken?.value ? "/onboarding" : "/login";
    return NextResponse.redirect(url);
  }

  if (accessToken?.value && guestPaths.includes(path)) {
    url.pathname = "/onboarding";
    return NextResponse.redirect(url);
  }

  if (isProtectedPage(path) && !accessToken?.value) {
    url.pathname = "/login";
    return NextResponse.redirect(url);
  }

  return null;
}
