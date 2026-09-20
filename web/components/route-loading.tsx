import { Loader2Icon } from "lucide-react";

export function RouteLoading({
  variant = "page",
  rows = 3,
}: {
  variant?: "page" | "table";
  rows?: number;
}) {
  if (variant === "table") {
    return (
      <div data-slot="route-loading" className="flex flex-col gap-2 p-4">
        {Array.from({ length: rows }, (_, i) => (
          <div key={i} className="h-8 animate-pulse rounded bg-muted" />
        ))}
      </div>
    );
  }

  return (
    <div data-slot="route-loading" className="flex items-center justify-center py-12">
      <Loader2Icon className="size-6 animate-spin text-muted-foreground" />
    </div>
  );
}
