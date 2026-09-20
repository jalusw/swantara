import { Badge } from "@/components/badge";

export function ActiveBadge({ active, children }: { active: boolean; children: React.ReactNode }) {
  return (
    <Badge data-slot="active-badge" variant={active ? "default" : "outline"}>
      {children}
    </Badge>
  );
}
