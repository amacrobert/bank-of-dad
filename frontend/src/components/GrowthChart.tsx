import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  ReferenceDot,
} from "recharts";
import { ProjectionDataPoint } from "../types";
import { GoalMarker, formatTimeToReach } from "../utils/goalMarkers";

export interface ScenarioLine {
  id: string;
  dataPoints: ProjectionDataPoint[];
  color: string;
  label: string;
}

interface GrowthChartProps {
  scenarios: ScenarioLine[];
  animationKey?: string;
  goalMarkers?: GoalMarker[];
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
  goalMarkersByWeek,
}: {
  active?: boolean;
  payload?: TooltipEntry[];
  scenarioLabels: string[];
  goalMarkersByWeek: Map<number, GoalMarker[]>;
}) {
  if (!active || !payload || payload.length === 0) return null;
  const date = payload[0]?.payload?.date;
  const weekIndex = payload[0]?.payload?.weekIndex;
  const weekGoals = weekIndex != null ? goalMarkersByWeek.get(weekIndex) : undefined;
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
      {weekGoals && weekGoals.length > 0 && (
        <>
          <div className="border-t border-sand my-1.5" />
          {weekGoals.map((m) => (
            <div key={`${m.goalId}-${m.scenarioId}`} className="flex items-center gap-1.5">
              <span
                className="inline-block w-2 h-2 rounded-full flex-shrink-0"
                style={{ backgroundColor: m.scenarioColor }}
              />
              {m.goalEmoji && <span className="text-sm">{m.goalEmoji}</span>}
              <span className="text-bark font-medium text-xs">{m.goalName}</span>
              <span className="text-bark-light text-xs ml-auto pl-2">
                {formatTimeToReach(m.weekIndex)}
              </span>
            </div>
          ))}
        </>
      )}
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

/** Build lookup from weekIndex → GoalMarker[] for tooltip integration */
function buildGoalMarkersByWeek(markers: GoalMarker[]): Map<number, GoalMarker[]> {
  const map = new Map<number, GoalMarker[]>();
  for (const m of markers) {
    const group = map.get(m.weekIndex);
    if (group) {
      group.push(m);
    } else {
      map.set(m.weekIndex, [m]);
    }
  }
  return map;
}

/** Deduplicate markers to one dot per scenario+weekIndex for ReferenceDot rendering */
function deduplicateMarkerDots(markers: GoalMarker[]): GoalMarker[] {
  const seen = new Set<string>();
  const result: GoalMarker[] = [];
  for (const m of markers) {
    const key = `${m.scenarioId}_${m.weekIndex}`;
    if (!seen.has(key)) {
      seen.add(key);
      result.push(m);
    }
  }
  return result;
}

export default function GrowthChart({ scenarios, animationKey, goalMarkers = [] }: GrowthChartProps) {
  const mergedData = mergeScenarioData(scenarios);
  const labels = scenarios.map((s) => s.label);
  const goalMarkersByWeek = buildGoalMarkersByWeek(goalMarkers);

  // Show a subset of x-axis labels to avoid crowding
  const tickCount = mergedData.length <= 14 ? mergedData.length : 6;
  const tickInterval = Math.max(1, Math.floor((mergedData.length - 1) / (tickCount - 1)));
  const ticks = mergedData
    .filter((_, i) => i % tickInterval === 0 || i === mergedData.length - 1)
    .map((dp) => dp.date);

  const uniqueMarkerDots = deduplicateMarkerDots(goalMarkers);

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
          <Tooltip content={<CustomTooltip scenarioLabels={labels} goalMarkersByWeek={goalMarkersByWeek} />} />
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
          {uniqueMarkerDots.map((m) => {
            const key = `${m.scenarioId}_${m.weekIndex}`;
            const dataKey = `balanceCents_${m.scenarioId}`;
            return (
              <ReferenceDot
                key={key}
                x={m.date}
                y={m.balanceCents}
                r={6}
                fill="#fff"
                stroke={m.scenarioColor}
                strokeWidth={2.5}
                ifOverflow="extendDomain"
                /* suppress the dataKey warning — we position via x/y */
                {...({ dataKey } as Record<string, string>)}
              />
            );
          })}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
