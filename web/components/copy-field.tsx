"use client";

import { CheckIcon, CopyIcon, EyeIcon, EyeOffIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { cn } from "@/lib/utils";
import { Button } from "./button";

const COPIED_TIMEOUT = 1500;

export type CopyFieldProps = {
  value: string;
  defaultValue?: string;
  masked?: boolean;
  showToggle?: boolean;
  onCopy?: (value: string) => void;
  className?: string;
  "aria-label"?: string;
};

export function CopyField({
  value,
  defaultValue = "—",
  masked = false,
  showToggle = true,
  onCopy,
  className,
  "aria-label": ariaLabel = "Copy to clipboard",
}: CopyFieldProps) {
  const [revealed, setRevealed] = useState(false);
  const [copied, setCopied] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (timer.current) {
        clearTimeout(timer.current);
      }
    };
  }, []);

  const displayValue = masked && !revealed ? "••••••••" : value || defaultValue;

  function handleCopy() {
    void navigator.clipboard?.writeText(value).catch(() => {});
    onCopy?.(value);
    setCopied(true);
    if (timer.current) {
      clearTimeout(timer.current);
    }
    timer.current = setTimeout(() => setCopied(false), COPIED_TIMEOUT);
  }

  return (
    <div
      data-slot="copy-field"
      className={cn(
        "flex min-h-11 h-11 items-center gap-1 rounded-md border border-input bg-transparent pr-1 pl-2.5",
        className,
      )}
    >
      <code className="min-w-0 flex-1 truncate text-sm tabular-nums">{displayValue}</code>
      {masked && showToggle ? (
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          aria-label={revealed ? "Hide value" : "Reveal value"}
          onClick={() => setRevealed((v) => !v)}
        >
          {revealed ? <EyeOffIcon /> : <EyeIcon />}
        </Button>
      ) : null}
      <Button
        type="button"
        variant="ghost"
        size="icon-sm"
        aria-label={ariaLabel}
        onClick={handleCopy}
      >
        {copied ? <CheckIcon className="text-success" /> : <CopyIcon />}
      </Button>
      <span aria-live="polite" className="sr-only">
        {copied ? "Copied" : ""}
      </span>
    </div>
  );
}
