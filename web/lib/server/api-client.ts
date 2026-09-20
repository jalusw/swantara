import { cookies } from "next/headers";
import { accessTokenKey } from "@/lib/constants/cookies";
import { serverConfig } from "@/lib/server/config";
import { SwantaraService } from "@/lib/services/swantara";

export async function createServerSwantaraService(): Promise<SwantaraService> {
  const origin = new URL(serverConfig.serviceApiUrl).origin;
  const service = new SwantaraService(origin, {
    getHeaders: async (): Promise<Record<string, string>> => {
      const cookieStore = await cookies();
      const accessToken = cookieStore.get(accessTokenKey)?.value;
      return accessToken ? { Authorization: `Bearer ${accessToken}` } : {};
    },
    refreshToken: async () => null,
  });
  return service;
}
