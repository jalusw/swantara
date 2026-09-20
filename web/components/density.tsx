"use client";

import { useEffect } from "react";
import { cn } from "@/lib/utils";
import { useDensityStore } from "@/stores/density.store";
import { ToggleGroup, ToggleGroupItem } from "./toggle-group";

const COMPACT = "compact" as const;
const NORMAL = "normal" as const;
const COMFORTABLE = "comfortable" as const;

type DataDensity = typeof COMPACT | typeof NORMAL | typeof COMFORTABLE;

export type DensityToggleProps = {
  value?: DataDensity;
  onChange?: (density: DataDensity) => void;
  className?: string;
  labels?: {
    label?: string;
    compact?: string;
    normal?: string;
    comfortable?: string;
  };
};

export function useDataDensity(): DataDensity {
  return useDensityStore((s) => s.density);
}

export function DensityToggle({ value, onChange, className, labels }: DensityToggleProps) {
  const storeDensity = useDensityStore((s) => s.density);
  const setDensity = useDensityStore((s) => s.setDensity);
  const hydrateDensity = useDensityStore((s) => s.hydrateDensity);
  const allLabels = {
    label: "Data density",
    compact: "Compact",
    normal: "Normal",
    comfortable: "Comfortable",
    ...labels,
  };

  const current = value ?? storeDensity;

  useEffect(() => {
    if (value === undefined) hydrateDensity();
  }, [value, hydrateDensity]);

  const handleChange = (next: DataDensity) => {
    if (value === undefined) setDensity(next);
    onChange?.(next);
  };

  return (
    <ToggleGroup
      data-slot="density-toggle"
      value={[current]}
      onValueChange={(values) => {
        const next = values[0];
        if (next && next !== current) handleChange(next as DataDensity);
      }}
      aria-label={allLabels.label}
      size="sm"
      variant="outline"
      className={cn("w-fit", className)}
    >
      <ToggleGroupItem value="compact" aria-label={allLabels.compact}>
        {allLabels.compact}
      </ToggleGroupItem>
      <ToggleGroupItem value="normal" aria-label={allLabels.normal}>
        {allLabels.normal}
      </ToggleGroupItem>
      <ToggleGroupItem value="comfortable" aria-label={allLabels.comfortable}>
        {allLabels.comfortable}
      </ToggleGroupItem>
    </ToggleGroup>
  );
}
