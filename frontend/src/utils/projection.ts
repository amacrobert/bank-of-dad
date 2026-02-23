import { ProjectionConfig, ProjectionResult, ProjectionDataPoint, Frequency } from "../types";

const WEEKS_PER_MONTH = 52 / 12;

function horizonToWeeks(months: number): number {
  return Math.round(months * WEEKS_PER_MONTH);
}

function periodsPerYear(frequency: Frequency): number {
  switch (frequency) {
    case "weekly":
      return 52;
    case "biweekly":
      return 26;
    case "monthly":
      return 12;
  }
}

/** Fractional weeks between payments for a given frequency */
export function weeksPerPeriod(frequency: Frequency): number {
  switch (frequency) {
    case "weekly":
      return 1;
    case "biweekly":
      return 2;
    case "monthly":
      return WEEKS_PER_MONTH; // ~4.333
  }
}

export function calculateProjection(config: ProjectionConfig): ProjectionResult {
  const {
    currentBalanceCents,
    interestRateBps,
    interestFrequency,
    allowanceAmountCents,
    allowanceFrequency,
    scenario,
  } = config;

  const totalWeeks = horizonToWeeks(scenario.horizonMonths);

  // Calculate effective starting balance with one-time adjustments
  const withdrawal = Math.min(scenario.oneTimeWithdrawalCents, currentBalanceCents);
  const startingBalanceCents =
    currentBalanceCents + scenario.oneTimeDepositCents - withdrawal;

  // Interest config
  const hasInterest = interestFrequency !== null && interestRateBps > 0;
  const interestPerPeriod = hasInterest
    ? interestRateBps / 10000 / periodsPerYear(interestFrequency!)
    : 0;
  const interestInterval = hasInterest ? weeksPerPeriod(interestFrequency!) : 0;
  let nextInterestWeek = hasInterest ? interestInterval : Infinity;

  // Allowance config
  const hasAllowance = allowanceFrequency !== null && allowanceAmountCents > 0;
  const allowanceInterval = hasAllowance ? weeksPerPeriod(allowanceFrequency!) : 0;
  let nextAllowanceWeek = hasAllowance ? allowanceInterval : Infinity;

  // Tracking
  let balance = startingBalanceCents;
  let totalInterest = 0;
  let totalAllowance = 0;
  let totalSpending = 0;
  let totalSavings = 0;
  let depletionWeek: number | null = null;

  const dataPoints: ProjectionDataPoint[] = [];
  const startDate = new Date();

  // Week 0 = today
  dataPoints.push({
    weekIndex: 0,
    date: startDate.toISOString(),
    balanceCents: Math.max(0, balance),
  });

  for (let week = 1; week <= totalWeeks; week++) {
    // Apply allowance if due this week (using fractional tracking)
    if (hasAllowance && week >= Math.round(nextAllowanceWeek)) {
      balance += allowanceAmountCents;
      totalAllowance += allowanceAmountCents;
      nextAllowanceWeek += allowanceInterval;
    }

    // Apply interest if due this week (using fractional tracking)
    if (hasInterest && week >= Math.round(nextInterestWeek)) {
      const interest = Math.round(balance * interestPerPeriod);
      if (interest > 0) {
        balance += interest;
        totalInterest += interest;
      }
      nextInterestWeek += interestInterval;
    }

    // Apply weekly savings
    if (scenario.weeklySavingsCents > 0) {
      balance += scenario.weeklySavingsCents;
      totalSavings += scenario.weeklySavingsCents;
    }

    // Apply weekly spending
    if (scenario.weeklySpendingCents > 0) {
      balance -= scenario.weeklySpendingCents;
      totalSpending += scenario.weeklySpendingCents;
    }

    // Floor at $0
    if (balance < 0) {
      // Adjust spending to not exceed what was available
      totalSpending += balance; // balance is negative, so this reduces totalSpending
      balance = 0;
    }

    // Track depletion
    if (balance === 0 && depletionWeek === null && scenario.weeklySpendingCents > 0) {
      depletionWeek = week;
    }

    const weekDate = new Date(startDate);
    weekDate.setDate(weekDate.getDate() + week * 7);

    dataPoints.push({
      weekIndex: week,
      date: weekDate.toISOString(),
      balanceCents: balance,
    });
  }

  return {
    dataPoints,
    finalBalanceCents: balance,
    totalInterestCents: totalInterest,
    totalAllowanceCents: totalAllowance,
    totalSpendingCents: totalSpending,
    totalSavingsCents: totalSavings,
    startingBalanceCents,
    depletionWeek,
  };
}
