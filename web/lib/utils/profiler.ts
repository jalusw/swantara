import type { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from "axios";
import { formatDuration } from "@/lib/utils";
import { logger } from "@/lib/utils/logger";
import { formatPerfLog, perfLogLevel } from "@/lib/utils/perf";

type RequestProfile = {
  method: string;
  url: string;
  startTime: number;
};

type ProfileResult = {
  method: string;
  url: string;
  status: number;
  duration: string;
};

const activeProfiles = new WeakMap<InternalAxiosRequestConfig, RequestProfile>();

export function startRequestProfile(config: InternalAxiosRequestConfig): void {
  activeProfiles.set(config, {
    method: (config.method ?? "GET").toUpperCase(),
    url: config.url ?? "?",
    startTime: performance.now(),
  });
}

function takeProfile(config: InternalAxiosRequestConfig): RequestProfile | null {
  const profile = activeProfiles.get(config);
  if (profile === undefined) {
    return null;
  }
  activeProfiles.delete(config);
  return profile;
}

function finishProfile(profile: RequestProfile, status: number, error?: unknown): void {
  const durationMs = performance.now() - profile.startTime;
  const result: ProfileResult = {
    method: profile.method,
    url: profile.url,
    status,
    duration: formatDuration(durationMs),
  };
  const outcome = error === undefined ? status : "ERROR";
  const logMsg = formatPerfLog(profile.method, profile.url, outcome, result.duration);

  const level = perfLogLevel(durationMs, status, error !== undefined);
  if (level === "error") {
    if (error !== undefined) {
      logger.error({ ...result, error }, logMsg);
    } else {
      logger.error(result, logMsg);
    }
  } else if (level === "warn") {
    logger.warn(result, logMsg);
  } else {
    logger.info(result, logMsg);
  }
}

export function finishResponseProfile(response: AxiosResponse): void {
  const profile = takeProfile(response.config);
  if (profile === null) {
    return;
  }
  finishProfile(profile, response.status);
}

export function finishErrorProfile(error: AxiosError): void {
  const config = error.config;
  if (config === undefined) {
    return;
  }
  const profile = takeProfile(config);
  if (profile === null) {
    return;
  }
  finishProfile(profile, error.response?.status ?? 0, error);
}
