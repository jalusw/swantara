import { pino } from "pino";

const level = process.env.NEXT_PUBLIC_LOG_LEVEL ?? "info";

export const logger = pino({
  level,
  browser: { asObject: true },
});
