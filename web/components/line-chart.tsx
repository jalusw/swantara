import { cn } from "@/lib/utils";

export type LineChartDatum = {
  label: string;
  value: number;
};

export type LineChartProps = {
  data: LineChartDatum[];
  variant?: "line" | "area";
  ariaLabel: string;
  valueFormatter?: (value: number) => string;
  showGridLines?: boolean;
  className?: string;
};

const width = 100;
const height = 40;
const padding = 4;

export function LineChart({
  data,
  variant = "area",
  ariaLabel,
  valueFormatter,
  showGridLines = true,
  className,
}: LineChartProps) {
  const values = data.map((datum) => datum.value);
  const max = Math.max(...values, 1);
  const min = Math.min(...values, 0);

  const range = max - min || 1;
  const points = data.map((datum, index) => {
    const x = data.length === 1 ? width / 2 : (index / (data.length - 1)) * width;
    const y = padding + (height - padding * 2) * (1 - (datum.value - min) / range);
    return `${x.toFixed(2)},${y.toFixed(2)}`;
  });

  const linePoints = points.join(" ");
  const areaPoints = `0,${height} ${linePoints} ${width},${height}`;

  return (
    <div
      data-slot="line-chart"
      role="img"
      aria-label={ariaLabel}
      className={cn("h-44 w-full", className)}
    >
      <svg
        viewBox={`0 0 ${width} ${height}`}
        preserveAspectRatio="none"
        className="block h-full w-full"
      >
        <title>{ariaLabel}</title>
        {showGridLines ? (
          <g data-slot="line-chart-grid" className="text-border">
            {[0, 0.5, 1].map((fraction) => (
              <line
                key={fraction}
                x1="0"
                x2={width}
                y1={padding + (height - padding * 2) * fraction}
                y2={padding + (height - padding * 2) * fraction}
                stroke="currentColor"
                strokeDasharray="2 2"
                strokeWidth="0.5"
                vectorEffect="non-scaling-stroke"
              />
            ))}
          </g>
        ) : null}
        {variant === "area" ? (
          <polygon
            points={areaPoints}
            fill="currentColor"
            opacity={0.15}
            className="text-primary"
          />
        ) : null}
        <polyline
          points={linePoints}
          fill="none"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          vectorEffect="non-scaling-stroke"
          className="stroke-primary"
        />
        {data.map((datum, index) => {
          const [x, y] = points[index]?.split(",") ?? ["0", "0"];
          return (
            <circle
              key={`${datum.label}-${index}`}
              cx={Number(x)}
              cy={Number(y)}
              r="1.4"
              vectorEffect="non-scaling-stroke"
              className="fill-background stroke-primary"
              strokeWidth="1.5"
              {...{
                title: `${
                  datum.label
                }: ${valueFormatter ? valueFormatter(datum.value) : datum.value}`,
              }}
            />
          );
        })}
      </svg>
      <ul className="mt-2 flex justify-between gap-2 text-[10px] text-muted-foreground" aria-hidden>
        {data.map((datum, index) => (
          <li key={`${datum.label}-${index}`} className="truncate">
            {datum.label}
          </li>
        ))}
      </ul>
    </div>
  );
}
