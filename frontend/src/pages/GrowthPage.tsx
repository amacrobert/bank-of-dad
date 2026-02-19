import { useEffect, useState, useMemo } from "react";
import { getBalance, getChildAllowance, getInterestSchedule } from "../api";
import {
  BalanceResponse,
  AllowanceSchedule,
  InterestSchedule,
  ScenarioInputs,
  ProjectionConfig,
} from "../types";
import { calculateProjection, weeksPerPeriod } from "../utils/projection";
import { useChildUser } from "../hooks/useAuthOutletContext";
import Card from "../components/ui/Card";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import GrowthChart from "../components/GrowthChart";
import ScenarioControls from "../components/ScenarioControls";
import GrowthExplanation from "../components/GrowthExplanation";
import { TrendingUp } from "lucide-react";

const HORIZON_OPTIONS = [
  { months: 3, label: "3mo" },
  { months: 6, label: "6mo" },
  { months: 12, label: "1yr" },
  { months: 24, label: "2yr" },
  { months: 60, label: "5yr" },
] as const;

const DEFAULT_SCENARIO: ScenarioInputs = {
  weeklySpendingCents: 0,
  oneTimeDepositCents: 0,
  oneTimeWithdrawalCents: 0,
  horizonMonths: 12,
};

export default function GrowthPage() {
  const user = useChildUser();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Data from API
  const [balanceData, setBalanceData] = useState<BalanceResponse | null>(null);
  const [allowance, setAllowance] = useState<AllowanceSchedule | null>(null);
  const [interestSchedule, setInterestSchedule] = useState<InterestSchedule | null>(null);

  // Scenario state
  const [scenario, setScenario] = useState<ScenarioInputs>(DEFAULT_SCENARIO);

  useEffect(() => {
    Promise.all([
      getBalance(user.user_id),
      getChildAllowance(user.user_id).catch(() => null),
      getInterestSchedule(user.user_id).catch(() => null),
    ])
      .then(([bal, allow, interest]) => {
        setBalanceData(bal);
        setAllowance(allow);
        setInterestSchedule(interest);

        // Pre-populate spending to half the allowance (converted to weekly)
        if (allow?.status === "active" && allow.amount_cents > 0) {
          const weeklyEquivalent = allow.amount_cents / weeksPerPeriod(allow.frequency);
          setScenario((s) => ({
            ...s,
            weeklySpendingCents: Math.round(weeklyEquivalent / 2),
          }));
        }
      })
      .catch(() => {
        setError("Failed to load account data.");
      })
      .finally(() => {
        setLoading(false);
      });
  }, [user.user_id]);

  // Build projection config from fetched data + scenario
  const projectionConfig: ProjectionConfig | null = useMemo(() => {
    if (!balanceData) return null;

    const isAllowanceActive = allowance?.status === "active";
    const isInterestActive = interestSchedule?.status === "active";

    return {
      currentBalanceCents: balanceData.balance_cents,
      interestRateBps: balanceData.interest_rate_bps,
      interestFrequency: isInterestActive ? interestSchedule!.frequency : null,
      allowanceAmountCents: isAllowanceActive ? allowance!.amount_cents : 0,
      allowanceFrequency: isAllowanceActive ? allowance!.frequency : null,
      scenario,
    };
  }, [balanceData, allowance, interestSchedule, scenario]);

  // Calculate projection
  const projection = useMemo(() => {
    if (!projectionConfig) return null;
    return calculateProjection(projectionConfig);
  }, [projectionConfig]);

  if (loading) {
    return (
      <div className="max-w-[960px] mx-auto flex items-center justify-center min-h-[300px]">
        <LoadingSpinner message="Loading growth projector..." />
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-[960px] mx-auto animate-fade-in-up">
        <Card>
          <p className="text-terracotta font-medium text-center py-8">{error}</p>
        </Card>
      </div>
    );
  }

  return (
    <div className="max-w-[960px] mx-auto animate-fade-in-up space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <TrendingUp className="h-6 w-6 text-forest" aria-hidden="true" />
        <h1 className="text-2xl font-bold text-bark">Growth Projector</h1>
      </div>

      {/* Time Horizon Selector + Chart */}
      <Card>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-sm font-bold text-bark uppercase tracking-wide">
            Projected Balance
          </h2>
          <div className="flex gap-1">
            {HORIZON_OPTIONS.map((opt) => (
              <button
                key={opt.months}
                onClick={() => setScenario((s) => ({ ...s, horizonMonths: opt.months }))}
                className={`
                  px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors cursor-pointer
                  ${scenario.horizonMonths === opt.months
                    ? "bg-forest text-white"
                    : "text-bark-light hover:bg-cream-dark"
                  }
                `}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>
        {projection && (
          <GrowthChart dataPoints={projection.dataPoints} />
        )}
      </Card>

      {/* Scenario Controls */}
      {balanceData && (
        <Card>
          <ScenarioControls
            scenario={scenario}
            onChange={setScenario}
            currentBalanceCents={balanceData.balance_cents}
          />
        </Card>
      )}

      {/* Plain-English Explanation */}
      {projection && balanceData && (
        <Card>
          <GrowthExplanation
            projection={projection}
            horizonMonths={scenario.horizonMonths}
            hasAllowance={allowance?.status === "active" && (allowance?.amount_cents ?? 0) > 0}
            hasInterest={interestSchedule?.status === "active" && balanceData.interest_rate_bps > 0}
            isAllowancePaused={allowance?.status === "paused"}
            isInterestPaused={interestSchedule?.status === "paused"}
            weeklyAllowanceCentsDisplay={allowance?.amount_cents ?? 0}
            allowanceFrequencyDisplay={allowance?.frequency ?? "weekly"}
            weeklySpendingCents={scenario.weeklySpendingCents}
          />
        </Card>
      )}
    </div>
  );
}
