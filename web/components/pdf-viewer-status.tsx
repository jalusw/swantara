"use client";

import { FileTextIcon, LoaderCircleIcon, TriangleAlertIcon } from "lucide-react";

import { Button } from "./button";
import { EmptyState } from "./empty-state";

export type PdfViewerStatusProps = {
  type: "loading" | "error" | "empty";
  onRetry?: () => void;
  labels?: {
    loading?: string;
    errorMessage?: string;
    retry?: string;
    emptyTitle?: string;
    emptyDescription?: string;
  };
};

function useDefaultLabels(): Required<NonNullable<PdfViewerStatusProps["labels"]>> {
  return {
    loading: "Loading document…",
    errorMessage: "Could not load this PDF.",
    retry: "Try again",
    emptyTitle: "No preview available",
    emptyDescription: "Provide a PDF file to preview it here.",
  };
}

export function PdfViewerStatus({ type, onRetry, labels }: PdfViewerStatusProps) {
  const translatedDefaults = useDefaultLabels();
  const t = { ...translatedDefaults, ...labels };

  if (type === "loading") {
    return (
      <output
        data-slot="pdf-viewer-status"
        aria-live="polite"
        className="flex h-full min-h-64 w-full flex-col items-center justify-center gap-3"
      >
        <LoaderCircleIcon className="size-6 animate-spin text-primary" />
        <span className="text-sm text-muted-foreground">{t.loading}</span>
      </output>
    );
  }

  if (type === "error") {
    return (
      <div
        data-slot="pdf-viewer-status"
        className="flex h-full min-h-64 w-full flex-col items-center justify-center gap-3 px-6 py-12 text-center"
      >
        <span className="grid size-12 place-items-center rounded-xl bg-destructive/10 text-destructive ring-1 ring-destructive/20">
          <TriangleAlertIcon className="size-6" aria-hidden />
        </span>
        <p className="max-w-sm text-sm text-muted-foreground text-pretty">{t.errorMessage}</p>
        {onRetry ? (
          <Button variant="outline" size="sm" onClick={onRetry}>
            {t.retry}
          </Button>
        ) : null}
      </div>
    );
  }

  return <EmptyState icon={FileTextIcon} title={t.emptyTitle} description={t.emptyDescription} />;
}
