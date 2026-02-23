import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
} from "recharts";
import { ProjectionDataPoint } from "../types";

export interface ScenarioLine {
  id: string;
  dataPoints: ProjectionDataPoint[];
  color: string;
  label: string;
}

interface GrowthChartProps {
  scenarios: ScenarioLine[];
  animationKey?: string;
}

function formatDollars(cents: number): string {
  return "$" + (cents / 100).toLocaleString(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  });
}

function formatDate(isoDate: string): string {
  const d = new Date(isoDate);
  return d.toLocaleDateString(undefined, { month: "short", year: "2-digit" });
}

function formatTooltipDate(isoDate: string): string {
  const d = new Date(isoDate);
  return d.toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" });
}

interface MergedDataPoint {
  weekIndex: number;
  date: string;
  [key: string]: number | string;
}

interface TooltipEntry {
  dataKey: string;
  value: number;
  color: string;
  payload: MergedDataPoint;
}

function CustomTooltip({
  active,
  payload,
  scenarioLabels,
}: {
  active?: boolean;
  payload?: TooltipEntry[];
  scenarioLabels: string[];
}) {
  if (!active || !payload || payload.length === 0) return null;
  const date = payload[0]?.payload?.date;
  return (
    <div className="bg-white border border-sand rounded-xl px-3 py-2 shadow-lg text-sm min-w-[140px]">
      {date && <p className="text-bark-light text-xs mb-1">{formatTooltipDate(date)}</p>}
      {payload.map((entry, i) => (
        <div key={entry.dataKey} className="flex items-center gap-2">
          <span
            className="inline-block w-2.5 h-2.5 rounded-full flex-shrink-0"
            style={{ backgroundColor: entry.color }}
          />
          <span className="text-bark-light text-xs truncate max-w-[200px]">{(scenarioLabels[i] ?? `Scenario ${i + 1}`).replace(/\*\*/g, "")}</span>
          <span className="font-bold ml-auto" style={{ color: entry.color }}>
            {formatDollars(entry.value)}
          </span>
        </div>
      ))}
    </div>
  );
}

function mergeScenarioData(scenarios: ScenarioLine[]): MergedDataPoint[] {
  if (scenarios.length === 0) return [];
  const basePoints = scenarios[0].dataPoints;
  return basePoints.map((bp, weekIdx) => {
    const point: MergedDataPoint = {
      weekIndex: bp.weekIndex,
      date: bp.date,
    };
    scenarios.forEach((s) => {
      point[`balanceCents_${s.id}`] = s.dataPoints[weekIdx]?.balanceCents ?? 0;
    });
    return point;
  });
}

export default function GrowthChart({ scenarios, animationKey }: GrowthChartProps) {
  const mergedData = mergeScenarioData(scenarios);
  const labels = scenarios.map((s) => s.label);

  // Show a subset of x-axis labels to avoid crowding
  const tickCount = mergedData.length <= 14 ? mergedData.length : 6;
  const tickInterval = Math.max(1, Math.floor((mergedData.length - 1) / (tickCount - 1)));
  const ticks = mergedData
    .filter((_, i) => i % tickInterval === 0 || i === mergedData.length - 1)
    .map((dp) => dp.date);

  return (
    <div className="w-full h-64 sm:h-72">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart key={animationKey} data={mergedData} margin={{ top: 8, right: 12, bottom: 0, left: 12 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#e5ddd4" vertical={false} />
          <XAxis
            dataKey="date"
            tickFormatter={formatDate}
            ticks={ticks}
            tick={{ fontSize: 11, fill: "#8a7a6b" }}
            axisLine={{ stroke: "#e5ddd4" }}
            tickLine={false}
          />
          <YAxis
            tickFormatter={formatDollars}
            tick={{ fontSize: 11, fill: "#8a7a6b" }}
            axisLine={false}
            tickLine={false}
            width={60}
          />
          <Tooltip content={<CustomTooltip scenarioLabels={labels} />} />
          {scenarios.map((s) => (
            <Line
              key={s.id}
              type="monotone"
              dataKey={`balanceCents_${s.id}`}
              stroke={s.color}
              strokeWidth={2.5}
              dot={false}
              activeDot={{ r: 4, fill: s.color, stroke: "#fff", strokeWidth: 2 }}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
