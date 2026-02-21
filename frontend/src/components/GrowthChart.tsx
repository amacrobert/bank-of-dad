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
import { useTheme } from "../context/ThemeContext";
import { getTheme } from "../themes";

interface GrowthChartProps {
  dataPoints: ProjectionDataPoint[];
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

interface TooltipPayload {
  value: number;
  payload: ProjectionDataPoint;
}

function CustomTooltip({ active, payload }: { active?: boolean; payload?: TooltipPayload[] }) {
  if (!active || !payload || payload.length === 0) return null;
  const point = payload[0];
  return (
    <div className="bg-white border border-sand rounded-xl px-3 py-2 shadow-lg text-sm">
      <p className="text-bark-light text-xs">{formatTooltipDate(point.payload.date)}</p>
      <p className="font-bold text-forest">
        {formatDollars(point.value)}
      </p>
    </div>
  );
}

export default function GrowthChart({ dataPoints }: GrowthChartProps) {
  const { theme } = useTheme();
  const themeColor = getTheme(theme).colors.forest;

  // Show a subset of x-axis labels to avoid crowding
  const tickCount = dataPoints.length <= 14 ? dataPoints.length : 6;
  const tickInterval = Math.max(1, Math.floor((dataPoints.length - 1) / (tickCount - 1)));
  const ticks = dataPoints
    .filter((_, i) => i % tickInterval === 0 || i === dataPoints.length - 1)
    .map((dp) => dp.date);

  return (
    <div className="w-full h-64 sm:h-72">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={dataPoints} margin={{ top: 8, right: 12, bottom: 0, left: 12 }}>
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
          <Tooltip content={<CustomTooltip />} />
          <Line
            type="monotone"
            dataKey="balanceCents"
            stroke={themeColor}
            strokeWidth={2.5}
            dot={false}
            activeDot={{ r: 4, fill: themeColor, stroke: "#fff", strokeWidth: 2 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
