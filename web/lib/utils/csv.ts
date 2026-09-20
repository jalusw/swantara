import Papa from "papaparse";

export function parseCsv(text: string): string[][] {
  if (!text) {
    return [];
  }
  const result = Papa.parse<string[]>(text, {
    header: false,
    skipEmptyLines: true,
  });
  return result.data;
}

type CsvOptions = {
  includeBom?: boolean;
};

export function toCsv(headers: string[], rows: unknown[][], options: CsvOptions = {}): string {
  const body = Papa.unparse([headers, ...rows], { newline: "\r\n" });
  return (options.includeBom ?? true) ? `\uFEFF${body}` : body;
}

export function exportCsv(
  filename: string,
  headers: string[],
  rows: unknown[][],
  options: CsvOptions = {},
): void {
  downloadTextFile(`${filename}.csv`, toCsv(headers, rows, options), {
    mimeType: "text/csv;charset=utf-8;",
  });
}

export function downloadTextFile(
  filename: string,
  content: string,
  options: { mimeType?: string } = {},
): void {
  const blob = new Blob([content], { type: options.mimeType ?? "text/plain" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
}
