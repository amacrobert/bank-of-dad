import { ScenarioConfig, ScenarioInputs, Frequency, SCENARIO_COLORS } from "../types";
import { weeksPerPeriod } from "./projection";

export function mapScenarioConfigToInputs(
  config: ScenarioConfig,
  horizonMonths: number
): ScenarioInputs {
  return {
    weeklySpendingCents: config.weeklyDirection === "spending" ? config.weeklyAmountCents : 0,
    weeklySavingsCents: config.weeklyDirection === "saving" ? config.weeklyAmountCents : 0,
    oneTimeDepositCents: config.oneTimeDirection === "deposit" ? config.oneTimeAmountCents : 0,
    oneTimeWithdrawalCents: config.oneTimeDirection === "withdrawal" ? config.oneTimeAmountCents : 0,
    horizonMonths,
  };
}

export function buildDefaultScenarios(
  allowanceAmountCents: number,
  allowanceFrequency: Frequency | null
): ScenarioConfig[] {
  const hasAllowance = allowanceFrequency !== null && allowanceAmountCents > 0;

  const base: Pick<ScenarioConfig, "oneTimeAmountCents" | "oneTimeDirection"> = {
    oneTimeAmountCents: 0,
    oneTimeDirection: "deposit",
  };

  if (hasAllowance) {
    const weeklyAllowance = Math.round(allowanceAmountCents / weeksPerPeriod(allowanceFrequency!));
    return [
      { id: "s0", weeklyAmountCents: 0, weeklyDirection: "spending", color: SCENARIO_COLORS[0], ...base },
      { id: "s1", weeklyAmountCents: weeklyAllowance, weeklyDirection: "spending", color: SCENARIO_COLORS[1], ...base },
    ];
  }

  return [
    { id: "s0", weeklyAmountCents: 0, weeklyDirection: "spending", color: SCENARIO_COLORS[0], ...base },
    { id: "s1", weeklyAmountCents: 500, weeklyDirection: "spending", color: SCENARIO_COLORS[1], ...base },
  ];
}

export function getNextScenarioColor(existingScenarios: ScenarioConfig[]): string {
  const usedColors = new Set(existingScenarios.map((s) => s.color));
  for (const color of SCENARIO_COLORS) {
    if (!usedColors.has(color)) return color;
  }
  return SCENARIO_COLORS[0];
}

export function getNextScenarioId(existingScenarios: ScenarioConfig[]): string {
  let maxNum = -1;
  for (const s of existingScenarios) {
    const num = parseInt(s.id.replace("s", ""), 10);
    if (!isNaN(num) && num > maxNum) maxNum = num;
  }
  return `s${maxNum + 1}`;
}
