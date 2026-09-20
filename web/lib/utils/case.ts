import camelcaseKeys from "camelcase-keys";
import snakecaseKeys from "snakecase-keys";

export function toSnakeCase(data: unknown): unknown {
  if (data === null || data === undefined || typeof data !== "object") {
    return data;
  }
  if (data instanceof FormData || data instanceof Blob) {
    return data;
  }
  return snakecaseKeys(data as Record<string, unknown> | readonly Record<string, unknown>[]);
}

export function toCamelCase(data: unknown): unknown {
  if (data === null || data === undefined || typeof data !== "object") {
    return data;
  }
  if (data instanceof Blob) {
    return data;
  }
  return camelcaseKeys(data, { deep: true });
}

const ACRONYMS = new Set([
  "crm",
  "pos",
  "hr",
  "planning",
  "recipe",
  "unit",
  "id",
  "sku",
  "po",
  "rma",
  "productionOrder",
]);

export function humanizeKey(value: string): string {
  const words = value
    .replace(/[_-]+/g, " ")
    .replace(/([a-z0-9])([A-Z])/g, "$1 $2")
    .split(" ")
    .filter(Boolean);
  return words
    .map((word) => {
      if (ACRONYMS.has(word.toLowerCase())) return word.toUpperCase();
      return word.charAt(0).toUpperCase() + word.slice(1);
    })
    .join(" ");
}
