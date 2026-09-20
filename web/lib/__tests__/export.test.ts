import { afterEach, describe, expect, it, vi } from "vitest";
import { downloadTextFile, exportCsv, toCsv } from "@/lib/utils";

describe("toCsv", () => {
  it("renders headers and rows with CRLF line endings and a UTF-8 BOM", () => {
    const csv = toCsv(
      ["Name", "Status"],
      [
        ["Acme Inc", "active"],
        ["Beta LLC", "paused"],
      ],
    );
    expect(csv).toBe("\uFEFFName,Status\r\nAcme Inc,active\r\nBeta LLC,paused");
  });

  it("quotes cells containing commas, quotes or newlines", () => {
    const csv = toCsv(["Note"], [["Hello, world"]]);
    expect(csv).toContain('"Hello, world"');
    const quoted = toCsv(["Note"], [['He said "hi"']]);
    expect(quoted).toContain('"He said ""hi"""');
  });

  it("normalises null and undefined to empty cells", () => {
    const csv = toCsv(["A", "B"], [[null, undefined]]);
    expect(csv).toBe("\uFEFFA,B\r\n,");
  });

  it("can omit the BOM", () => {
    const csv = toCsv(["A"], [["1"]], { includeBom: false });
    expect(csv.startsWith("\uFEFF")).toBe(false);
  });
});

describe("downloadTextFile", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("creates a blob URL and clicks a download anchor", () => {
    const createObjectURL = vi.fn().mockReturnValue("blob:mock").mockName("URL.createObjectURL");
    const revokeObjectURL = vi.fn().mockName("URL.revokeObjectURL");
    vi.stubGlobal("URL", { ...URL, createObjectURL, revokeObjectURL });
    const click = vi.fn();
    vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(click);

    downloadTextFile("report.txt", "hello");

    expect(createObjectURL).toHaveBeenCalledOnce();
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:mock");
    expect(click).toHaveBeenCalledOnce();
  });
});

describe("exportCsv", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("downloads the serialised CSV with a .csv suffix", () => {
    const click = vi.fn();
    vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(click);
    const download = vi.spyOn(HTMLAnchorElement.prototype, "download", "set");

    exportCsv("customers", ["Name"], [["Acme"]]);

    expect(download).toHaveBeenCalledWith("customers.csv");
    expect(click).toHaveBeenCalledOnce();
  });
});
