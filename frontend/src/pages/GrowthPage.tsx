import { useEffect, useState, useMemo } from "react";
import { useOutletContext, useNavigate, useParams } from "react-router-dom";
import { get, getBalance, getChildAllowance, getInterestSchedule } from "../api";
import {
  AuthUser,
  Child,
  ChildListResponse,
  BalanceResponse,
  AllowanceSchedule,
  InterestSchedule,
  ScenarioInputs,
  ProjectionConfig,
} from "../types";
import { calculateProjection, weeksPerPeriod } from "../utils/projection";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import ChildSelectorBar from "../components/ChildSelectorBar";
import GrowthChart from "../components/GrowthChart";
import ScenarioControls from "../components/ScenarioControls";
import GrowthExplanation from "../components/GrowthExplanation";
import { TrendingUp, Users } from "lucide-react";

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
  const { user } = useOutletContext<{ user: AuthUser }>();
  const navigate = useNavigate();
  const { childName } = useParams<{ childName?: string }>();
  const isParent = user.user_type === "parent";

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Data from API
  const [balanceData, setBalanceData] = useState<BalanceResponse | null>(null);
  const [allowance, setAllowance] = useState<AllowanceSchedule | null>(null);
  const [interestSchedule, setInterestSchedule] = useState<InterestSchedule | null>(null);

  // Scenario state
  const [scenario, setScenario] = useState<ScenarioInputs>(DEFAULT_SCENARIO);

  // Parent-mode state
  const [children, setChildren] = useState<Child[]>([]);
  const [childrenLoading, setChildrenLoading] = useState(true);

  // Derive selected child from URL param
  const selectedChild = useMemo(() => {
    if (!isParent || !childName || children.length === 0) return null;
    return children.find(
      (c) => c.first_name.toLowerCase() === childName.toLowerCase()
    ) ?? null;
  }, [isParent, childName, children]);

  // Redirect if child name in URL is invalid
  useEffect(() => {
    if (isParent && childName && children.length > 0 && !selectedChild) {
      navigate("/growth", { replace: true });
    }
  }, [isParent, childName, children, selectedChild, navigate]);

  // Fetch child list for parent mode
  useEffect(() => {
    if (!isParent) return;
    get<ChildListResponse>("/children")
      .then((data) => {
        setChildren(data.children || []);
      })
      .catch(() => {})
      .finally(() => setChildrenLoading(false));
  }, [isParent]);

  // Derive childId based on user type
  const childId: number | null = isParent ? selectedChild?.id ?? null : user.user_id;

  // Fetch child data
  useEffect(() => {
    // Always reset state on childId change
    setBalanceData(null);
    setAllowance(null);
    setInterestSchedule(null);
    setScenario(DEFAULT_SCENARIO);
    setError("");
    setLoading(true);

    if (childId === null) {
      setLoading(false);
      return;
    }

    Promise.all([
      getBalance(childId),
      getChildAllowance(childId).catch(() => null),
      getInterestSchedule(childId).catch(() => null),
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
  }, [childId]);

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

  // Parent: render selector bar first (always visible), then handle empty/no-selection states
  if (isParent) {
    // No children empty state
    if (children.length === 0 && !childrenLoading) {
      return (
        <div className="max-w-[960px] mx-auto animate-fade-in-up space-y-6">
          <div className="flex items-center gap-3">
            <TrendingUp className="h-6 w-6 text-forest" aria-hidden="true" />
            <h1 className="text-2xl font-bold text-bark">Growth Projector</h1>
          </div>
          <Card padding="lg">
            <div className="text-center py-8">
              <div className="flex justify-center mb-4">
                <div className="w-14 h-14 bg-forest/10 rounded-2xl flex items-center justify-center">
                  <Users className="h-7 w-7 text-forest" aria-hidden="true" />
                </div>
              </div>
              <h3 className="text-lg font-bold text-bark mb-2">No children yet</h3>
              <p className="text-bark-light mb-4">
                Add your first child in Settings to view their growth projector.
              </p>
              <Button onClick={() => navigate("/settings/children")}>
                Go to Settings &rarr; Children
              </Button>
            </div>
          </Card>
        </div>
      );
    }

    return (
      <div className="max-w-[960px] mx-auto animate-fade-in-up space-y-6">
        <div className="flex items-center gap-3">
          <TrendingUp className="h-6 w-6 text-forest" aria-hidden="true" />
          <h1 className="text-2xl font-bold text-bark">Growth Projector</h1>
        </div>

        <ChildSelectorBar
          children={children}
          selectedChildId={selectedChild?.id ?? null}
          onSelectChild={(child) => {
            if (child) {
              navigate(`/growth/${child.first_name.toLowerCase()}`);
            } else {
              navigate("/growth");
            }
          }}
          loading={childrenLoading}
        />

        {!selectedChild && !childrenLoading && (
          <Card padding="lg">
            <p className="text-bark-light text-center py-4">
              Select a child to view their growth projector.
            </p>
          </Card>
        )}

        {selectedChild && loading && (
          <div className="flex items-center justify-center min-h-[300px]">
            <LoadingSpinner message="Loading growth projector..." />
          </div>
        )}

        {selectedChild && error && (
          <Card>
            <p className="text-terracotta font-medium text-center py-8">{error}</p>
          </Card>
        )}

        {selectedChild && !loading && !error && (
          <ProjectorContent
            scenario={scenario}
            setScenario={setScenario}
            projection={projection}
            balanceData={balanceData}
            allowance={allowance}
            interestSchedule={interestSchedule}
          />
        )}
      </div>
    );
  }

  // Child mode: original behavior
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
      <div className="flex items-center gap-3">
        <TrendingUp className="h-6 w-6 text-forest" aria-hidden="true" />
        <h1 className="text-2xl font-bold text-bark">Growth Projector</h1>
      </div>
      <ProjectorContent
        scenario={scenario}
        setScenario={setScenario}
        projection={projection}
        balanceData={balanceData}
        allowance={allowance}
        interestSchedule={interestSchedule}
      />
    </div>
  );
}

interface ProjectorContentProps {
  scenario: ScenarioInputs;
  setScenario: React.Dispatch<React.SetStateAction<ScenarioInputs>>;
  projection: ReturnType<typeof calculateProjection> | null;
  balanceData: BalanceResponse | null;
  allowance: AllowanceSchedule | null;
  interestSchedule: InterestSchedule | null;
}

function ProjectorContent({
  scenario,
  setScenario,
  projection,
  balanceData,
  allowance,
  interestSchedule,
}: ProjectorContentProps) {
  return (
    <>
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
          <GrowthChart dataPoints={projection.dataPoints} animationKey={JSON.stringify(scenario)} />
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
    </>
  );
}
