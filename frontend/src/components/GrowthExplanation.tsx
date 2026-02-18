import { ProjectionResult, Frequency } from "../types";
import { weeksPerPeriod } from "../utils/projection";

interface GrowthExplanationProps {
  projection: ProjectionResult;
  horizonMonths: number;
  hasAllowance: boolean;
  hasInterest: boolean;
  isAllowancePaused: boolean;
  isInterestPaused: boolean;
  weeklyAllowanceCentsDisplay: number;
  allowanceFrequencyDisplay: string;
  weeklySpendingCents: number;
}

function formatDollars(cents: number): string {
  return "$" + (cents / 100).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

function horizonLabel(months: number): string {
  if (months < 12) return `${months} month${months === 1 ? "" : "s"}`;
  const years = months / 12;
  if (Number.isInteger(years)) return `${years} year${years === 1 ? "" : "s"}`;
  return `${months} months`;
}

export default function GrowthExplanation({
  projection,
  horizonMonths,
  hasAllowance,
  hasInterest,
  isAllowancePaused,
  isInterestPaused,
  weeklyAllowanceCentsDisplay,
  allowanceFrequencyDisplay,
  weeklySpendingCents,
}: GrowthExplanationProps) {
  const {
    finalBalanceCents,
    startingBalanceCents,
    totalInterestCents,
    totalAllowanceCents,
    totalSpendingCents,
    depletionWeek,
  } = projection;

  const horizon = horizonLabel(horizonMonths);
  const noGrowthSources = !hasAllowance && !hasInterest;

  // No growth sources and no spending — encouraging message
  if (noGrowthSources && weeklySpendingCents === 0) {
    return (
      <div className="text-sm text-bark-light leading-relaxed">
        <p>
          Your balance will stay at{" "}
          <span className="font-semibold text-bark">{formatDollars(startingBalanceCents)}</span>{" "}
          unless you get an allowance or interest set up.{" "}
          <span className="text-forest font-medium">Ask your parent!</span>
        </p>
      </div>
    );
  }

  // Build intro sentence based on allowance + spending relationship
  let intro: string;

  if (hasAllowance && weeklySpendingCents > 0) {
    const spendingPerPeriod = Math.round(
      weeklySpendingCents * weeksPerPeriod(allowanceFrequencyDisplay as Frequency)
    );
    const savingsPerPeriod = weeklyAllowanceCentsDisplay - spendingPerPeriod;

    if (savingsPerPeriod > 0) {
      intro = `If you spend ${formatDollars(weeklySpendingCents)} each week and save ${formatDollars(savingsPerPeriod)} from your ${formatDollars(weeklyAllowanceCentsDisplay)} ${allowanceFrequencyDisplay} allowance, in ${horizon}`;
    } else {
      intro = `If you spend ${formatDollars(weeklySpendingCents)} each week and save none of your ${formatDollars(weeklyAllowanceCentsDisplay)} ${allowanceFrequencyDisplay} allowance, in ${horizon}`;
    }
  } else if (hasAllowance && weeklySpendingCents === 0) {
    intro = `If you keep saving your ${formatDollars(weeklyAllowanceCentsDisplay)} ${allowanceFrequencyDisplay} allowance, in ${horizon}`;
  } else if (weeklySpendingCents > 0) {
    intro = `If you spend ${formatDollars(weeklySpendingCents)} each week, in ${horizon}`;
  } else {
    intro = `In ${horizon}`;
  }

  // Build breakdown pieces
  const breakdown: string[] = [];
  breakdown.push(`${formatDollars(startingBalanceCents)} from what you have now`);
  if (totalInterestCents > 0) {
    breakdown.push(`${formatDollars(totalInterestCents)} from interest`);
  }
  if (totalAllowanceCents > 0) {
    breakdown.push(`${formatDollars(totalAllowanceCents)} from allowance deposits`);
  }

  const spendingNote = totalSpendingCents > 0
    ? ` That's after subtracting ${formatDollars(totalSpendingCents)} in spending.`
    : "";

  return (
    <div className="text-sm text-bark-light leading-relaxed space-y-2">
      <p>
        {intro} your account will have{" "}
        <span className="font-semibold text-forest text-base">{formatDollars(finalBalanceCents)}</span>{" "}
        total. That's {breakdown.join(", ")}.{spendingNote}
      </p>

      {depletionWeek !== null && (
        <p className="text-terracotta font-medium">
          At this rate, your account will run out of money in about {depletionWeek} week{depletionWeek === 1 ? "" : "s"}.
        </p>
      )}

      {(isAllowancePaused || isInterestPaused) && (
        <p className="text-amber-600 text-xs">
          {isAllowancePaused && "Your allowance is currently paused."}
          {isAllowancePaused && isInterestPaused && " "}
          {isInterestPaused && "Your interest is currently paused."}
          {" "}This projection doesn't include paused schedules.
        </p>
      )}
    </div>
  );
}
