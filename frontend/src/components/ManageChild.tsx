import { useState, useEffect, useCallback } from "react";
import { put, getBalance, getTransactions, getChildAllowance, getInterestSchedule, ApiRequestError } from "../api";
import { Child, Transaction, AllowanceSchedule, InterestSchedule } from "../types";
import Card from "./ui/Card";
import Input from "./ui/Input";
import Button from "./ui/Button";
import BalanceDisplay from "./BalanceDisplay";
import DepositForm from "./DepositForm";
import WithdrawForm from "./WithdrawForm";
import InterestForm from "./InterestForm";
import TransactionHistory from "./TransactionHistory";
import UpcomingPayments from "./UpcomingPayments";
import ChildAllowanceForm from "./ChildAllowanceForm";
import { AlertTriangle, X, ArrowDownCircle, ArrowUpCircle } from "lucide-react";

interface ManageChildProps {
  child: Child;
  onUpdated: () => void;
  onClose: () => void;
}

export default function ManageChild({ child, onUpdated, onClose }: ManageChildProps) {
  const [newPassword, setNewPassword] = useState("");
  const [newName, setNewName] = useState(child.first_name);
  const [passwordMsg, setPasswordMsg] = useState<string | null>(null);
  const [nameMsg, setNameMsg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
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

  const handleResetPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setPasswordMsg(null);

    try {
      const result = await put<{ message: string; account_unlocked: boolean }>(
        `/children/${child.id}/password`,
        { password: newPassword }
      );
      setPasswordMsg(
        result.account_unlocked
          ? "Password updated and account unlocked."
          : "Password updated."
      );
      setNewPassword("");
      onUpdated();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to reset password.");
      }
    }
  };

  const handleUpdateName = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setNameMsg(null);

    try {
      const result = await put<{ message: string; first_name: string }>(
        `/children/${child.id}/name`,
        { first_name: newName }
      );
      setNameMsg(`Name updated to ${result.first_name}.`);
      onUpdated();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to update name.");
      }
    }
  };

  return (
    <div className="space-y-4">
      {/* Header with close */}
      <div className="flex items-center justify-between">
        <h3 className="text-xl font-bold text-forest">Manage {child.first_name}</h3>
        <button
          onClick={onClose}
          className="p-2 rounded-xl text-bark-light hover:text-bark hover:bg-cream-dark transition-colors cursor-pointer"
          aria-label="Close"
        >
          <X className="h-5 w-5" />
        </button>
      </div>

      {child.is_locked && (
        <div className="flex items-center gap-2 p-3 bg-terracotta/10 border border-terracotta/20 rounded-xl">
          <AlertTriangle className="h-5 w-5 text-terracotta flex-shrink-0" aria-hidden="true" />
          <p className="text-sm text-terracotta font-medium">This account is locked. Reset the password to unlock it.</p>
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

      <UpcomingPayments childId={child.id} balanceCents={currentBalance} interestRateBps={interestRateBps} />

      {/* Transaction history */}
      <Card padding="md">
        <h4 className="text-base font-bold text-bark mb-3">Transaction History</h4>
        <TransactionHistory transactions={transactions} />
      </Card>

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

      {/* Reset password */}
      <Card padding="md">
        <h4 className="text-base font-bold text-bark mb-4">Reset Password</h4>
        <form onSubmit={handleResetPassword} className="space-y-4">
          <Input
            label="New Password (min 6 characters)"
            id="new-password"
            type="text"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            minLength={6}
            required
          />
          <Button type="submit" className="w-full">Reset Password</Button>
          {passwordMsg && (
            <div className="bg-forest/5 border border-forest/15 rounded-xl p-3">
              <p className="text-sm text-forest font-medium">{passwordMsg}</p>
            </div>
          )}
        </form>
      </Card>

      {/* Update name */}
      <Card padding="md">
        <h4 className="text-base font-bold text-bark mb-4">Update Name</h4>
        <form onSubmit={handleUpdateName} className="space-y-4">
          <Input
            label="First Name"
            id="new-name"
            type="text"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            required
          />
          <Button type="submit" className="w-full">Update Name</Button>
          {nameMsg && (
            <div className="bg-forest/5 border border-forest/15 rounded-xl p-3">
              <p className="text-sm text-forest font-medium">{nameMsg}</p>
            </div>
          )}
        </form>
      </Card>

      {error && (
        <div className="bg-terracotta/10 border border-terracotta/20 rounded-xl p-3">
          <p className="text-sm text-terracotta font-medium">{error}</p>
        </div>
      )}
    </div>
  );
}
