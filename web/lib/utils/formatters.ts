type FormatterOptions = {
  locale?: string;
  timeZone?: string;
  maximumFractionDigits?: number;
  nullFallback?: string;
};

function resolveLocale(locale?: string): string {
  return locale ?? "id-ID";
}

export function formatNumber(value: number, options: FormatterOptions = {}): string {
  const { locale, maximumFractionDigits = 2 } = options;
  return new Intl.NumberFormat(resolveLocale(locale), {
    style: "decimal",
    maximumFractionDigits,
  }).format(value);
}

export function formatDate(
  value: Date | string | number | null | undefined,
  options: FormatterOptions = {},
): string {
  if (value == null) return options.nullFallback ?? "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return options.nullFallback ?? "";
  return new Intl.DateTimeFormat(resolveLocale(options.locale), {
    dateStyle: "medium",
    timeZone: options.timeZone,
  }).format(date);
}

export function formatDateTime(
  value: Date | string | number | null | undefined,
  options: FormatterOptions = {},
): string {
  if (value == null) return options.nullFallback ?? "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return options.nullFallback ?? "";
  return new Intl.DateTimeFormat(resolveLocale(options.locale), {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: options.timeZone,
  }).format(date);
}

export function formatBytes(bytes: number, options: FormatterOptions = {}): string {
  if (!Number.isFinite(bytes) || bytes === 0) {
    return "0 B";
  }
  const units = ["B", "KB", "MB", "GB", "TB"];
  const unitIndex = Math.min(
    Math.floor(Math.log(Math.abs(bytes)) / Math.log(1024)),
    units.length - 1,
  );
  const value = bytes / 1024 ** unitIndex;
  const maximumFractionDigits = unitIndex === 0 ? 0 : 1;
  return `${new Intl.NumberFormat(resolveLocale(options.locale), {
    maximumFractionDigits,
  }).format(value)} ${units[unitIndex]}`;
}

export function formatMoney(
  value: number | null | undefined,
  options: { locale?: string; currency?: string; nullFallback?: string } = {},
): string {
  const { locale = "id-ID", currency, nullFallback } = options;
  if (value == null) return nullFallback ?? "";
  return new Intl.NumberFormat(locale, {
    style: currency ? "currency" : "decimal",
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}

export function formatDuration(ms: number): string {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)}s`;
  }
  return `${ms.toFixed(2)}ms`;
}

export function formatHours(hours: number): string {
  return `${hours.toFixed(1)}h`;
}

export function getLocalDateString(date: Date = new Date()): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}
