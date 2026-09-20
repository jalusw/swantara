import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { PdfViewerToolbar } from "@/components/pdf-viewer-toolbar";
import { renderWithProviders } from "@/lib/tests";

const defaultProps = {
  page: 2,
  pages: 5,
  scale: 1,
  fitWidth: false,
  canFileActions: true,
  annotationTool: "select" as const,
  hasAnnotations: false,
  onOpen: vi.fn(),
  onPrint: vi.fn(),
  onDownload: vi.fn(),
  onPageChange: vi.fn(),
  onZoomChange: vi.fn(),
  onFitWidthChange: vi.fn(),
  onRotateChange: vi.fn(),
  onAnnotationToolChange: vi.fn(),
  onClearAnnotations: vi.fn(),
};

describe("PdfViewerToolbar", () => {
  it("renders with data-slot", () => {
    renderWithProviders(<PdfViewerToolbar {...defaultProps} />);

    expect(document.querySelector("[data-slot='pdf-viewer-toolbar']")).toBeInTheDocument();
  });

  it("disables navigation at page bounds", () => {
    renderWithProviders(<PdfViewerToolbar {...defaultProps} page={1} pages={1} />);

    expect(screen.getByRole("button", { name: "Previous page" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Next page" })).toBeDisabled();
  });

  it("disables file actions when canFileActions is false", () => {
    renderWithProviders(<PdfViewerToolbar {...defaultProps} canFileActions={false} />);

    expect(screen.getByRole("button", { name: "Open in new tab" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Print" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Download" })).toBeDisabled();
  });
});
