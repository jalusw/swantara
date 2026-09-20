import { Skeleton } from "@/components/skeleton";

export function DetailPageSkeleton() {
  return (
    <div data-slot="detail-page-skeleton" className="flex flex-col gap-4">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-40 w-full" />
    </div>
  );
}
