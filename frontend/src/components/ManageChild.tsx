import { useState, useEffect, useCallback } from "react";
import { getBalance, getTransactions, getChildAllowance, getInterestSchedule } from "../api";
import { Child, Transaction, AllowanceSchedule, InterestSchedule } from "../types";
import Card from "./ui/Card";
import Button from "./ui/Button";
import BalanceDisplay from "./BalanceDisplay";
import DepositForm from "./DepositForm";
import WithdrawForm from "./WithdrawForm";
import InterestForm from "./InterestForm";
import TransactionsCard from "./TransactionsCard";
import ChildAllowanceForm from "./ChildAllowanceForm";
import { AlertTriangle, ArrowDownCircle, ArrowUpCircle } from "lucide-react";

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

  const loadTransactions = useCallback(() => {
    getTransactions(child.id).then((data) => {
      setTransactions(data.transactions || []);
    }).catch(() => {});
  }, [child.id]);

  useEffect(() => {
    getBalance(child.id).then((data) => {
      setInterestRateBps(data.interest_rate_bps);
    }).catch(() => {});
    loadTransactions();
    getChildAllowance(child.id).then(setAllowance).catch(() => {});
    getInterestSchedule(child.id).then(setInterestSchedule).catch(() => {});
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
          <BalanceDisplay balanceCents={currentBalance} size="large" />
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

      {showDeposit && (
        <DepositForm
          child={childWithCurrentBalance}
          onSuccess={handleDepositSuccess}
          onCancel={() => setShowDeposit(false)}
        />
      )}

      {showWithdraw && (
        <WithdrawForm
          child={childWithCurrentBalance}
          onSuccess={handleWithdrawSuccess}
          onCancel={() => setShowWithdraw(false)}
        />
      )}

      <TransactionsCard childId={child.id} balanceCents={currentBalance} interestRateBps={interestRateBps} transactions={transactions} />

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
  );
}
