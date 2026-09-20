"use client";

import { XIcon } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { Input } from "./input";

export type TagInputProps = {
  value: string[];
  onChange: (tags: string[]) => void;
  placeholder?: string;
  max?: number;
  disabled?: boolean;
  className?: string;
  "aria-label"?: string;
};

export function TagInput({
  value,
  onChange,
  placeholder = "Add tag and press Enter",
  max,
  disabled = false,
  className,
  "aria-label": ariaLabel = "Tags",
}: TagInputProps) {
  const [draft, setDraft] = useState("");
  const [announcement, setAnnouncement] = useState("");

  const canAdd = !max || value.length < max;

  function commit() {
    const tag = draft.trim().replace(/,$/, "");
    if (!tag) {
      return;
    }
    if (value.includes(tag) || !canAdd) {
      setDraft("");
      if (!canAdd) {
        setAnnouncement(`Cannot add more tags. Maximum ${max} allowed.`);
      }
      return;
    }
    onChange([...value, tag]);
    setAnnouncement(`Added tag ${tag}`);
    setDraft("");
  }

  return (
    <div
      data-slot="tag-input"
      className={cn(
        "flex min-h-12 flex-wrap items-center gap-1.5 rounded-md border border-input bg-transparent px-2 py-1.5 text-sm transition-colors focus-within:border-ring focus-within:ring-3 ring-offset-2 ring-offset-background focus-within:ring-ring",
        disabled && "cursor-not-allowed bg-input/50 opacity-50",
        className,
      )}
    >
      <output className="sr-only" aria-live="polite">
        {announcement}
      </output>
      <ul className="flex flex-wrap items-center gap-1.5" aria-label={ariaLabel}>
        {value.map((tag) => (
          <li key={tag} className="flex items-center">
            <span className="flex h-6 items-center gap-1 rounded-sm bg-muted pl-2 text-xs">
              {tag}
              <button
                type="button"
                aria-label={`Remove ${tag}`}
                disabled={disabled}
                onClick={() => {
                  onChange(value.filter((t) => t !== tag));
                  setAnnouncement(`Removed tag ${tag}`);
                }}
                className="relative grid size-6 place-items-center rounded-sm opacity-60 outline-none transition-opacity after:absolute after:-inset-3 after:content-[''] hover:opacity-100 focus-visible:ring-2 ring-offset-2 ring-offset-background focus-visible:ring-ring disabled:pointer-events-none"
              >
                <XIcon className="size-3" aria-hidden />
              </button>
            </span>
          </li>
        ))}
      </ul>
      {canAdd ? (
        <Input
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Enter" || event.key === ",") {
              event.preventDefault();
              commit();
            } else if (event.key === "Backspace" && !draft && value.length > 0) {
              const removed = value[value.length - 1];
              onChange(value.slice(0, -1));
              if (removed) {
                setAnnouncement(`Removed tag ${removed}`);
              }
            }
          }}
          onBlur={commit}
          placeholder={placeholder}
          disabled={disabled}
          aria-label={ariaLabel}
          className="h-7 min-h-11 w-32 flex-1 rounded-none border-none p-0 shadow-none focus-visible:ring-0"
        />
      ) : null}
    </div>
  );
}
