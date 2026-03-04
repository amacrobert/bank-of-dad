import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getBalance, getTransactions, getSavingsGoals } from "../api";
import { Transaction, SavingsGoal } from "../types";
import { useChildUser } from "../hooks/useAuthOutletContext";
import Card from "../components/ui/Card";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import BalanceDisplay from "../components/BalanceDisplay";
import TransactionsCard from "../components/TransactionsCard";
import GoalProgressRing from "../components/GoalProgressRing";
import { TrendingUp, Target } from "lucide-react";

export default function ChildDashboard() {
  const user = useChildUser();
  const [balance, setBalance] = useState<number>(0);
  const [interestRateBps, setInterestRateBps] = useState<number>(0);
  const [interestRateDisplay, setInterestRateDisplay] = useState<string>("");
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loadingData, setLoadingData] = useState(true);
  const [availableBalanceCents, setAvailableBalanceCents] = useState<number | undefined>();
  const [totalSavedCents, setTotalSavedCents] = useState<number | undefined>();
  const [activeGoals, setActiveGoals] = useState<SavingsGoal[]>([]);

  useEffect(() => {
    Promise.all([
      getBalance(user.user_id),
      getTransactions(user.user_id),
      getSavingsGoals(user.user_id).catch(() => ({ goals: [], available_balance_cents: 0, total_saved_cents: 0 })),
    ]).then(([balanceRes, txRes, goalsRes]) => {
      setBalance(balanceRes.balance_cents);
      setInterestRateBps(balanceRes.interest_rate_bps);
      setInterestRateDisplay(balanceRes.interest_rate_display);
      setAvailableBalanceCents(balanceRes.available_balance_cents);
      setTotalSavedCents(balanceRes.total_saved_cents);
      setTransactions(txRes.transactions || []);
      setActiveGoals(goalsRes.goals.filter((g: SavingsGoal) => g.status === "active").slice(0, 3));
    }).catch(() => {
      // Silently fail
    }).finally(() => {
      setLoadingData(false);
    });
  }, [user.user_id]);

  return (
    <div className="max-w-[480px] mx-auto space-y-6 animate-fade-in-up">
      {/* Welcome */}
      <div>
        <h2 className="text-2xl font-bold text-forest">
          Welcome, {user.first_name}!{user.avatar ? ` ${user.avatar}` : ''}
        </h2>
      </div>

      {/* Hero balance card */}
      <Card padding="lg" className="text-center">
        <p className="text-sm font-semibold text-bark-light uppercase tracking-wide mb-2">
          Your Balance
        </p>
        {loadingData ? (
          <LoadingSpinner variant="inline" />
        ) : (
          <BalanceDisplay
            balanceCents={balance}
            size="large"
            breakdown={totalSavedCents && totalSavedCents > 0 && availableBalanceCents !== undefined
              ? { availableCents: availableBalanceCents, savedCents: totalSavedCents }
              : undefined
            }
          />
        )}
        {interestRateBps > 0 && !loadingData && (
          <div className="mt-4 flex justify-center items-center gap-1.5 bg-sage-light/30 text-forest text-sm font-medium px-3 py-1.5 rounded-full">
            <TrendingUp className="h-4 w-4" aria-hidden="true" />
            {interestRateDisplay} annual interest
          </div>
        )}
      </Card>

      {/* Savings Goals section */}
      {!loadingData && (
        <Card padding="sm" className="space-y-3">
          <Link to="/child/goals" className="flex items-center gap-3 hover:opacity-80 transition-opacity">
            <div className="p-2 bg-sage-light/30 rounded-lg">
              <Target className="h-5 w-5 text-forest" />
            </div>
            <div className="flex-1">
              <p className="font-semibold text-bark">Savings Goals</p>
              <p className="text-xs text-bark-light">
                {activeGoals.length > 0 ? `${activeGoals.length} active goal${activeGoals.length > 1 ? "s" : ""}` : "Set targets and track your progress"}
              </p>
            </div>
            <span className="text-bark-light text-sm">&rarr;</span>
          </Link>
          {activeGoals.length > 0 && (
            <div className="flex gap-3 overflow-x-auto scrollbar-hide pb-1">
              {activeGoals.map((goal) => {
                const pct = goal.target_cents > 0 ? Math.min(100, Math.round((goal.saved_cents / goal.target_cents) * 100)) : 0;
                return (
                  <Link key={goal.id} to="/child/goals" className="flex-shrink-0 flex items-center gap-2 px-3 py-2 bg-cream/50 rounded-lg hover:bg-cream transition-colors">
                    <span className="text-base">{goal.emoji || "🎯"}</span>
                    <span className="text-xs font-medium text-bark truncate max-w-[80px]">{goal.name}</span>
                    <GoalProgressRing percent={pct} size={32} strokeWidth={3} milestone={pct >= 75} />
                  </Link>
                );
              })}
            </div>
          )}
        </Card>
      )}

      {/* Transactions */}
      {!loadingData && (
        <TransactionsCard
          childId={user.user_id}
          balanceCents={balance}
          interestRateBps={interestRateBps}
          transactions={transactions}
        />
      )}
    </div>
  );
}
