import { ProjectionDataPoint, SavingsGoal } from "../types";
import { ScenarioLine } from "../components/GrowthChart";

export interface GoalMarker {
  goalId: number;
  goalName: string;
  goalEmoji?: string;
  remainingCents: number;
  scenarioId: string;
  scenarioColor: string;
  weekIndex: number;
  date: string;
  balanceCents: number;
}

/**
 * Format a week count as a human-friendly time string.
 * e.g., 1 week → "in ~1 week", 8 weeks → "in ~2 months", 52 weeks → "in ~1 year"
 */
export function formatTimeToReach(weekIndex: number): string {
  if (weekIndex === 0) return "now";
  if (weekIndex <= 1) return "in ~1 week";
  if (weekIndex < 8) return `in ~${weekIndex} weeks`;

  const months = Math.round(weekIndex / (52 / 12));
  if (months < 2) return "in ~1 month";
  if (months < 12) return `in ~${months} months`;

  const years = Math.round(months / 12);
  if (years === 1) return "in ~1 year";
  return `in ~${years} years`;
}

/**
 * Calculate goal marker positions for each active savings goal on each scenario line.
 * Returns markers where the projected balance first reaches or exceeds the goal's remaining amount.
 */
export function calculateGoalMarkers(
  scenarios: ScenarioLine[],
  goals: SavingsGoal[],
): GoalMarker[] {
  const markers: GoalMarker[] = [];

  // Only consider active goals with remaining amount > 0
  const activeGoals = goals.filter(
    (g) => g.status === "active" && g.target_cents > g.saved_cents,
  );

  for (const goal of activeGoals) {
    const remainingCents = goal.target_cents - goal.saved_cents;

    for (const scenario of scenarios) {
      const marker = findFirstReach(scenario.dataPoints, remainingCents);
      if (marker !== null) {
        markers.push({
          goalId: goal.id,
          goalName: goal.name,
          goalEmoji: goal.emoji,
          remainingCents,
          scenarioId: scenario.id,
          scenarioColor: scenario.color,
          weekIndex: marker.weekIndex,
          date: marker.date,
          balanceCents: marker.balanceCents,
        });
      }
    }
  }

  return markers;
}

/**
 * Find the first data point where balance reaches or exceeds the target amount.
 */
function findFirstReach(
  dataPoints: ProjectionDataPoint[],
  targetCents: number,
): ProjectionDataPoint | null {
  for (const dp of dataPoints) {
    if (dp.balanceCents >= targetCents) {
      return dp;
    }
  }
  return null;
}
