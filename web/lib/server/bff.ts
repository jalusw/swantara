import type { SessionTokens } from "./session";

export function extractSessionTokens(payload: unknown): SessionTokens | null {
  const rawData = (payload as { data?: Record<string, unknown> })?.data as
    | {
        accessToken?: string;
        refreshToken?: string;
        access_token?: string;
        refresh_token?: string;
      }
    | undefined;
  const accessToken = rawData?.accessToken ?? rawData?.access_token;
  const refreshToken = rawData?.refreshToken ?? rawData?.refresh_token;
  if (accessToken && refreshToken) {
    return { accessToken, refreshToken };
  }
  return null;
}
