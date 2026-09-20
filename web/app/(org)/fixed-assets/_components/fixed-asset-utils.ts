import type { AssetDepreciationLine, FixedAsset } from "@/lib/services/swantara";

export type FixedAssetState = FixedAsset["state"];

export function canGenerateSchedule(state: FixedAssetState): boolean {
  return state === "running";
}

export function canPostDepreciation(
  state: FixedAssetState,
  lines: AssetDepreciationLine[],
): boolean {
  return state === "running" && lines.some((l) => !l.posted);
}

export function canDispose(state: FixedAssetState): boolean {
  return state === "running";
}

export function fixedAssetStateTone(
  state: FixedAssetState,
): "neutral" | "success" | "warning" | "danger" | "info" {
  switch (state) {
    case "draft":
      return "neutral";
    case "running":
      return "info";
    case "sold":
      return "success";
    case "disposed":
      return "warning";
    default:
      return "neutral";
  }
}

export function nbv(purchaseValue: number, lines: AssetDepreciationLine[]): number {
  const accumulated = lines.filter((l) => l.posted).reduce((sum, l) => sum + l.accumulated, 0);
  return purchaseValue - accumulated;
}

export function totalDepreciation(lines: AssetDepreciationLine[]): number {
  return lines.filter((l) => l.posted).reduce((sum, l) => sum + l.amount, 0);
}
