import { describe, expect, it } from "vitest";
import { parseCsv } from "@/lib/utils";

describe("parseCsv", () => {
  it("parses a simple spreadsheet", () => {
    expect(parseCsv("name,status\nAcme,active\nBeta,paused")).toEqual([
      ["name", "status"],
      ["Acme", "active"],
      ["Beta", "paused"],
    ]);
  });

  it("handles CRLF and a trailing newline", () => {
    expect(parseCsv("a,b\r\n1,2\r\n")).toEqual([
      ["a", "b"],
      ["1", "2"],
    ]);
  });

  it("keeps quoted fields with commas, quotes and newlines intact", () => {
    expect(parseCsv('note\n"hello, world"\n"say ""hi"""')).toEqual([
      ["note"],
      ["hello, world"],
      ['say "hi"'],
    ]);
  });

  it("keeps empty trailing fields", () => {
    expect(parseCsv("a,b,c\n1,,")).toEqual([
      ["a", "b", "c"],
      ["1", "", ""],
    ]);
  });

  it("returns empty array for empty input", () => {
    expect(parseCsv("")).toEqual([]);
  });
});
