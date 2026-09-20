"use client";

import { useMemo } from "react";

import { cn } from "@/lib/utils";

const PATTERNS = [
  "11011001100",
  "11001101100",
  "11001100110",
  "10010011000",
  "10010001100",
  "10001001100",
  "10011001000",
  "10011000100",
  "10001100100",
  "11001001000",
  "11001000100",
  "11000100100",
  "10110011100",
  "10011011100",
  "10011001110",
  "10111001100",
  "10011101100",
  "10011100110",
  "11001110010",
  "11001011100",
  "11001001110",
  "11011100100",
  "11001110100",
  "11101101110",
  "11101001100",
  "11100101100",
  "11100100110",
  "11101100100",
  "11100110100",
  "11100110010",
  "11011011000",
  "11011000110",
  "11000110110",
  "10100011000",
  "10001011000",
  "10001000110",
  "10110001000",
  "10001101000",
  "10001100010",
  "11010001000",
  "11000101000",
  "11000100010",
  "10110111000",
  "10110001110",
  "10001101110",
  "10111011000",
  "10111000110",
  "10001110110",
  "11101110110",
  "11010001110",
  "11000101110",
  "11011101000",
  "11011100010",
  "11011101110",
  "11101011000",
  "11101000110",
  "11100010110",
  "11101101000",
  "11101100010",
  "11100011010",
  "11101111010",
  "11001000010",
  "11110001010",
  "10100110000",
  "10100001100",
  "10010110000",
  "10010000110",
  "10000101100",
  "10000100110",
  "10110010000",
  "10110000100",
  "10011010000",
  "10011000010",
  "10000110100",
  "10000110010",
  "11000010010",
  "11001010000",
  "11110111010",
  "11000010100",
  "10001111010",
  "10100111100",
  "10010111100",
  "10010011110",
  "10111100100",
  "10011110100",
  "10011110010",
  "11110100100",
  "11110010100",
  "11110010010",
  "11011011110",
  "11011110110",
  "11110110110",
  "10101111000",
  "10100011110",
  "10001011110",
  "10111101000",
  "10111100010",
  "11110101000",
  "11110100010",
  "10111011110",
  "10111101110",
  "11101011110",
  "11110101110",
  "11010000100",
  "11010010000",
  "11010011100",
];

const START_B = 104;
const STOP_PATTERN = "1100011101011";

export function encodeCode128B(value: string): string {
  let bits = `${PATTERNS[START_B]}`;
  let checksum = START_B;

  let weight = 1;
  for (let index = 0; index < value.length; index += 1) {
    const code = value.charCodeAt(index);
    if (code < 32 || code > 126) {
      continue;
    }
    const glyph = code - 32;
    bits += PATTERNS[glyph];
    checksum = (checksum + glyph * weight) % 103;
    weight += 1;
  }

  return `${bits}${PATTERNS[checksum]}${STOP_PATTERN}`;
}

export type BarcodeProps = {
  value: string;
  height?: number;
  barWidth?: number;
  quietZone?: number;
  showText?: boolean;
  className?: string;
  "aria-label"?: string;
};

export function Barcode({
  value,
  height = 48,
  barWidth = 1,
  quietZone = 10,
  showText = true,
  className,
  "aria-label": ariaLabel,
}: BarcodeProps) {
  const bits = useMemo(() => {
    if (!value.trim()) {
      return "";
    }
    try {
      return encodeCode128B(value);
    } catch {
      return "";
    }
  }, [value]);

  const bars: number[] = [];
  for (let index = 0; index < bits.length; index += 1) {
    if (bits[index] === "1") {
      bars.push(index);
    }
  }

  if (bars.length === 0) {
    return (
      <span data-slot="barcode" className={cn("text-sm text-muted-foreground", className)}>
        Empty barcode
      </span>
    );
  }

  const modules = bits.length + quietZone * 2;
  const width = modules * barWidth;

  return (
    <div data-slot="barcode" className={cn("flex w-full flex-col items-center gap-1.5", className)}>
      <svg
        width="100%"
        height={height}
        viewBox={`0 0 ${width} ${height}`}
        role="img"
        aria-label={ariaLabel ?? `Barcode ${value}`}
        className="mx-auto block max-w-full text-foreground"
        preserveAspectRatio="none"
      >
        {bars.map((position) => (
          <rect
            key={position}
            x={(position + quietZone) * barWidth}
            y={0}
            width={barWidth}
            height={height}
            fill="currentColor"
          />
        ))}
      </svg>
      {showText ? (
        <span className="font-mono text-xs tracking-widest text-foreground tabular-nums">
          {value}
        </span>
      ) : null}
    </div>
  );
}
