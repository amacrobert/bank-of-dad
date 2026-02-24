import { ScenarioTitleContext } from "../types";

function formatDollars(cents: number): string {
  const dollars = cents / 100;
  return "$" + (Number.isInteger(dollars) ? dollars.toString() : dollars.toFixed(2));
}

export function generateScenarioTitle(ctx: ScenarioTitleContext): string {
  const {
    hasAllowance,
    weeklyAllowanceCents,
    weeklyAmountCents,
    weeklyDirection,
    oneTimeAmountCents,
    oneTimeDirection,
  } = ctx;

  const hasOneTime = oneTimeAmountCents > 0;
  const hasWeekly = weeklyAmountCents > 0;
  const allowanceStr = formatDollars(weeklyAllowanceCents);

  let weeklyPart: string;
  let oneTimePart = "";

  if (hasAllowance) {
    if (weeklyDirection === "saving" && hasWeekly) {
      // Saving on top of full allowance
      weeklyPart = `save **all** of my ${allowanceStr} allowance plus an additional **${formatDollars(weeklyAmountCents)}** per week`;
    } else if (!hasWeekly) {
      // Spending 0 = saving all
      weeklyPart = `save **all** of my ${allowanceStr} allowance`;
    } else if (weeklyAmountCents > weeklyAllowanceCents) {
      // Spending more than allowance
      const excessCents = weeklyAmountCents - weeklyAllowanceCents;
      weeklyPart = `spend my whole ${allowanceStr} allowance plus another **${formatDollars(excessCents)}** per week`;
    } else if (weeklyAmountCents === weeklyAllowanceCents) {
      // Spending exactly all
      weeklyPart = `save **none** of my ${allowanceStr} allowance`;
    } else {
      // Spending some, saving the rest
      const savingCents = weeklyAllowanceCents - weeklyAmountCents;
      weeklyPart = `save **${formatDollars(savingCents)}** per week from my ${allowanceStr} allowance`;
    }
  } else {
    // No allowance
    if (!hasWeekly && !hasOneTime) {
      return "If I don't do anything";
    }
    if (!hasWeekly) {
      // Only one-time action, no weekly
      weeklyPart = "";
    } else if (weeklyDirection === "spending") {
      weeklyPart = `spend **${formatDollars(weeklyAmountCents)}** per week`;
    } else {
      weeklyPart = `save **${formatDollars(weeklyAmountCents)}** per week`;
    }
  }

  // Build one-time suffix
  if (hasOneTime) {
    const amountStr = `**${formatDollars(oneTimeAmountCents)}**`;
    if (oneTimeDirection === "deposit") {
      if (!hasAllowance && !hasWeekly) {
        // Deposit is the only action
        return `If I deposit ${amountStr} now`;
      } else if (!hasAllowance && weeklyDirection === "spending") {
        // "deposit X now and spend Y per week"
        oneTimePart = `deposit ${amountStr} now and ${weeklyPart}`;
        return `If I ${oneTimePart}`;
      } else {
        oneTimePart = `, and deposit ${amountStr} now`;
      }
    } else {
      if (!hasAllowance && !hasWeekly) {
        return `If I withdraw ${amountStr} now`;
      } else if (!hasAllowance && weeklyDirection === "spending") {
        oneTimePart = `, and withdraw ${amountStr} now`;
      } else {
        oneTimePart = `, but withdraw ${amountStr} now`;
      }
    }
  }

  return `If I ${weeklyPart}${oneTimePart}`;
}
