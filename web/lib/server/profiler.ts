import { formatDuration } from "@/lib/utils";
import { formatPerfLog, perfLogLevel } from "@/lib/utils/perf";

import { logger } from "./logger";

type ProfilerContext = {
  method: string;
  path: string;
  startTime: number;
};

type ProfileResult = {
  method: string;
  path: string;
  duration: string;
  statusCode: number;
  contentLength: string | null;
  memory: {
    rss: string;
    heapUsed: string;
    heapTotal: string;
  };
};

function formatMemoryMB(bytes: number): string {
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`;
}

function getMemoryUsage() {
  const mem = process.memoryUsage();
  return {
    rss: formatMemoryMB(mem.rss),
    heapUsed: formatMemoryMB(mem.heapUsed),
    heapTotal: formatMemoryMB(mem.heapTotal),
  };
}

export function createProfiler(method: string, path: string): ProfilerContext {
  return {
    method,
    path,
    startTime: performance.now(),
  };
}

function formatContentLength(length: string | null): string | null {
  if (length === null) return null;
  return `${length} B`;
}

export function finishProfiler(ctx: ProfilerContext, response: Response): Response {
  const durationMs = performance.now() - ctx.startTime;
  const result: ProfileResult = {
    method: ctx.method,
    path: ctx.path,
    duration: formatDuration(durationMs),
    statusCode: response.status,
    contentLength: formatContentLength(response.headers.get("content-length")),
    memory: getMemoryUsage(),
  };

  const logMsg = formatPerfLog(ctx.method, ctx.path, response.status, result.duration);

  const level = perfLogLevel(durationMs, response.status);
  if (level === "warn") {
    logger.warn(result, logMsg);
  } else if (level === "error") {
    logger.error(result, logMsg);
  } else {
    logger.info(result, logMsg);
  }

  return response;
}

export function finishProfilerWithError(ctx: ProfilerContext, error: unknown): void {
  const durationMs = performance.now() - ctx.startTime;
  const duration = formatDuration(durationMs);
  const logMsg = formatPerfLog(ctx.method, ctx.path, "ERROR", duration);
  logger.error(
    {
      method: ctx.method,
      path: ctx.path,
      duration,
      memory: getMemoryUsage(),
      error,
    },
    logMsg,
  );
}
