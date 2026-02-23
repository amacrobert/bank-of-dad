import { useState, useEffect, useRef } from "react";
import { ScenarioConfig, WeeklyDirection, OneTimeDirection, SCENARIO_COLORS } from "../types";
import { generateScenarioTitle } from "../utils/scenarioTitle";
import { getNextScenarioColor, getNextScenarioId } from "../utils/scenarioHelpers";
import { Plus, X } from "lucide-react";

interface ScenarioControlsProps {
  scenarios: ScenarioConfig[];
  onChange: (scenarios: ScenarioConfig[]) => void;
  currentBalanceCents: number;
  hasAllowance: boolean;
  weeklyAllowanceCents: number;
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
}: {
  value: T;
  options: readonly { value: T; label: string }[];
  onChange: (value: T) => void;
}) {
  return (
    <div className="inline-flex rounded-lg bg-cream-dark p-0.5 gap-0.5">
      {options.map((opt) => (
        <button
          key={opt.value}
          type="button"
          onClick={() => onChange(opt.value)}
          className={`
            px-2 py-0.5 rounded-md text-xs font-medium transition-colors cursor-pointer
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

function ScenarioRow({
  scenario,
  onUpdate,
  onDelete,
  canDelete,
  currentBalanceCents,
  hasAllowance,
  weeklyAllowanceCents,
}: {
  scenario: ScenarioConfig;
  onUpdate: (updated: ScenarioConfig) => void;
  onDelete: () => void;
  canDelete: boolean;
  currentBalanceCents: number;
  hasAllowance: boolean;
  weeklyAllowanceCents: number;
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
    <div className="flex items-start gap-3">
      <span
        className="w-3 h-3 rounded-full flex-shrink-0 mt-1"
        style={{ backgroundColor: scenario.color }}
      />
      <div className="flex-1 space-y-2">
        <div className="flex items-start justify-between gap-2">
          <p className="text-sm text-bark leading-snug">
            {renderBoldMarkdown(title)}
          </p>
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
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <label htmlFor={`weekly-${scenario.id}`} className="text-sm font-semibold text-bark-light">
              Weekly
            </label>
            <DirectionToggle<WeeklyDirection>
              value={scenario.weeklyDirection}
              options={WEEKLY_OPTIONS}
              onChange={(dir) => onUpdate({ ...scenario, weeklyDirection: dir })}
            />
          </div>
          <input
            id={`weekly-${scenario.id}`}
            type="number"
            min="0"
            step="0.01"
            placeholder="$0.00"
            value={localWeekly}
            onChange={(e) => setLocalWeekly(e.target.value)}
            onBlur={handleWeeklyBlur}
            className="w-full min-h-[48px] px-4 py-3 rounded-xl border border-sand bg-white text-bark text-base placeholder:text-bark-light/50 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-forest/30 focus:border-forest"
          />
        </div>
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <label htmlFor={`onetime-${scenario.id}`} className="text-sm font-semibold text-bark-light">
              One time
            </label>
            <DirectionToggle<OneTimeDirection>
              value={scenario.oneTimeDirection}
              options={ONE_TIME_OPTIONS}
              onChange={(dir) => onUpdate({ ...scenario, oneTimeDirection: dir })}
            />
          </div>
          <input
            id={`onetime-${scenario.id}`}
            type="number"
            min="0"
            step="0.01"
            placeholder="$0.00"
            value={localOneTime}
            onChange={(e) => setLocalOneTime(e.target.value)}
            onBlur={handleOneTimeBlur}
            className={`w-full min-h-[48px] px-4 py-3 rounded-xl border bg-white text-bark text-base placeholder:text-bark-light/50 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-forest/30 focus:border-forest ${withdrawalExceedsBalance ? "border-terracotta ring-1 ring-terracotta/30" : "border-sand"}`}
          />
          {withdrawalExceedsBalance && (
            <p className="text-sm text-terracotta font-medium">
              Can&apos;t withdraw more than ${(currentBalanceCents / 100).toFixed(2)} balance
            </p>
          )}
        </div>
        </div>
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
      <div className="space-y-4">
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
