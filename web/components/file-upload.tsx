"use client";

import { UploadCloudIcon, XIcon } from "lucide-react";
import { useState } from "react";
import { cn, formatBytes } from "@/lib/utils";
import { Button } from "./button";

let idCounter = 0;

function makeId(): string {
  idCounter += 1;
  return `file-${Date.now()}-${idCounter}`;
}

export type UploadStatus = "idle" | "uploading" | "done" | "error";

export type UploadedFile = {
  id: string;
  file: File;
  status: UploadStatus;
  progress: number;
  error?: string;
};

export type FileUploadOptions = {
  maxSize?: number;
  accept?: string[];
};

function makeEntry(file: File, options: FileUploadOptions): UploadedFile {
  const entry: UploadedFile = {
    id: makeId(),
    file,
    status: "idle",
    progress: 0,
  };
  if (options.maxSize != null && file.size > options.maxSize) {
    entry.status = "error";
    entry.error = `File exceeds ${formatBytes(options.maxSize)}`;
  } else if (options.accept && options.accept.length > 0) {
    const matches = options.accept.some((pattern) =>
      pattern.startsWith(".")
        ? file.name.toLowerCase().endsWith(pattern.toLowerCase())
        : file.type === pattern,
    );
    if (!matches) {
      entry.status = "error";
      entry.error = `Unsupported type. Accepted: ${options.accept.join(", ")}`;
    }
  }
  return entry;
}

export type UseUploadResult = {
  files: UploadedFile[];
  addFiles: (input: FileList | File[] | null) => Promise<UploadedFile[]>;
  removeFile: (id: string) => void;
};

export function useFileUpload(
  options: FileUploadOptions = {},
  onUpload: (
    file: UploadedFile,
    onProgress: (progress: number) => void,
  ) => void | Promise<void> = () => {},
): UseUploadResult {
  const [files, setFiles] = useState<UploadedFile[]>([]);

  function updateFile(id: string, patch: Omit<Partial<UploadedFile>, "id">) {
    setFiles((previous) =>
      previous.map((entry) => (entry.id === id ? { ...entry, ...patch } : entry)),
    );
  }

  async function addFiles(input: FileList | File[] | null) {
    if (!input) {
      return [];
    }
    const entries = Array.from(input).map((file) => makeEntry(file, options));
    setFiles((previous) => [...previous, ...entries]);

    if (onUpload) {
      for (const entry of entries) {
        if (entry.status !== "idle") {
          continue;
        }
        updateFile(entry.id, { status: "uploading" });
        try {
          await onUpload(entry, (progress) => updateFile(entry.id, { progress }));
          updateFile(entry.id, { status: "done", progress: 100 });
        } catch {
          updateFile(entry.id, { status: "error", error: "Upload failed." });
        }
      }
    }
    return entries;
  }

  function removeFile(id: string) {
    setFiles((previous) => previous.filter((entry) => entry.id !== id));
  }

  return { files, addFiles, removeFile };
}

export type FileUploadProps = FileUploadOptions & {
  label?: string;
  description?: string;
  multiple?: boolean;
  onFilesChange?: (files: UploadedFile[]) => void;
  className?: string;
};

function StatusText({ entry }: { entry: UploadedFile }) {
  if (entry.status === "error") {
    return (
      <span className="text-xs text-destructive" role="alert" aria-live="assertive">
        {entry.error}
      </span>
    );
  }
  const detail =
    entry.status === "uploading"
      ? ` · ${Math.round(entry.progress)}%`
      : entry.status === "done"
        ? " · uploaded"
        : "";
  return (
    <span className="text-xs text-muted-foreground tabular-nums" aria-live="polite">
      {formatBytes(entry.file.size)}
      {detail}
    </span>
  );
}

export function FileUpload({
  label = "Attach files",
  description = "Drag and drop or click to choose files.",
  multiple = false,
  maxSize,
  accept,
  onFilesChange,
  className,
}: FileUploadProps) {
  const [dragging, setDragging] = useState(false);
  const upload = useFileUpload({ maxSize, accept });
  const { files } = upload;

  function handleInput(input: FileList | File[] | null) {
    void upload.addFiles(input);
    onFilesChange?.(files);
  }

  return (
    <div data-slot="file-upload" className={cn("space-y-2", className)}>
      {/* biome-ignore lint/a11y/noNoninteractiveElementInteractions: dropzone; the wrapped file input provides keyboard access */}
      <label
        data-dragging={dragging || undefined}
        onDragOver={(event) => {
          event.preventDefault();
          setDragging(true);
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={(event) => {
          event.preventDefault();
          setDragging(false);
          handleInput(event.dataTransfer.files);
        }}
        className={cn(
          "flex cursor-pointer flex-col items-center justify-center gap-1 rounded-lg border border-dashed border-border px-4 py-8 text-center text-sm transition-colors hover:bg-muted/40 focus-within:ring-2 ring-offset-2 ring-offset-background focus-within:ring-ring",
          dragging && "border-primary bg-primary/5",
        )}
      >
        <UploadCloudIcon className="size-6 text-muted-foreground" aria-hidden />
        <span className="">{label}</span>
        <span className="text-xs text-muted-foreground">{description}</span>
        <input
          type="file"
          multiple={multiple}
          accept={accept?.join(",")}
          className="sr-only"
          onChange={(event) => handleInput(event.target.files)}
        />
      </label>
      {files.length > 0 ? (
        <ul data-slot="file-upload-list" className="flex flex-col gap-1.5">
          {files.map((entry) => (
            <li
              key={entry.id}
              className="flex items-center gap-2 rounded-md border border-border px-2.5 py-2 text-sm"
            >
              <div className="min-w-0 flex-1 space-y-0.5">
                <p className="truncate ">{entry.file.name}</p>
                <StatusText entry={entry} />
              </div>
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                aria-label={`Remove ${entry.file.name}`}
                onClick={() => upload.removeFile(entry.id)}
              >
                <XIcon aria-hidden />
              </Button>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}
