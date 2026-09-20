export type LogConfig = {
  level: string;
  pretty: boolean;
  file: string | undefined;
  fileLevel: string | undefined;
  compress: boolean;
};

export type ServerConfig = {
  appName: string;
  serviceApiUrl: string;
  log: LogConfig;
};

if (!process.env.SWANTARA_API_URL) {
  throw new Error("API_URL IS NOT SET");
}

const serverConfig: ServerConfig = {
  appName: process.env.APP_NAME || "",
  serviceApiUrl: process.env.SWANTARA_API_URL || "",
  log: {
    level: process.env.LOG_LEVEL ?? "info",
    pretty: process.env.LOG_PRETTY === "true" || process.env.NODE_ENV === "development",
    file: process.env.LOG_FILE || undefined,
    fileLevel: process.env.LOG_FILE_LEVEL || undefined,
    compress: process.env.LOG_COMPRESS === "true",
  },
};

export { serverConfig };
