import { fireEvent, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { PdfViewer } from "../pdf-viewer";
import { PdfViewerStatus } from "../pdf-viewer-status";
import { PdfViewerToolbar } from "../pdf-viewer-toolbar";

vi.mock("react-pdf", async () => {
  const React = await import("react");

  const Document = ({
    file,
    onLoadSuccess,
    onLoadError,
    loading,
    error,
    children,
  }: {
    file: string;
    onLoadSuccess?: (data: { numPages: number; getData: () => Promise<Uint8Array> }) => void;
    onLoadError?: (error: Error) => void;
    loading: React.ReactNode;
    error: React.ReactNode;
    children: React.ReactNode;
  }) => {
    const [failed, setFailed] = React.useState(false);
    const [loaded, setLoaded] = React.useState(false);
    React.useEffect(() => {
      if (file === "bad.pdf") {
        onLoadError?.(new Error("boom"));
        setFailed(true);
      } else {
        onLoadSuccess?.({ numPages: 5, getData: async () => new Uint8Array([1, 2, 3]) });
        setLoaded(true);
      }
    }, [file]);
    if (failed) {
      return error;
    }
    if (!loaded) {
      return loading;
    }
    return children;
  };

  const Page = (props: Record<string, unknown>) =>
    React.createElement("div", { "data-testid": "page", ...props });

  const pdfjs = { GlobalWorkerOptions: { workerSrc: "" } };

  return { Document, Page, pdfjs };
});

function renderToolbar(extra?: Partial<React.ComponentProps<typeof PdfViewerToolbar>>) {
  return renderWithProviders(
    <PdfViewerToolbar
      page={2}
      pages={5}
      scale={1}
      fitWidth={false}
      canFileActions
      onOpen={vi.fn()}
      onPrint={vi.fn()}
      onDownload={vi.fn()}
      onPageChange={vi.fn()}
      onZoomChange={vi.fn()}
      onFitWidthChange={vi.fn()}
      onRotateChange={vi.fn()}
      annotationTool="select"
      hasAnnotations={false}
      onAnnotationToolChange={vi.fn()}
      onClearAnnotations={vi.fn()}
      {...extra}
    />,
  );
}

describe("PdfViewerStatus", () => {
  it("renders loading", () => {
    renderWithProviders(<PdfViewerStatus type="loading" />);
    expect(screen.getByText("Loading document…")).toBeInTheDocument();
  });

  it("renders error with retry", () => {
    renderWithProviders(<PdfViewerStatus type="error" onRetry={() => {}} />);
    expect(screen.getByText("Could not load this PDF.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Try again" })).toBeInTheDocument();
  });

  it("renders empty state", () => {
    renderWithProviders(<PdfViewerStatus type="empty" />);
    expect(screen.getByText("No preview available")).toBeInTheDocument();
  });
});

describe("PdfViewerToolbar", () => {
  it("disables prev/next at page bounds", () => {
    renderToolbar({ page: 1, pages: 1 });
    expect(screen.getByRole("button", { name: "Previous page" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Next page" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "First page" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Last page" })).toBeDisabled();
  });

  it("reports first/last/next page jumps", async () => {
    const user = userEvent.setup();
    const onPageChange = vi.fn();
    renderToolbar({ onPageChange });

    await user.click(screen.getByRole("button", { name: "Next page" }));
    expect(onPageChange).toHaveBeenCalledWith(3);

    await user.click(screen.getByRole("button", { name: "First page" }));
    expect(onPageChange).toHaveBeenCalledWith(1);

    await user.click(screen.getByRole("button", { name: "Last page" }));
    expect(onPageChange).toHaveBeenCalledWith(5);
  });

  it("steps zoom within presets", async () => {
    const user = userEvent.setup();
    const onZoomChange = vi.fn();
    renderToolbar({ scale: 1, onZoomChange });

    await user.click(screen.getByRole("button", { name: "Zoom in" }));
    expect(onZoomChange).toHaveBeenCalledWith(1.25);

    await user.click(screen.getByRole("button", { name: "Zoom out" }));
    expect(onZoomChange).toHaveBeenCalledWith(0.75);
  });

  it("reports rotate actions in both directions", async () => {
    const user = userEvent.setup();
    const onRotateChange = vi.fn();
    renderToolbar({ onRotateChange });
    await user.click(screen.getByRole("button", { name: "Rotate page" }));
    expect(onRotateChange).toHaveBeenCalledWith(90);
    await user.click(screen.getByRole("button", { name: "Rotate page left" }));
    expect(onRotateChange).toHaveBeenCalledWith(-90);
  });

  it("reports file actions when the document is loaded", async () => {
    const user = userEvent.setup();
    const onOpen = vi.fn();
    const onPrint = vi.fn();
    const onDownload = vi.fn();
    renderToolbar({ onOpen, onPrint, onDownload });

    await user.click(screen.getByRole("button", { name: "Open in new tab" }));
    await user.click(screen.getByRole("button", { name: "Print" }));
    await user.click(screen.getByRole("button", { name: "Download" }));

    expect(onOpen).toHaveBeenCalled();
    expect(onPrint).toHaveBeenCalled();
    expect(onDownload).toHaveBeenCalled();
  });

  it("disables file actions while the document has not loaded", () => {
    renderToolbar({ canFileActions: false });
    expect(screen.getByRole("button", { name: "Open in new tab" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Print" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Download" })).toBeDisabled();
  });

  it("toggles fit width", async () => {
    const user = userEvent.setup();
    const onFitWidthChange = vi.fn();
    renderToolbar({ fitWidth: false, onFitWidthChange });
    await user.click(screen.getByRole("button", { name: "Fit width" }));
    expect(onFitWidthChange).toHaveBeenCalledWith(true);
  });

  it("reports annotation tool changes", async () => {
    const user = userEvent.setup();
    const onAnnotationToolChange = vi.fn();
    renderToolbar({ onAnnotationToolChange });

    await user.click(screen.getByRole("button", { name: "Highlight" }));
    expect(onAnnotationToolChange).toHaveBeenCalledWith("highlight");

    await user.click(screen.getByRole("button", { name: "Ink" }));
    expect(onAnnotationToolChange).toHaveBeenCalledWith("ink");

    await user.click(screen.getByRole("button", { name: "Eraser" }));
    expect(onAnnotationToolChange).toHaveBeenCalledWith("eraser");
  });

  it("marks the active annotation tool as pressed", () => {
    renderToolbar({ annotationTool: "ink" });
    expect(screen.getByRole("button", { name: "Ink" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "Select" })).toHaveAttribute("aria-pressed", "false");
  });

  it("reports clear annotations when there are any", async () => {
    const user = userEvent.setup();
    const onClearAnnotations = vi.fn();
    renderToolbar({ hasAnnotations: true, onClearAnnotations });
    const clear = screen.getByRole("button", { name: "Clear all annotations" });
    expect(clear).toBeEnabled();
    await user.click(clear);
    expect(onClearAnnotations).toHaveBeenCalled();
  });

  it("disables clear annotations while there are none", () => {
    renderToolbar({ hasAnnotations: false });
    expect(screen.getByRole("button", { name: "Clear all annotations" })).toBeDisabled();
  });
});

describe("PdfViewer (container)", () => {
  it("renders a loading status then the page with totals", async () => {
    renderWithProviders(<PdfViewer file="ok.pdf" />);
    expect(await screen.findByText("of 5")).toBeInTheDocument();
    expect(await screen.findByTestId("page")).toBeInTheDocument();
  });

  it("surfaces an error state when loading fails", async () => {
    renderWithProviders(<PdfViewer file="bad.pdf" />);
    expect(await screen.findByText("Could not load this PDF.")).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: "Try again" })).toBeInTheDocument();
  });
});

describe("PdfViewer (toolbar wiring)", () => {
  const blobUrl = "blob:fake-pdf";

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function stubFileApis() {
    for (const element of document.querySelectorAll("iframe")) {
      element.remove();
    }
    vi.spyOn(URL, "createObjectURL").mockReturnValue(blobUrl);
    vi.spyOn(URL, "revokeObjectURL").mockImplementation(() => {});
    vi.spyOn(window, "open").mockImplementation(() => null);
    vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(function (
      this: HTMLAnchorElement,
    ) {
      expect(this.getAttribute("download")).toBe("document.pdf");
    });
  }

  async function renderLoaded(file = "ok.pdf") {
    const onLoadSuccess = vi.fn();
    const onError = vi.fn();
    renderWithProviders(<PdfViewer file={file} onLoadSuccess={onLoadSuccess} onError={onError} />);
    await screen.findByTestId("page");
    return { onLoadSuccess, onError };
  }

  it("notifies load success and error callbacks", async () => {
    const loaded = await renderLoaded("ok.pdf");
    expect(loaded.onLoadSuccess).toHaveBeenCalledWith(5);

    const failed = await renderLoaded("bad.pdf");
    expect(failed.onError).toHaveBeenCalledWith(expect.any(Error));
  });

  it("clamps page navigation to the loaded page count", async () => {
    await renderLoaded();
    const input = screen.getByLabelText("Go to page") as HTMLInputElement;
    fireEvent.change(input, { target: { value: "99" } });
    expect(input.value).toBe("5");
    fireEvent.change(input, { target: { value: "0" } });
    expect(input.value).toBe("1");
  });

  it("retries after a load error", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PdfViewer file="bad.pdf" />);
    await user.click(await screen.findByRole("button", { name: "Try again" }));
    expect(await screen.findByText("Could not load this PDF.")).toBeInTheDocument();
  });

  it("zooms, toggles fit width, and rotates from the toolbar", async () => {
    const user = userEvent.setup();
    await renderLoaded();
    const fitWidth = screen.getByRole("button", { name: "Fit width" });
    expect(fitWidth).toHaveAttribute("aria-pressed", "true");

    await user.click(screen.getByRole("button", { name: "Zoom in" }));
    expect(fitWidth).toHaveAttribute("aria-pressed", "false");

    await user.click(fitWidth);
    expect(fitWidth).toHaveAttribute("aria-pressed", "true");

    await user.click(screen.getByRole("button", { name: "Rotate page" }));
    await user.click(screen.getByRole("button", { name: "Rotate page left" }));
    expect(fitWidth).toBeInTheDocument();
  });

  it("opens, prints, and downloads the loaded document", async () => {
    stubFileApis();
    const user = userEvent.setup();
    await renderLoaded();

    await user.click(screen.getByRole("button", { name: "Open in new tab" }));
    expect(window.open).toHaveBeenCalledWith(blobUrl, "_blank", "noopener");

    await user.click(screen.getByRole("button", { name: "Print" }));
    expect(document.querySelector("iframe")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Download" }));
    expect(URL.createObjectURL).toHaveBeenCalled();
  });

  it("cleans up the print iframe after its timeout", async () => {
    stubFileApis();
    await renderLoaded();
    vi.useFakeTimers();
    try {
      fireEvent.click(screen.getByRole("button", { name: "Print" }));
      await vi.advanceTimersByTimeAsync(0);
      const iframe = document.querySelector("iframe") as HTMLIFrameElement;
      expect(iframe).toBeInTheDocument();
      iframe.dispatchEvent(new Event("load"));
      await vi.advanceTimersByTimeAsync(30_000);
      expect(document.querySelector("iframe")).not.toBeInTheDocument();
      expect(URL.revokeObjectURL).toHaveBeenCalledWith(blobUrl);
    } finally {
      vi.useRealTimers();
    }
  });

  it("revokes the download url after its delay", async () => {
    stubFileApis();
    await renderLoaded();
    vi.useFakeTimers();
    try {
      fireEvent.click(screen.getByRole("button", { name: "Download" }));
      await vi.advanceTimersByTimeAsync(1_000);
      expect(URL.revokeObjectURL).toHaveBeenCalledWith(blobUrl);
    } finally {
      vi.useRealTimers();
    }
  });

  it("revokes blob urls when unmounted", async () => {
    stubFileApis();
    const user = userEvent.setup();
    const { unmount } = renderWithProviders(<PdfViewer file="ok.pdf" />);
    await screen.findByTestId("page");
    await user.click(screen.getByRole("button", { name: "Download" }));
    unmount();
    expect(URL.revokeObjectURL).toHaveBeenCalledWith(blobUrl);
  });

  it("unmounts cleanly before the worker Finishes loading", () => {
    const { unmount } = renderWithProviders(<PdfViewer file="ok.pdf" />);
    unmount();
    expect(screen.queryByText("Loading document…")).not.toBeInTheDocument();
  });

  it("draws a highlight annotation and clears it", async () => {
    const user = userEvent.setup();
    await renderLoaded();

    await user.click(screen.getByRole("button", { name: "Highlight" }));
    const surface = document.querySelector("[data-slot='pdf-annotation-surface']");
    expect(surface).toBeInTheDocument();
    const layer = document.querySelector("[data-slot='pdf-annotation-layer']") as HTMLElement;
    vi.spyOn(layer, "getBoundingClientRect").mockReturnValue({
      left: 0,
      top: 0,
      width: 400,
      height: 400,
    } as DOMRect);

    fireEvent.pointerDown(surface as Element, { clientX: 40, clientY: 40, pointerId: 1 });
    fireEvent.pointerMove(surface as Element, { clientX: 120, clientY: 80, pointerId: 1 });
    fireEvent.pointerUp(surface as Element, { pointerId: 1 });

    const clear = screen.getByRole("button", { name: "Clear all annotations" });
    expect(clear).toBeEnabled();
    await user.click(clear);
    expect(clear).toBeDisabled();
  });

  it("deletes a note annotation with the eraser", async () => {
    const user = userEvent.setup();
    await renderLoaded();

    await user.click(screen.getByRole("button", { name: "Note" }));
    const surface = document.querySelector("[data-slot='pdf-annotation-surface']") as Element;
    const layer = document.querySelector("[data-slot='pdf-annotation-layer']") as HTMLElement;
    vi.spyOn(layer, "getBoundingClientRect").mockReturnValue({
      left: 0,
      top: 0,
      width: 400,
      height: 400,
    } as DOMRect);

    fireEvent.pointerDown(surface, { clientX: 40, clientY: 40, pointerId: 1 });
    fireEvent.pointerMove(surface, { clientX: 120, clientY: 80, pointerId: 1 });
    fireEvent.pointerUp(surface, { pointerId: 1 });

    await user.click(screen.getByRole("button", { name: "Eraser" }));
    await user.click(screen.getByRole("button", { name: "Delete annotation" }));
    expect(screen.getByRole("button", { name: "Clear all annotations" })).toBeDisabled();
  });

  it("edits and saves a note annotation", async () => {
    const user = userEvent.setup();
    await renderLoaded();

    await user.click(screen.getByRole("button", { name: "Note" }));
    const surface = document.querySelector("[data-slot='pdf-annotation-surface']") as Element;
    const layer = document.querySelector("[data-slot='pdf-annotation-layer']") as HTMLElement;
    vi.spyOn(layer, "getBoundingClientRect").mockReturnValue({
      left: 0,
      top: 0,
      width: 400,
      height: 400,
    } as DOMRect);

    fireEvent.pointerDown(surface, { clientX: 40, clientY: 40, pointerId: 1 });
    fireEvent.pointerMove(surface, { clientX: 120, clientY: 80, pointerId: 1 });
    fireEvent.pointerUp(surface, { pointerId: 1 });

    const editor = await screen.findByLabelText("Write a note…");
    await user.type(editor, "Review this clause");
    await user.click(screen.getByRole("button", { name: "Save note" }));
    expect(screen.queryByLabelText("Write a note…")).not.toBeInTheDocument();
    expect(document.querySelector("[data-slot='pdf-annotation-rect']")).toBeInTheDocument();
  });

  it("applies custom height, labels, and aria-label", async () => {
    renderWithProviders(
      <PdfViewer file="ok.pdf" height="min-h-[50svh]" aria-label="Contract PDF" />,
    );
    const section = await screen.findByLabelText("Contract PDF");
    expect(section.className).toContain("min-h-[50svh]");
  });
});
