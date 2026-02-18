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

export default function ScenarioControls({
  scenario,
  onChange,
  currentBalanceCents,
}: ScenarioControlsProps) {
  const withdrawalExceedsBalance =
    scenario.oneTimeWithdrawalCents > currentBalanceCents;

  const handleChange = (field: keyof ScenarioInputs, value: string) => {
    const cents = dollarsToCents(value);
    onChange({ ...scenario, [field]: cents });
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
          value={centsToDollars(scenario.weeklySpendingCents)}
          onChange={(e) => handleChange("weeklySpendingCents", e.target.value)}
        />
        <Input
          label="One-time extra deposit"
          id="one-time-deposit"
          type="number"
          min="0"
          step="0.01"
          placeholder="$0.00"
          value={centsToDollars(scenario.oneTimeDepositCents)}
          onChange={(e) => handleChange("oneTimeDepositCents", e.target.value)}
        />
        <Input
          label="One-time withdrawal"
          id="one-time-withdrawal"
          type="number"
          min="0"
          step="0.01"
          placeholder="$0.00"
          value={centsToDollars(scenario.oneTimeWithdrawalCents)}
          onChange={(e) => handleChange("oneTimeWithdrawalCents", e.target.value)}
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
