import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { PdfViewerStatus } from "@/components/pdf-viewer-status";
import { renderWithProviders } from "@/lib/tests";

describe("PdfViewerStatus", () => {
  it("renders loading state", () => {
    renderWithProviders(<PdfViewerStatus type="loading" />);

    expect(screen.getByText("Loading document…")).toBeInTheDocument();
    expect(screen.getByRole("status")).toBeInTheDocument();
  });

  it("renders error state with retry button", () => {
    renderWithProviders(<PdfViewerStatus type="error" onRetry={() => {}} />);

    expect(screen.getByText("Could not load this PDF.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Try again" })).toBeInTheDocument();
  });

  it("renders empty state", () => {
    renderWithProviders(<PdfViewerStatus type="empty" />);

    expect(screen.getByText("No preview available")).toBeInTheDocument();
    expect(screen.getByText("Provide a PDF file to preview it here.")).toBeInTheDocument();
  });
});
