import { createWriteStream } from "node:fs";
import { createGzip } from "node:zlib";
import { multistream, pino } from "pino";
import pretty from "pino-pretty";
import { serverConfig as cfg } from "./config";

const PERF_PREFIX = "\u001b[38;5;214m[PERF]\u001b[0m";
const PREFIX_INFO = "\u001b[97m[INFO]\u001b[0m";
const PREFIX_WARN = "\u001b[93m[WARN]\u001b[0m";
const PREFIX_ERROR = "\u001b[91m[ERROR]\u001b[0m";

const streams = [];

if (cfg.log.pretty) {
  streams.push({
    stream: pretty({
      colorize: true,
      translateTime: "HH:MM:ss.l",
      ignore: "pid,hostname",
      messageFormat: (log: Record<string, unknown>, messageKey: string) => {
        const msg = log[messageKey];
        if (typeof msg === "string" && msg.startsWith("[PERF]")) {
          return msg.slice(6);
        }
        return String(msg);
      },
      customPrettifiers: {
        level: (
          inputData: string | object,
          _key: string,
          log: object,
          { label }: { label: string },
        ) => {
          const msg = (log as Record<string, unknown>).msg ?? inputData;
          if (typeof msg === "string" && msg.startsWith("[PERF]")) {
            return PERF_PREFIX;
          }
          switch (label) {
            case "INFO":
              return PREFIX_INFO;
            case "WARN":
              return PREFIX_WARN;
            case "ERROR":
            case "FATAL":
              return PREFIX_ERROR;
            default:
              return `[${label}]`;
          }
        },
      },
    }),
  });
}

if (cfg.log.file) {
  const filePath = cfg.log.compress ? `${cfg.log.file}.gz` : cfg.log.file;
  const fileStream = createWriteStream(filePath, { flags: "a" });
  const output = cfg.log.compress ? createGzip().pipe(fileStream) : fileStream;

  streams.push({
    level: cfg.log.fileLevel ?? cfg.log.level,
    stream: output,
  });
}

export const logger =
  streams.length > 0
    ? pino({ level: cfg.log.level }, multistream(streams))
    : pino({ level: cfg.log.level });
