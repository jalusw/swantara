"use client";

import { SendIcon } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { Button } from "./button";
import { Textarea } from "./textarea";

export type ThreadMessage = {
  id: string;
  author: string;
  body: React.ReactNode;
  timestamp?: string;
};

export type ThreadProps = {
  messages: ThreadMessage[];
  onSubmit: (content: string) => void | Promise<void>;
  placeholder?: string;
  className?: string;
};

function initials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join("");
}

export function Thread({
  messages,
  onSubmit,
  placeholder = "Add a comment…",
  className,
}: ThreadProps) {
  const [draft, setDraft] = useState("");

  async function submit() {
    const content = draft.trim();
    if (!content) {
      return;
    }
    try {
      await onSubmit(content);
      setDraft("");
    } catch {
      // draft stays in input so user can retry
    }
  }

  return (
    <div data-slot="thread" className={cn("flex flex-col gap-3", className)}>
      <ul className="flex flex-col gap-3" aria-live="polite" aria-relevant="additions">
        {messages.map((message) => (
          <li key={message.id} data-slot="thread-message" className="flex gap-2.5">
            <span
              aria-hidden
              className="grid size-8 shrink-0 place-items-center rounded-full bg-muted text-xs text-muted-foreground ring-1 ring-border"
            >
              {initials(message.author)}
            </span>
            <div className="min-w-0 flex-1 space-y-0.5">
              <div className="flex items-baseline gap-2">
                <span className="text-sm">{message.author}</span>
                {message.timestamp ? (
                  <time className="text-xs text-muted-foreground" dateTime={message.timestamp}>
                    {message.timestamp}
                  </time>
                ) : null}
              </div>
              <p className="text-sm text-foreground">{message.body}</p>
            </div>
          </li>
        ))}
      </ul>

      <div className="flex items-start gap-2">
        <Textarea
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          placeholder={placeholder}
          aria-label={"Write a comment"}
          rows={2}
          className="flex-1"
        />
        <Button
          type="button"
          size="icon"
          aria-label={"Send comment"}
          disabled={draft.trim().length === 0}
          onClick={() => void submit()}
        >
          <SendIcon aria-hidden />
        </Button>
      </div>
    </div>
  );
}
