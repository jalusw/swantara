import { act, renderHook, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { FileUpload, useFileUpload } from "@/components/file-upload";
import { renderWithProviders } from "@/lib/tests";

describe("useFileUpload extra2", () => {
  it("returns an empty list for null input", async () => {
    const { result } = renderHook(() => useFileUpload());
    let added: unknown[] = [];
    await act(async () => {
      added = await result.current.addFiles(null);
    });
    expect(added).toHaveLength(0);
    expect(result.current.files).toHaveLength(0);
  });

  it("removes a file from the queue", async () => {
    const { result } = renderHook(() => useFileUpload());
    const file = new File(["content"], "invoice.pdf", { type: "application/pdf" });
    await act(async () => {
      await result.current.addFiles([file]);
    });
    const id = result.current.files[0]?.id as string;
    act(() => {
      result.current.removeFile(id);
    });
    expect(result.current.files).toHaveLength(0);
  });

  it("reports upload progress and failure", async () => {
    const seen: number[] = [];
    const { result } = renderHook(() =>
      useFileUpload({}, async (_entry, onProgress) => {
        onProgress(42);
        throw new Error("network down");
      }),
    );
    const file = new File(["x"], "data.csv");
    await act(async () => {
      await result.current.addFiles([file]);
    });
    expect(result.current.files[0]?.status).toBe("error");
    expect(result.current.files[0]?.error).toBe("Upload failed.");
    expect(seen).toHaveLength(0);
  });

  it("accepts files by mime type", async () => {
    const { result } = renderHook(() =>
      useFileUpload({ accept: ["application/pdf"] }, async () => {}),
    );
    const file = new File(["content"], "invoice.pdf", { type: "application/pdf" });
    await act(async () => {
      await result.current.addFiles([file]);
    });
    expect(result.current.files[0]?.status).toBe("done");
  });

  it("matches extension patterns case-insensitively", async () => {
    const { result } = renderHook(() => useFileUpload({ accept: [".CSV"] }, async () => {}));
    const file = new File(["x"], "DATA.CSV");
    await act(async () => {
      await result.current.addFiles([file]);
    });
    expect(result.current.files[0]?.status).toBe("done");
  });
});

describe("FileUpload extra2", () => {
  it("renders the dropzone label and description", () => {
    renderWithProviders(<FileUpload label="Attach invoice" description="PDF only" />);

    expect(screen.getByText("Attach invoice")).toBeInTheDocument();
    expect(screen.getByText("PDF only")).toBeInTheDocument();
  });

  it("lists chosen files and removes them", async () => {
    const user = userEvent.setup();
    renderWithProviders(<FileUpload />);
    const file = new File(["content"], "invoice.pdf", { type: "application/pdf" });

    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    await user.upload(input, file);

    expect(await screen.findByText("invoice.pdf")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Remove invoice.pdf" }));

    expect(screen.queryByText("invoice.pdf")).not.toBeInTheDocument();
  });

  it("shows validation errors for rejected files", async () => {
    const user = userEvent.setup();
    renderWithProviders(<FileUpload maxSize={5} />);
    const file = new File(["a very long payload"], "big.txt");

    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    await user.upload(input, file);

    expect(await screen.findByRole("alert")).toHaveTextContent(/exceeds/i);
  });

  it("notifies on files change", async () => {
    const onFilesChange = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(<FileUpload onFilesChange={onFilesChange} />);
    const file = new File(["x"], "note.txt");

    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    await user.upload(input, file);

    expect(await screen.findByText("note.txt")).toBeInTheDocument();
    expect(onFilesChange).toHaveBeenCalled();
  });
});
