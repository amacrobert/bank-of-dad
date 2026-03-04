import { useState, useEffect, useCallback } from "react";
import { getBalance, getTransactions, getChildAllowance, getInterestSchedule, getSavingsGoals } from "../api";
import { Child, Transaction, AllowanceSchedule, InterestSchedule, SavingsGoal } from "../types";
import Card from "./ui/Card";
import Button from "./ui/Button";
import Modal from "./ui/Modal";
import BalanceDisplay from "./BalanceDisplay";
import DepositForm from "./DepositForm";
import WithdrawForm from "./WithdrawForm";
import InterestForm from "./InterestForm";
import TransactionsCard from "./TransactionsCard";
import ChildAllowanceForm from "./ChildAllowanceForm";
import GoalCard from "./GoalCard";
import { AlertTriangle, ArrowDownCircle, ArrowUpCircle, Target, CheckCircle2 } from "lucide-react";

interface ManageChildProps {
  child: Child;
  onUpdated: () => void;
}

export default function ManageChild({ child, onUpdated }: ManageChildProps) {
  const [showDeposit, setShowDeposit] = useState(false);
  const [showWithdraw, setShowWithdraw] = useState(false);
  const [currentBalance, setCurrentBalance] = useState(child.balance_cents);
  const [interestRateBps, setInterestRateBps] = useState(0);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [allowance, setAllowance] = useState<AllowanceSchedule | null>(null);
  const [interestSchedule, setInterestSchedule] = useState<InterestSchedule | null>(null);
  const [savingsGoals, setSavingsGoals] = useState<SavingsGoal[]>([]);
  const [availableBalanceCents, setAvailableBalanceCents] = useState<number | undefined>(undefined);
  const [totalSavedCents, setTotalSavedCents] = useState<number>(0);

  const loadTransactions = useCallback(() => {
    getTransactions(child.id).then((data) => {
      setTransactions(data.transactions || []);
    }).catch(() => {});
  }, [child.id]);

  useEffect(() => {
    getBalance(child.id).then((data) => {
      setInterestRateBps(data.interest_rate_bps);
      setAvailableBalanceCents(data.available_balance_cents);
      setTotalSavedCents(data.total_saved_cents || 0);
    }).catch(() => {});
    loadTransactions();
    getChildAllowance(child.id).then(setAllowance).catch(() => {});
    getInterestSchedule(child.id).then(setInterestSchedule).catch(() => {});
    getSavingsGoals(child.id).then((res) => setSavingsGoals(res.goals)).catch(() => {});
  }, [child.id, loadTransactions]);

  const handleDepositSuccess = (newBalance: number) => {
    setCurrentBalance(newBalance);
    setShowDeposit(false);
    loadTransactions();
    onUpdated();
  };

  const handleWithdrawSuccess = (newBalance: number) => {
    setCurrentBalance(newBalance);
    setShowWithdraw(false);
    loadTransactions();
    onUpdated();
  };

  const childWithCurrentBalance = { ...child, balance_cents: currentBalance };

  return (
    <div className="animate-fade-in-up space-y-4">
      <h3 className="text-xl font-bold text-forest">Manage {child.first_name}</h3>

      {child.is_locked && (
        <div className="flex items-center gap-2 p-3 bg-terracotta/10 border border-terracotta/20 rounded-xl">
          <AlertTriangle className="h-5 w-5 text-terracotta flex-shrink-0" aria-hidden="true" />
          <p className="text-sm text-terracotta font-medium">This account is locked. Go to Settings &rarr; Children to reset the password.</p>
        </div>
      )}

      {/* Balance + actions */}
      <Card padding="md">
        <div className="text-center mb-4">
          <p className="text-sm font-semibold text-bark-light uppercase tracking-wide mb-1">Balance</p>
          <BalanceDisplay
            balanceCents={currentBalance}
            size="large"
            breakdown={totalSavedCents > 0 && availableBalanceCents !== undefined
              ? { availableCents: availableBalanceCents, savedCents: totalSavedCents }
              : undefined
            }
          />
        </div>
        <div className="flex gap-3">
          <Button
            variant="primary"
            onClick={() => { setShowDeposit(true); setShowWithdraw(false); }}
            className="flex-1"
          >
            <ArrowDownCircle className="h-4 w-4" aria-hidden="true" />
            Deposit
          </Button>
          <Button
            variant="secondary"
            onClick={() => { setShowWithdraw(true); setShowDeposit(false); }}
            disabled={currentBalance === 0}
            className="flex-1"
          >
            <ArrowUpCircle className="h-4 w-4" aria-hidden="true" />
            Withdraw
          </Button>
        </div>
      </Card>

      <Modal open={showDeposit} onClose={() => setShowDeposit(false)}>
        <DepositForm
          child={childWithCurrentBalance}
          onSuccess={handleDepositSuccess}
          onCancel={() => setShowDeposit(false)}
        />
      </Modal>

      <Modal open={showWithdraw} onClose={() => setShowWithdraw(false)}>
        <WithdrawForm
          child={childWithCurrentBalance}
          onSuccess={handleWithdrawSuccess}
          onCancel={() => setShowWithdraw(false)}
        />
      </Modal>

      <TransactionsCard childId={child.id} balanceCents={currentBalance} interestRateBps={interestRateBps} transactions={transactions} />

      {/* Savings Goals (read-only for parent) */}
      {savingsGoals.length > 0 && (
        <Card padding="md">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Target className="h-5 w-5 text-forest" />
              <h4 className="font-semibold text-bark">Savings Goals</h4>
            </div>
            {savingsGoals.filter((g) => g.status === "completed").length > 0 && (
              <div className="flex items-center gap-1 text-xs text-forest">
                <CheckCircle2 className="h-3.5 w-3.5" />
                <span>{savingsGoals.filter((g) => g.status === "completed").length} achieved</span>
              </div>
            )}
          </div>
          <div className="space-y-2">
            {savingsGoals.filter((g) => g.status === "active").length > 0 ? (
              savingsGoals
                .filter((g) => g.status === "active")
                .map((goal) => (
                  <GoalCard key={goal.id} goal={goal} childId={child.id} />
                ))
            ) : (
              <p className="text-sm text-bark-light">No active savings goals.</p>
            )}
          </div>
        </Card>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <ChildAllowanceForm
          childId={child.id}
          childName={child.first_name}
          allowance={allowance}
          onUpdated={setAllowance}
        />

        <InterestForm
          childId={child.id}
          childName={child.first_name}
          currentRateBps={interestRateBps}
          schedule={interestSchedule}
          onUpdated={(rateBps, schedule) => {
            setInterestRateBps(rateBps);
            setInterestSchedule(schedule);
          }}
        />
      </div>
    </div>
  );
}
