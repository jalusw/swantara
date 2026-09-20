"use client";

import { Search } from "lucide-react";
import dynamic from "next/dynamic";
import { Button } from "@/components/button";
import { useCommandPalette } from "@/components/command-palette";

const OrgCommandDialog = dynamic(
  () => import("./org-command-dialog").then((mod) => ({ default: mod.OrgCommandDialog })),
  {
    ssr: false,
  },
);

export function OrgSearch() {
  const { open, setOpen } = useCommandPalette();

  return (
    <>
      <Button
        variant="outline"
        className="w-full max-w-xs justify-start text-muted-foreground"
        onClick={() => setOpen(true)}
      >
        <Search aria-hidden />
        <span className="truncate">{"Search..."}</span>
        <kbd className="ml-auto rounded border border-border bg-muted px-1.5 py-0.5 font-sans text-[10px] ">
          ⌘K
        </kbd>
      </Button>
      {open && <OrgCommandDialog open={open} onOpenChange={setOpen} />}
    </>
  );
}
