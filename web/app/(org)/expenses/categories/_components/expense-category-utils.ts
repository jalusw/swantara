export function formatAccountName(code: string | null, name: string | null): string {
  if (!code && !name) return "—";
  if (!code) return name ?? "—";
  if (!name) return code;
  return `${code} - ${name}`;
}
