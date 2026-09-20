export const accessTokenKey = "access_token";
export const refreshTokenKey = "refresh_token";
export const csrfKey = "csrf_token";

export function hasAccessToken(): boolean {
  if (typeof document === "undefined") return false;
  return hasSessionHint();
}

export function hasSessionHint(): boolean {
  if (typeof document === "undefined") return false;
  return document.cookie.split("; ").some((c) => c.startsWith(`${csrfKey}=`));
}

export function getCsrfToken(): string | undefined {
  if (typeof document === "undefined") return undefined;
  return document.cookie
    .split("; ")
    .find((c) => c.startsWith(`${csrfKey}=`))
    ?.split("=")
    .slice(1)
    .join("=");
}
