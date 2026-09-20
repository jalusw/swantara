import { logger } from "./logger";
import { createProfiler, finishProfiler } from "./profiler";
import { createErrorResponse } from "./response";

export function withHandler<Args extends unknown[]>(
  method: string,
  path: string,
  handler: (...args: Args) => Response | Promise<Response>,
  onError: (error: unknown) => Response = createErrorResponse,
) {
  return async (...args: Args): Promise<Response> => {
    const profiler = createProfiler(method, path);
    try {
      return finishProfiler(profiler, await handler(...args));
    } catch (error: unknown) {
      logger.error(error, `${method} ${path} failed`);
      return finishProfiler(profiler, onError(error));
    }
  };
}
