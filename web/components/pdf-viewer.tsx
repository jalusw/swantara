"use client";

import "@/lib/utils";

import type { PDFDocumentProxy } from "pdfjs-dist";
import type { RefObject } from "react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useResizeObserver } from "usehooks-ts";
import { cn } from "@/lib/utils";
import type { PdfAnnotation, PdfAnnotationTool } from "./pdf-annotation";
import { PdfAnnotationLayer } from "./pdf-annotation-layer";
import { PdfViewerStatus } from "./pdf-viewer-status";
import { PdfViewerToolbar, type PdfViewerToolbarLabels } from "./pdf-viewer-toolbar";

const PDF_URL_REVOKE_DELAY = 1_000;

type PdfComponents = {
  Document: typeof import("react-pdf").Document;
  Page: typeof import("react-pdf").Page;
  pdfjs: typeof import("react-pdf").pdfjs;
};

export type PdfSource = string | File | Blob | ArrayBuffer | { url: string } | { data: Uint8Array };

export type PdfViewerLabels = PdfViewerToolbarLabels & {
  loading?: string;
  errorMessage?: string;
  retry?: string;
  emptyTitle?: string;
  emptyDescription?: string;
  toolNote?: string;
  toolInk?: string;
  deleteAnnotation?: string;
  notePlaceholder?: string;
  cancelNote?: string;
  saveNote?: string;
};

export type PdfViewerProps = {
  file: PdfSource;
  initialPage?: number;
  initialScale?: number;
  fitWidth?: boolean;
  height?: string;
  className?: string;
  "aria-label"?: string;
  labels?: PdfViewerLabels;
  onLoadSuccess?: (numPages: number) => void;
  onError?: (error: Error) => void;
};

const DEFAULT_HEIGHT = "min-h-[70svh]";

export function PdfViewer({
  file,
  initialPage = 1,
  initialScale = 1,
  fitWidth: initialFitWidth = true,
  height = DEFAULT_HEIGHT,
  className,
  "aria-label": ariaLabel = "PDF viewer",
  labels,
  onLoadSuccess,
  onError,
}: PdfViewerProps) {
  const [pdfComponents, setPdfComponents] = useState<PdfComponents | null>(null);
  const [numPages, setNumPages] = useState(0);
  const [pageNumber, setPageNumber] = useState(initialPage);
  const [scale, setScale] = useState(initialScale);
  const [fitWidth, setFitWidth] = useState(initialFitWidth);
  const [rotate, setRotate] = useState(0);
  const [loadError, setLoadError] = useState<Error | null>(null);
  const [attempt, setAttempt] = useState(0);
  const [annotationTool, setAnnotationTool] = useState<PdfAnnotationTool>("select");
  const [annotations, setAnnotations] = useState<PdfAnnotation[]>([]);

  const scrollRef = useRef<HTMLDivElement>(null);
  const pdfRef = useRef<PDFDocumentProxy | null>(null);
  const blobUrlsRef = useRef<Set<string>>(new Set());
  const { width: scrollWidth = 0 } = useResizeObserver({
    ref: scrollRef as RefObject<HTMLElement>,
  });

  const memoizedFile = useMemo(() => file, [file]);

  useEffect(() => {
    return () => {
      for (const url of blobUrlsRef.current) {
        URL.revokeObjectURL(url);
      }
      blobUrlsRef.current.clear();
    };
  }, []);

  useEffect(() => {
    let cancelled = false;
    import("react-pdf").then((mod) => {
      if (cancelled) return;
      mod.pdfjs.GlobalWorkerOptions.workerSrc = new URL(
        "pdfjs-dist/build/pdf.worker.min.mjs",
        import.meta.url,
      ).toString();
      setPdfComponents({
        Document: mod.Document,
        Page: mod.Page,
        pdfjs: mod.pdfjs,
      });
    });
    return () => {
      cancelled = true;
    };
  }, []);

  const clampedPageNumber = numPages > 0 && pageNumber > numPages ? numPages : pageNumber;

  function goToPage(next: number) {
    const clamped = Math.min(Math.max(next, 1), numPages || 1);
    setPageNumber(clamped);
  }

  function retry() {
    setLoadError(null);
    setPageNumber(initialPage);
    setAttempt((value) => value + 1);
  }

  async function getPendingBlobUrl(): Promise<string | null> {
    const pdf = pdfRef.current;
    if (!pdf) {
      return null;
    }
    const data = await pdf.getData();
    const blob = new Blob([data as Uint8Array<ArrayBuffer>], {
      type: "application/pdf",
    });
    const url = URL.createObjectURL(blob);
    blobUrlsRef.current.add(url);
    return url;
  }

  async function openDocument() {
    const url = await getPendingBlobUrl();
    if (url) {
      window.open(url, "_blank", "noopener");
      URL.revokeObjectURL(url);
      blobUrlsRef.current.delete(url);
    }
  }

  async function printDocument() {
    const url = await getPendingBlobUrl();
    if (!url) {
      return;
    }
    const iframe = document.createElement("iframe");
    iframe.style.position = "fixed";
    iframe.style.left = "-1000px";
    iframe.style.border = "none";
    iframe.style.width = "0";
    iframe.style.height = "0";
    document.body.appendChild(iframe);
    iframe.onload = () => {
      iframe.contentWindow?.focus();
      iframe.contentWindow?.print();
    };
    iframe.src = url;
    window.setTimeout(() => {
      iframe.remove();
      URL.revokeObjectURL(url);
      blobUrlsRef.current.delete(url);
    }, 30_000);
  }

  async function downloadDocument() {
    const url = await getPendingBlobUrl();
    if (!url) {
      return;
    }
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = "document.pdf";
    anchor.click();
    window.setTimeout(() => {
      URL.revokeObjectURL(url);
      blobUrlsRef.current.delete(url);
    }, PDF_URL_REVOKE_DELAY);
  }

  const pageWidth = fitWidth && scrollWidth > 0 ? scrollWidth : undefined;
  const canFileActions = numPages > 0 && loadError === null;

  function addAnnotation(annotation: PdfAnnotation) {
    setAnnotations((current) => [...current, annotation]);
  }

  function deleteAnnotation(id: string) {
    setAnnotations((current) => current.filter((item) => item.id !== id));
  }

  function updateNote(id: string, note: string) {
    setAnnotations((current) =>
      current.map((item) => (item.id === id && item.tool === "note" ? { ...item, note } : item)),
    );
  }

  if (!pdfComponents) {
    return (
      <div
        data-slot="pdf-viewer"
        className={cn("overflow-hidden rounded-xl bg-background border border-border", className)}
      >
        <PdfViewerToolbar
          page={clampedPageNumber}
          pages={numPages}
          scale={scale}
          fitWidth={fitWidth}
          canFileActions={false}
          annotationTool={annotationTool}
          hasAnnotations={annotations.length > 0}
          disabled
          labels={labels}
          onOpen={openDocument}
          onPrint={printDocument}
          onDownload={downloadDocument}
          onPageChange={goToPage}
          onZoomChange={(next) => {
            setScale(next);
            setFitWidth(false);
          }}
          onFitWidthChange={setFitWidth}
          onRotateChange={(direction) => setRotate((value) => (value + direction + 360) % 360)}
          onAnnotationToolChange={setAnnotationTool}
          onClearAnnotations={() => setAnnotations([])}
        />
        <section
          aria-label={ariaLabel}
          className={cn("overflow-auto bg-muted/40 p-4 [scrollbar-width:thin]", height)}
        >
          <PdfViewerStatus type="loading" labels={labels} />
        </section>
      </div>
    );
  }

  const { Document, Page } = pdfComponents;

  return (
    <div
      data-slot="pdf-viewer"
      className={cn("overflow-hidden rounded-xl bg-background border border-border", className)}
    >
      <PdfViewerToolbar
        page={clampedPageNumber}
        pages={numPages}
        scale={scale}
        fitWidth={fitWidth}
        canFileActions={canFileActions}
        annotationTool={annotationTool}
        hasAnnotations={annotations.length > 0}
        disabled={loadError !== null}
        labels={labels}
        onOpen={openDocument}
        onPrint={printDocument}
        onDownload={downloadDocument}
        onPageChange={goToPage}
        onZoomChange={(next) => {
          setScale(next);
          setFitWidth(false);
        }}
        onFitWidthChange={setFitWidth}
        onRotateChange={(direction) => setRotate((value) => (value + direction + 360) % 360)}
        onAnnotationToolChange={setAnnotationTool}
        onClearAnnotations={() => setAnnotations([])}
      />

      <section
        ref={scrollRef}
        aria-label={ariaLabel}
        className={cn("overflow-auto bg-muted/40 p-4 [scrollbar-width:thin]", height)}
      >
        <div className="flex min-h-full items-start justify-center">
          <Document
            key={attempt}
            file={memoizedFile}
            loading={<PdfViewerStatus type="loading" labels={labels} />}
            error={<PdfViewerStatus type="error" onRetry={retry} labels={labels} />}
            noData={<PdfViewerStatus type="empty" labels={labels} />}
            onLoadSuccess={(loadedPdf: PDFDocumentProxy) => {
              pdfRef.current = loadedPdf;
              setNumPages(loadedPdf.numPages ?? 0);
              setLoadError(null);
              if (loadedPdf) {
                onLoadSuccess?.(loadedPdf.numPages);
              }
            }}
            onLoadError={(error: Error) => {
              setLoadError(error);
              onError?.(error);
            }}
          >
            <div className="relative bg-card shadow-md border border-border">
              <Page
                pageNumber={clampedPageNumber}
                scale={pageWidth === undefined ? scale : undefined}
                width={pageWidth}
                rotate={rotate}
                renderTextLayer={false}
                renderAnnotationLayer={false}
                loading={<PdfViewerStatus type="loading" labels={labels} />}
                error={<PdfViewerStatus type="error" labels={labels} />}
              />
              <PdfAnnotationLayer
                tool={annotationTool}
                annotations={annotations}
                disabled={loadError !== null}
                labels={labels}
                onAdd={addAnnotation}
                onDelete={deleteAnnotation}
                onUpdateNote={updateNote}
              />
            </div>
          </Document>
        </div>
      </section>
    </div>
  );
}
