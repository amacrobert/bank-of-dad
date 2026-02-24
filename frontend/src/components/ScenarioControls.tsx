import { useState, useEffect, useRef } from "react";
import { ScenarioConfig, ScenarioOutcome, WeeklyDirection, OneTimeDirection, SCENARIO_COLORS } from "../types";
import { generateScenarioTitle } from "../utils/scenarioTitle";
import { getNextScenarioColor, getNextScenarioId } from "../utils/scenarioHelpers";
import { Plus, X } from "lucide-react";

interface ScenarioControlsProps {
  scenarios: ScenarioConfig[];
  onChange: (scenarios: ScenarioConfig[]) => void;
  currentBalanceCents: number;
  hasAllowance: boolean;
  weeklyAllowanceCents: number;
  outcomes: Record<string, ScenarioOutcome>;
  horizonMonths: number;
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

function DirectionToggle<T extends string>({
  value,
  options,
  onChange,
  className,
}: {
  value: T;
  options: readonly { value: T; label: string }[];
  onChange: (value: T) => void;
  className?: string;
}) {
  return (
    <div className={`inline-flex items-center bg-cream-dark p-0.5 gap-0.5 ${className ?? "rounded-lg"}`}>
      {options.map((opt) => (
        <button
          key={opt.value}
          type="button"
          onClick={() => onChange(opt.value)}
          className={`
            px-2 py-1.5 rounded-md text-xs font-medium transition-colors cursor-pointer
            ${value === opt.value
              ? "bg-white text-bark shadow-sm"
              : "text-bark-light hover:text-bark"
            }
          `}
        >
          {opt.label}
        </button>
      ))}
    </div>
  );
}

const WEEKLY_OPTIONS = [
  { value: "spending" as const, label: "Spend" },
  { value: "saving" as const, label: "Save" },
] as const;

const ONE_TIME_OPTIONS = [
  { value: "deposit" as const, label: "Deposit" },
  { value: "withdrawal" as const, label: "Withdraw" },
] as const;

function renderBoldMarkdown(text: string) {
  const parts = text.split(/(\*\*[^*]+\*\*)/g);
  return parts.map((part, i) => {
    if (part.startsWith("**") && part.endsWith("**")) {
      return <strong key={i}>{part.slice(2, -2)}</strong>;
    }
    return part;
  });
}

function formatDollars(cents: number): string {
  return "$" + (Math.abs(cents) / 100).toFixed(2);
}

function horizonLabel(months: number): string {
  if (months >= 12 && months % 12 === 0) {
    const years = months / 12;
    return years === 1 ? "1 year" : `${years} years`;
  }
  return `${months} months`;
}

function buildOutcomeSuffix(outcome: ScenarioOutcome, horizonMonths: number): string {
  const period = horizonLabel(horizonMonths);

  if (outcome.depletionWeek !== null) {
    return `I'll run out of money in **${outcome.depletionWeek} weeks**.`;
  }

  if (outcome.finalBalanceCents === 0) {
    return `I'll have **$0.00** in ${period}.`;
  }

  const parts: string[] = [];
  if (outcome.totalAllowanceCents > 0) parts.push(`${formatDollars(outcome.totalAllowanceCents)} from allowance`);
  if (outcome.totalSavingsCents > 0) parts.push(`${formatDollars(outcome.totalSavingsCents)} from savings`);
  if (outcome.totalInterestCents > 0) parts.push(`${formatDollars(outcome.totalInterestCents)} from interest`);

  const totalOutflowCents = outcome.totalSpendingCents + outcome.oneTimeWithdrawalCents;

  let text = `I'll have **${formatDollars(outcome.finalBalanceCents)}** in ${period}`;
  if (parts.length > 0 || totalOutflowCents > 0) {
    const breakdown =
      parts.length === 1
        ? parts[0]
        : parts.length === 2
          ? `${parts[0]} and ${parts[1]}`
          : `${parts.slice(0, -1).join(", ")}, and ${parts[parts.length - 1]}`;
    text += ` — ${breakdown}`;
    if (totalOutflowCents > 0) {
      text += `${parts.length > 0 ? "," : ""} minus ${formatDollars(totalOutflowCents)} in spending`;
    }
    text += ".";
  } else {
    text += ".";
  }

  return text;
}

function ScenarioRow({
  scenario,
  onUpdate,
  onDelete,
  canDelete,
  currentBalanceCents,
  hasAllowance,
  weeklyAllowanceCents,
  outcome,
  horizonMonths,
}: {
  scenario: ScenarioConfig;
  onUpdate: (updated: ScenarioConfig) => void;
  onDelete: () => void;
  canDelete: boolean;
  currentBalanceCents: number;
  hasAllowance: boolean;
  weeklyAllowanceCents: number;
  outcome?: ScenarioOutcome;
  horizonMonths: number;
}) {
  const withdrawalExceedsBalance =
    scenario.oneTimeDirection === "withdrawal" &&
    scenario.oneTimeAmountCents > currentBalanceCents;

  const [localWeekly, setLocalWeekly] = useState(() =>
    centsToDollars(scenario.weeklyAmountCents)
  );
  const [localOneTime, setLocalOneTime] = useState(() =>
    centsToDollars(scenario.oneTimeAmountCents)
  );

  const prevCents = useRef({
    weeklyAmountCents: scenario.weeklyAmountCents,
    oneTimeAmountCents: scenario.oneTimeAmountCents,
  });

  useEffect(() => {
    if (scenario.weeklyAmountCents !== prevCents.current.weeklyAmountCents) {
      setLocalWeekly(centsToDollars(scenario.weeklyAmountCents));
    }
    if (scenario.oneTimeAmountCents !== prevCents.current.oneTimeAmountCents) {
      setLocalOneTime(centsToDollars(scenario.oneTimeAmountCents));
    }
    prevCents.current = {
      weeklyAmountCents: scenario.weeklyAmountCents,
      oneTimeAmountCents: scenario.oneTimeAmountCents,
    };
  }, [scenario.weeklyAmountCents, scenario.oneTimeAmountCents]);

  const handleWeeklyBlur = () => {
    const cents = dollarsToCents(localWeekly);
    onUpdate({ ...scenario, weeklyAmountCents: cents });
    setLocalWeekly(centsToDollars(cents));
    prevCents.current = { ...prevCents.current, weeklyAmountCents: cents };
  };

  const handleOneTimeBlur = () => {
    const cents = dollarsToCents(localOneTime);
    onUpdate({ ...scenario, oneTimeAmountCents: cents });
    setLocalOneTime(centsToDollars(cents));
    prevCents.current = { ...prevCents.current, oneTimeAmountCents: cents };
  };

  const title = generateScenarioTitle({
    hasAllowance,
    weeklyAllowanceCents,
    weeklyAmountCents: scenario.weeklyAmountCents,
    weeklyDirection: scenario.weeklyDirection,
    oneTimeAmountCents: scenario.oneTimeAmountCents,
    oneTimeDirection: scenario.oneTimeDirection,
  });

  return (
    <div className="flex items-start gap-3 border-b border-gray-200 pb-4">
      <span
        className="w-3 h-3 rounded-full flex-shrink-0 mt-1"
        style={{ backgroundColor: scenario.color }}
      />
      <div className="flex-1 space-y-2">
        <div className="flex items-start justify-between gap-2">
          <div className="text-base text-bark leading-snug">
            <p className="text-lg pb-2">{renderBoldMarkdown(title)},</p>
            {outcome && (
              <p className={
                outcome.finalBalanceCents >= currentBalanceCents
                  ? "text-[#2D5A3D] font-medium pb-1"
                  : "text-terracotta font-medium pb-1"
              }>
                {renderBoldMarkdown(buildOutcomeSuffix(outcome, horizonMonths))}
              </p>
            )}
          </div>
          {canDelete && (
            <button
              type="button"
              onClick={onDelete}
              className="flex-shrink-0 p-1 rounded-lg text-bark-light hover:text-terracotta hover:bg-terracotta/10 transition-colors cursor-pointer"
              aria-label="Remove scenario"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>
        <div className="flex flex-wrap items-center gap-x-6 gap-y-2">
          <div className="flex items-center gap-2">
            <div className="flex">
              <DirectionToggle<WeeklyDirection>
                value={scenario.weeklyDirection}
                options={WEEKLY_OPTIONS}
                onChange={(dir) => onUpdate({ ...scenario, weeklyDirection: dir })}
                className="h-9 rounded-l-lg rounded-r-none border border-r-0 border-sand"
              />
              <input
                id={`weekly-${scenario.id}`}
                type="number"
                min="0"
                step="0.01"
                placeholder="$0.00"
                value={localWeekly}
                onChange={(e) => setLocalWeekly(e.target.value)}
                onBlur={handleWeeklyBlur}
                className="w-24 h-9 px-3 py-1.5 rounded-r-lg rounded-l-none border border-sand bg-white text-bark text-sm placeholder:text-bark-light/50 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-forest/30 focus:border-forest"
              />
            </div>
            <span className="text-sm text-bark-light">weekly</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex">
              <DirectionToggle<OneTimeDirection>
                value={scenario.oneTimeDirection}
                options={ONE_TIME_OPTIONS}
                onChange={(dir) => onUpdate({ ...scenario, oneTimeDirection: dir })}
                className={`h-9 rounded-l-lg rounded-r-none border border-r-0 ${withdrawalExceedsBalance ? "border-terracotta" : "border-sand"}`}
              />
              <input
                id={`onetime-${scenario.id}`}
                type="number"
                min="0"
                step="0.01"
                placeholder="$0.00"
                value={localOneTime}
                onChange={(e) => setLocalOneTime(e.target.value)}
                onBlur={handleOneTimeBlur}
                className={`w-24 h-9 px-3 py-1.5 rounded-r-xl rounded-l-none border bg-white text-bark text-sm placeholder:text-bark-light/50 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-forest/30 focus:border-forest ${withdrawalExceedsBalance ? "border-terracotta ring-1 ring-terracotta/30" : "border-sand"}`}
              />
            </div>
            <span className="text-sm text-bark-light">one time</span>
          </div>
        </div>
        {withdrawalExceedsBalance && (
          <p className="text-sm text-terracotta font-medium">
            Can&apos;t withdraw more than ${(currentBalanceCents / 100).toFixed(2)} balance
          </p>
        )}

      </div>
    </div>
  );
}

export default function ScenarioControls({
  scenarios,
  onChange,
  currentBalanceCents,
  hasAllowance,
  weeklyAllowanceCents,
  outcomes,
  horizonMonths,
}: ScenarioControlsProps) {
  const handleUpdate = (updated: ScenarioConfig) => {
    onChange(scenarios.map((s) => (s.id === updated.id ? updated : s)));
  };

  const handleDelete = (id: string) => {
    onChange(scenarios.filter((s) => s.id !== id));
  };

  const handleAdd = () => {
    const newScenario: ScenarioConfig = {
      id: getNextScenarioId(scenarios),
      weeklyAmountCents: 0,
      weeklyDirection: "spending",
      oneTimeAmountCents: 0,
      oneTimeDirection: "deposit",
      color: getNextScenarioColor(scenarios),
    };
    onChange([...scenarios, newScenario]);
  };

  const canDelete = scenarios.length > 1;
  const canAdd = scenarios.length < SCENARIO_COLORS.length;

  return (
    <div className="space-y-4">
      <h3 className="text-sm font-bold text-bark uppercase tracking-wide">
        What if...
      </h3>
      <div className="space-y-5">
        {scenarios.map((s) => (
          <ScenarioRow
            key={s.id}
            scenario={s}
            onUpdate={handleUpdate}
            onDelete={() => handleDelete(s.id)}
            canDelete={canDelete}
            currentBalanceCents={currentBalanceCents}
            hasAllowance={hasAllowance}
            weeklyAllowanceCents={weeklyAllowanceCents}
            outcome={outcomes[s.id]}
            horizonMonths={horizonMonths}
          />
        ))}
      </div>
      {canAdd && (
        <button
          type="button"
          onClick={handleAdd}
          className="flex items-center gap-1.5 text-sm font-medium text-forest hover:text-forest/80 transition-colors cursor-pointer"
        >
          <Plus className="w-4 h-4" />
          Add scenario
        </button>
      )}
    </div>
  );
}
