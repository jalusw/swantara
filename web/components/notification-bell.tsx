"use client";

import { BellIcon } from "lucide-react";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";
import { Badge } from "./badge";
import { Button } from "./button";
import { Popover, PopoverContent, PopoverHeader, PopoverTitle, PopoverTrigger } from "./popover";
import { ScrollArea } from "./scroll-area";
import { Tooltip, TooltipContent, TooltipTrigger } from "./tooltip";

export type Notification = {
  id: string;
  title: string;
  description?: string;
  unread?: boolean;
};

export type NotificationBellProps = {
  items?: Notification[];
  onRead?: (id: string) => void;
  unreadCount?: number;
  tooltip?: string;
  className?: string;
  "aria-label"?: string;
};

export function NotificationBell({
  items = [],
  onRead,
  unreadCount,
  tooltip,
  className,
  "aria-label": ariaLabel,
}: NotificationBellProps) {
  const t = useTranslations("Notifications");
  const resolvedAriaLabel = ariaLabel ?? tooltip ?? t("title");
  const count = unreadCount ?? items.filter((item) => item.unread).length;
  const label = count > 0 ? t("unread", { count }) : resolvedAriaLabel;

  const trigger = (
    <PopoverTrigger
      render={<Button variant="ghost" size="icon" aria-label={label} className={className} />}
    >
      <BellIcon aria-hidden />
      {count > 0 ? (
        <Badge
          variant="destructive"
          className={cn("absolute -top-1 -right-1 h-4 min-w-4 rounded-full px-1 text-[10px]")}
        >
          {count > 99 ? "99+" : count}
        </Badge>
      ) : null}
    </PopoverTrigger>
  );

  return (
    <Popover>
      {tooltip ? (
        <Tooltip>
          <TooltipTrigger render={trigger} />
          <TooltipContent>{tooltip}</TooltipContent>
        </Tooltip>
      ) : (
        trigger
      )}
      <PopoverContent align="end" className="w-80">
        <PopoverHeader>
          <PopoverTitle>{t("title")}</PopoverTitle>
        </PopoverHeader>
        <ScrollArea className="max-h-80">
          <ul data-slot="notification-list" className="flex flex-col">
            {items.length === 0 ? (
              <li className="px-2 py-6 text-center text-sm text-muted-foreground">{t("empty")}</li>
            ) : (
              items.map((item) => (
                <li key={item.id}>
                  <button
                    type="button"
                    data-unread={item.unread || undefined}
                    onClick={() => onRead?.(item.id)}
                    className="flex min-h-11 w-full flex-col gap-0.5 rounded-md px-2 py-2 text-left text-sm outline-none transition-colors hover:bg-muted focus-visible:ring-2 ring-offset-2 ring-offset-background focus-visible:ring-ring data-unread:bg-muted/60"
                  >
                    <span className="flex items-center gap-2 ">
                      {item.unread ? (
                        <>
                          <span className="sr-only">Unread: </span>
                          <span aria-hidden className="size-1.5 rounded-full bg-primary" />
                        </>
                      ) : null}
                      {item.title}
                    </span>
                    {item.description ? (
                      <span className="text-xs text-muted-foreground">{item.description}</span>
                    ) : null}
                  </button>
                </li>
              ))
            )}
          </ul>
        </ScrollArea>
      </PopoverContent>
    </Popover>
  );
}
