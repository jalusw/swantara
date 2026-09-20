"use client";

import {
  DownloadIcon,
  FileArchiveIcon,
  FileCodeIcon,
  FileIcon,
  FileImageIcon,
  FileSpreadsheetIcon,
  FileTextIcon,
  Trash2Icon,
} from "lucide-react";
import { cn, formatBytes } from "@/lib/utils";
import { Button } from "./button";

export type AttachmentItem = {
  id: string;
  name: string;
  size: number;
  downloadUrl?: string;
};

export type AttachmentListProps = {
  attachments: AttachmentItem[];
  onRemove?: (id: string) => void;
  className?: string;
};

function fileExtension(name: string): string {
  const parts = name.toLowerCase().split(".");
  return parts.length > 1 ? (parts.at(-1) ?? "") : "";
}

function FileTypeIcon({ name }: { name: string }) {
  const ext = fileExtension(name);
  if (["png", "jpg", "jpeg", "gif", "webp", "svg"].includes(ext)) {
    return <FileImageIcon className="size-4" aria-hidden />;
  }
  if (["csv", "xls", "xlsx"].includes(ext)) {
    return <FileSpreadsheetIcon className="size-4" aria-hidden />;
  }
  if (["pdf"].includes(ext)) {
    return <FileTextIcon className="size-4" aria-hidden />;
  }
  if (["zip", "gz", "rar"].includes(ext)) {
    return <FileArchiveIcon className="size-4" aria-hidden />;
  }
  if (["ts", "tsx", "js", "json", "xml", "html"].includes(ext)) {
    return <FileCodeIcon className="size-4" aria-hidden />;
  }
  return <FileIcon className="size-4" aria-hidden />;
}

export function AttachmentList({ attachments, onRemove, className }: AttachmentListProps) {
  return (
    <ul data-slot="attachment-list" className={cn("flex flex-col gap-1.5", className)}>
      {attachments.map((attachment) => (
        <li
          key={attachment.id}
          className="flex items-center gap-2 rounded-md border border-border px-2.5 py-2 text-sm"
        >
          <span
            data-slot="attachment-type-icon"
            className="grid size-8 shrink-0 place-items-center rounded-md bg-muted text-muted-foreground"
          >
            <FileTypeIcon name={attachment.name} />
          </span>
          <div className="min-w-0 flex-1">
            <p className="truncate ">{attachment.name}</p>
            <p className="text-xs text-muted-foreground tabular-nums">
              {formatBytes(attachment.size)}
            </p>
          </div>
          {attachment.downloadUrl ? (
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label={`Download ${attachment.name}`}
              asChild
            >
              <a href={attachment.downloadUrl} download>
                <DownloadIcon aria-hidden />
              </a>
            </Button>
          ) : null}
          {onRemove ? (
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label={`Remove ${attachment.name}`}
              onClick={() => onRemove(attachment.id)}
            >
              <Trash2Icon aria-hidden />
            </Button>
          ) : null}
        </li>
      ))}
    </ul>
  );
}
