"use client";

import type { LucideIcon } from "lucide-react";
import {
  ChevronLeftIcon,
  ChevronRightIcon,
  ChevronsLeftIcon,
  ChevronsRightIcon,
  DownloadIcon,
  EraserIcon,
  ExternalLinkIcon,
  HighlighterIcon,
  MaximizeIcon,
  MousePointerIcon,
  PenIcon,
  PrinterIcon,
  RotateCcwIcon,
  RotateCwIcon,
  StickyNoteIcon,
  Trash2Icon,
  UnderlineIcon,
  ZoomInIcon,
  ZoomOutIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "./button";
import { Input } from "./input";
import type { PdfAnnotationTool } from "./pdf-annotation";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./select";

const ZOOM_PRESETS = [50, 75, 100, 125, 150, 200];
const MIN_ZOOM = ZOOM_PRESETS[0]!;
const MAX_ZOOM = ZOOM_PRESETS[ZOOM_PRESETS.length - 1]!;

const ANNOTATION_TOOLS: Array<{
  value: PdfAnnotationTool;
  icon: LucideIcon;
}> = [
  { value: "select", icon: MousePointerIcon },
  { value: "highlight", icon: HighlighterIcon },
  { value: "underline", icon: UnderlineIcon },
  { value: "note", icon: StickyNoteIcon },
  { value: "ink", icon: PenIcon },
  { value: "eraser", icon: EraserIcon },
];

export type PdfViewerToolbarLabels = {
  open?: string;
  print?: string;
  download?: string;
  firstPage?: string;
  previousPage?: string;
  nextPage?: string;
  lastPage?: string;
  goToPage?: string;
  of?: string;
  annotations?: string;
  toolSelect?: string;
  toolHighlight?: string;
  toolUnderline?: string;
  toolNote?: string;
  toolInk?: string;
  toolEraser?: string;
  clearAnnotations?: string;
  zoomOut?: string;
  zoomIn?: string;
  zoom?: string;
  fitWidth?: string;
  rotateLeft?: string;
  rotate?: string;
};

export type PdfViewerToolbarProps = {
  page: number;
  pages: number;
  scale: number;
  fitWidth: boolean;
  canFileActions: boolean;
  annotationTool: PdfAnnotationTool;
  hasAnnotations: boolean;
  disabled?: boolean;
  className?: string;
  labels?: PdfViewerToolbarLabels;
  onOpen: () => void;
  onPrint: () => void;
  onDownload: () => void;
  onPageChange: (page: number) => void;
  onZoomChange: (scale: number) => void;
  onFitWidthChange: (fitWidth: boolean) => void;
  onRotateChange: (direction: -90 | 90) => void;
  onAnnotationToolChange: (tool: PdfAnnotationTool) => void;
  onClearAnnotations: () => void;
};

function useDefaultLabels(): Required<PdfViewerToolbarLabels> {
  return {
    open: "Open in new tab",
    print: "Print",
    download: "Download",
    firstPage: "First page",
    previousPage: "Previous page",
    nextPage: "Next page",
    lastPage: "Last page",
    goToPage: "Go to page",
    of: "of",
    annotations: "Annotations",
    toolSelect: "Select",
    toolHighlight: "Highlight",
    toolUnderline: "Underline",
    toolNote: "Note",
    toolInk: "Ink",
    toolEraser: "Eraser",
    clearAnnotations: "Clear all annotations",
    zoomOut: "Zoom out",
    zoomIn: "Zoom in",
    zoom: "Zoom",
    fitWidth: "Fit width",
    rotateLeft: "Rotate page left",
    rotate: "Rotate page",
  };
}

const toolLabelMap: Record<PdfAnnotationTool, keyof PdfViewerToolbarLabels> = {
  select: "toolSelect",
  highlight: "toolHighlight",
  underline: "toolUnderline",
  note: "toolNote",
  ink: "toolInk",
  eraser: "toolEraser",
};

export function PdfViewerToolbar({
  page,
  pages,
  scale,
  fitWidth,
  canFileActions,
  annotationTool,
  hasAnnotations,
  disabled = false,
  className,
  labels,
  onOpen,
  onPrint,
  onDownload,
  onPageChange,
  onZoomChange,
  onFitWidthChange,
  onRotateChange,
  onAnnotationToolChange,
  onClearAnnotations,
}: PdfViewerToolbarProps) {
  const translatedDefaults = useDefaultLabels();
  const t = { ...translatedDefaults, ...labels };
  const zoomPercent = Math.round(scale * 100);

  const atFirst = disabled || page <= 1;
  const atLast = disabled || page >= pages;

  function stepZoom(direction: -1 | 1) {
    const current = Math.round(scale * 100);
    const index = ZOOM_PRESETS.indexOf(current);
    const base = index === -1 ? ZOOM_PRESETS.indexOf(100) : index;
    const next = Math.min(Math.max(base + direction, 0), ZOOM_PRESETS.length - 1);
    onZoomChange(ZOOM_PRESETS[next]! / 100);
  }

  return (
    <div
      data-slot="pdf-viewer-toolbar"
      className={cn(
        "flex flex-wrap items-center justify-between gap-2 border-b border-border-subtle px-3 py-2",
        className,
      )}
    >
      <div className="flex items-center gap-1">
        <div className="flex items-center gap-1 pr-1">
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t.open}
            disabled={!canFileActions || disabled}
            onClick={onOpen}
          >
            <ExternalLinkIcon aria-hidden />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t.print}
            disabled={!canFileActions || disabled}
            onClick={onPrint}
          >
            <PrinterIcon aria-hidden />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t.download}
            disabled={!canFileActions || disabled}
            onClick={onDownload}
          >
            <DownloadIcon aria-hidden />
          </Button>
          <span className="mx-1 h-5 w-px bg-border-subtle" aria-hidden />
        </div>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label={t.firstPage}
          disabled={atFirst}
          onClick={() => onPageChange(1)}
        >
          <ChevronsLeftIcon aria-hidden />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label={t.previousPage}
          disabled={atFirst}
          onClick={() => onPageChange(page - 1)}
        >
          <ChevronLeftIcon aria-hidden />
        </Button>
        <div className="flex items-center gap-1 px-1 text-sm">
          <label htmlFor="pdf-viewer-page" className="sr-only">
            {t.goToPage}
          </label>
          <Input
            id="pdf-viewer-page"
            type="number"
            min={1}
            max={Math.max(pages, 1)}
            value={page}
            disabled={disabled}
            onChange={(event) => {
              const next = Number.parseInt(event.target.value, 10);
              if (Number.isFinite(next)) {
                onPageChange(next);
              }
            }}
            className="w-14 min-h-11 px-1.5 text-center text-sm"
          />
          <span className="text-muted-foreground">
            {t.of} {pages || "—"}
          </span>
        </div>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label={t.nextPage}
          disabled={atLast}
          onClick={() => onPageChange(page + 1)}
        >
          <ChevronRightIcon aria-hidden />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label={t.lastPage}
          disabled={atLast}
          onClick={() => onPageChange(pages)}
        >
          <ChevronsRightIcon aria-hidden />
        </Button>
      </div>

      <div className="flex items-center gap-1">
        <fieldset className="flex items-center gap-0.5">
          <legend className="sr-only">{t.annotations}</legend>
          {ANNOTATION_TOOLS.map(({ value, icon: Icon }) => (
            <Button
              key={value}
              variant={annotationTool === value ? "secondary" : "ghost"}
              size="icon-sm"
              aria-pressed={annotationTool === value}
              aria-label={t[toolLabelMap[value]]}
              disabled={disabled}
              onClick={() => onAnnotationToolChange(value)}
            >
              <Icon aria-hidden />
            </Button>
          ))}
        </fieldset>
        <span className="mx-1 h-5 w-px bg-border-subtle" aria-hidden />
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label={t.clearAnnotations}
          disabled={disabled || !hasAnnotations}
          onClick={onClearAnnotations}
        >
          <Trash2Icon aria-hidden />
        </Button>
      </div>

      <div className="flex items-center gap-1">
        <div className="flex items-center gap-0.5">
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t.zoomOut}
            disabled={disabled || zoomPercent <= MIN_ZOOM}
            onClick={() => stepZoom(-1)}
          >
            <ZoomOutIcon aria-hidden />
          </Button>
          <div className="flex items-center gap-0.5">
            <Select
              value={fitWidth ? "fit" : String(zoomPercent)}
              disabled={disabled}
              onValueChange={(value) => {
                if (value === null) {
                  return;
                }
                if (value === "fit") {
                  onFitWidthChange(true);
                } else {
                  onFitWidthChange(false);
                  onZoomChange(Number.parseInt(value, 10) / 100);
                }
              }}
            >
              <SelectTrigger size="sm" aria-label={t.zoom}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="fit">{t.fitWidth}</SelectItem>
                {ZOOM_PRESETS.map((preset) => (
                  <SelectItem key={preset} value={String(preset)}>
                    {preset}%
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={t.zoomIn}
            disabled={disabled || zoomPercent >= MAX_ZOOM}
            onClick={() => stepZoom(1)}
          >
            <ZoomInIcon aria-hidden />
          </Button>
        </div>

        <span className="mx-1 h-5 w-px bg-border-subtle" aria-hidden />

        <Button
          variant={fitWidth ? "secondary" : "ghost"}
          size="sm"
          aria-pressed={fitWidth}
          aria-label={t.fitWidth}
          disabled={disabled}
          onClick={() => onFitWidthChange(!fitWidth)}
        >
          <MaximizeIcon aria-hidden />
          <span className="hidden sm:inline">{t.fitWidth}</span>
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label={t.rotateLeft}
          disabled={disabled}
          onClick={() => onRotateChange(-90)}
        >
          <RotateCcwIcon aria-hidden />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label={t.rotate}
          disabled={disabled}
          onClick={() => onRotateChange(90)}
        >
          <RotateCwIcon aria-hidden />
        </Button>
      </div>
    </div>
  );
}
