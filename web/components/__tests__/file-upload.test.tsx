import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { useFileUpload } from "@/components/file-upload";

describe("useFileUpload", () => {
  it("adds files to the queue", async () => {
    const { result } = renderHook(() => useFileUpload());
    const file = new File(["content"], "invoice.pdf", {
      type: "application/pdf",
    });
    await act(async () => {
      await result.current.addFiles([file]);
    });
    expect(result.current.files).toHaveLength(1);
    expect(result.current.files[0]?.file.name).toBe("invoice.pdf");
  });

  it("flags files over the max size", async () => {
    const { result } = renderHook(() => useFileUpload({ maxSize: 5 }));
    const file = new File(["a very long payload"], "big.txt");
    await act(async () => {
      await result.current.addFiles([file]);
    });
    expect(result.current.files[0]?.status).toBe("error");
    expect(result.current.files[0]?.error).toMatch(/exceeds/i);
  });

  it("rejects files not matching the accept list", async () => {
    const { result } = renderHook(() => useFileUpload({ accept: [".csv"] }));
    const file = new File(["x"], "data.xlsx");
    await act(async () => {
      await result.current.addFiles([file]);
    });
    expect(result.current.files[0]?.status).toBe("error");
  });

  it("marks uploads done after onUpload resolves", async () => {
    const { result } = renderHook(() => useFileUpload({}, async () => {}));
    const file = new File(["x"], "data.csv");
    await act(async () => {
      await result.current.addFiles([file]);
    });
    expect(result.current.files[0]?.status).toBe("done");
  });
});
