import { useState, useEffect, useRef } from "react";
import { ScenarioInputs } from "../types";
import Input from "./ui/Input";

interface ScenarioControlsProps {
  scenario: ScenarioInputs;
  onChange: (scenario: ScenarioInputs) => void;
  currentBalanceCents: number;
}

function centsToDollars(cents: number): string {
  if (cents === 0) return "";
  return (cents / 100).toFixed(2);
}

function dollarsToCents(value: string): number {
  const parsed = parseFloat(value);
  if (isNaN(parsed) || parsed < 0) return 0;
  return Math.round(parsed * 100);
}

const fields = [
  "weeklySpendingCents",
  "oneTimeDepositCents",
  "oneTimeWithdrawalCents",
] as const;

export default function ScenarioControls({
  scenario,
  onChange,
  currentBalanceCents,
}: ScenarioControlsProps) {
  const withdrawalExceedsBalance =
    scenario.oneTimeWithdrawalCents > currentBalanceCents;

  const [localWeekly, setLocalWeekly] = useState(() =>
    centsToDollars(scenario.weeklySpendingCents)
  );
  const [localDeposit, setLocalDeposit] = useState(() =>
    centsToDollars(scenario.oneTimeDepositCents)
  );
  const [localWithdrawal, setLocalWithdrawal] = useState(() =>
    centsToDollars(scenario.oneTimeWithdrawalCents)
  );

  // Track last-known cents to detect external changes (e.g. reset)
  const prevCents = useRef({
    weeklySpendingCents: scenario.weeklySpendingCents,
    oneTimeDepositCents: scenario.oneTimeDepositCents,
    oneTimeWithdrawalCents: scenario.oneTimeWithdrawalCents,
  });

  useEffect(() => {
    for (const field of fields) {
      if (scenario[field] !== prevCents.current[field]) {
        const formatted = centsToDollars(scenario[field]);
        if (field === "weeklySpendingCents") setLocalWeekly(formatted);
        if (field === "oneTimeDepositCents") setLocalDeposit(formatted);
        if (field === "oneTimeWithdrawalCents") setLocalWithdrawal(formatted);
      }
    }
    prevCents.current = {
      weeklySpendingCents: scenario.weeklySpendingCents,
      oneTimeDepositCents: scenario.oneTimeDepositCents,
      oneTimeWithdrawalCents: scenario.oneTimeWithdrawalCents,
    };
  }, [scenario]);

  const handleBlur = (field: keyof ScenarioInputs, localValue: string) => {
    const cents = dollarsToCents(localValue);
    // Update parent
    onChange({ ...scenario, [field]: cents });
    // Reformat display
    const formatted = centsToDollars(cents);
    if (field === "weeklySpendingCents") setLocalWeekly(formatted);
    if (field === "oneTimeDepositCents") setLocalDeposit(formatted);
    if (field === "oneTimeWithdrawalCents") setLocalWithdrawal(formatted);
    // Keep ref in sync so useEffect doesn't re-override
    prevCents.current = { ...prevCents.current, [field]: cents };
  };

  return (
    <div className="space-y-4">
      <h3 className="text-sm font-bold text-bark uppercase tracking-wide">
        What if...
      </h3>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Input
          label="Weekly spending"
          id="weekly-spending"
          type="number"
          min="0"
          step="0.01"
          placeholder="$0.00"
          value={localWeekly}
          onChange={(e) => setLocalWeekly(e.target.value)}
          onBlur={() => handleBlur("weeklySpendingCents", localWeekly)}
        />
        <Input
          label="One-time extra deposit"
          id="one-time-deposit"
          type="number"
          min="0"
          step="0.01"
          placeholder="$0.00"
          value={localDeposit}
          onChange={(e) => setLocalDeposit(e.target.value)}
          onBlur={() => handleBlur("oneTimeDepositCents", localDeposit)}
        />
        <Input
          label="One-time withdrawal"
          id="one-time-withdrawal"
          type="number"
          min="0"
          step="0.01"
          placeholder="$0.00"
          value={localWithdrawal}
          onChange={(e) => setLocalWithdrawal(e.target.value)}
          onBlur={() => handleBlur("oneTimeWithdrawalCents", localWithdrawal)}
          error={
            withdrawalExceedsBalance
              ? `You can't withdraw more than your $${(currentBalanceCents / 100).toFixed(2)} balance`
              : null
          }
        />
      </div>
    </div>
  );
}
