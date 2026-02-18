import { useState, useEffect, useCallback, useRef } from "react";
import { put, deleteChild, getBalance, getTransactions, getChildAllowance, getInterestSchedule, ApiRequestError } from "../api";
import { Child, Transaction, AllowanceSchedule, InterestSchedule } from "../types";
import Card from "./ui/Card";
import Input from "./ui/Input";
import Button from "./ui/Button";
import BalanceDisplay from "./BalanceDisplay";
import DepositForm from "./DepositForm";
import WithdrawForm from "./WithdrawForm";
import InterestForm from "./InterestForm";
import TransactionsCard from "./TransactionsCard";
import ChildAllowanceForm from "./ChildAllowanceForm";
import AvatarPicker from "./AvatarPicker";
import { AlertTriangle, Trash2, X, ArrowDownCircle, ArrowUpCircle, ChevronDown } from "lucide-react";

interface ManageChildProps {
  child: Child;
  onUpdated: () => void;
  onClose: () => void;
}

export default function ManageChild({ child, onUpdated, onClose }: ManageChildProps) {
  const [newPassword, setNewPassword] = useState("");
  const [newName, setNewName] = useState(child.first_name);
  const [newAvatar, setNewAvatar] = useState<string | null>(child.avatar ?? null);
  const [passwordMsg, setPasswordMsg] = useState<string | null>(null);
  const [nameMsg, setNameMsg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [showSettings, setShowSettings] = useState(false);
  const settingsRef = useRef<HTMLButtonElement>(null);
  const [showDeposit, setShowDeposit] = useState(false);
  const [showWithdraw, setShowWithdraw] = useState(false);
  const [currentBalance, setCurrentBalance] = useState(child.balance_cents);
  const [interestRateBps, setInterestRateBps] = useState(0);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [allowance, setAllowance] = useState<AllowanceSchedule | null>(null);
  const [interestSchedule, setInterestSchedule] = useState<InterestSchedule | null>(null);
  const [deleteConfirmName, setDeleteConfirmName] = useState("");
  const [deleting, setDeleting] = useState(false);

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

  const nameMatches = deleteConfirmName.toLowerCase() === child.first_name.toLowerCase();

  const handleDeleteChild = async () => {
    if (!nameMatches) return;
    setDeleting(true);
    setError(null);
    try {
      await deleteChild(child.id);
      onClose();
      onUpdated();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to delete account.");
      }
      setDeleting(false);
    }
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
      const result = await put<{ message: string; first_name: string; avatar: string | null }>(
        `/children/${child.id}/name`,
        { first_name: newName, avatar: newAvatar }
      );
      setNameMsg(`Updated successfully.`);
      setNewAvatar(result.avatar ?? null);
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
    <div className="animate-fade-in-up space-y-4">
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

      {/* Account Settings (collapsible) */}
      <button
        ref={settingsRef}
        onClick={() => {
          const expanding = !showSettings;
          setShowSettings(expanding);
          if (expanding) {
            setTimeout(() => settingsRef.current?.scrollIntoView({ behavior: "smooth", block: "start" }), 0);
          }
        }}
        className="w-full flex items-center justify-between p-3 rounded-xl bg-cream hover:bg-cream-dark transition-colors cursor-pointer"
      >
        <span className="text-base font-bold text-bark">Account Settings</span>
        <ChevronDown
          className="h-5 w-5 text-bark-light transition-transform"
          style={{ transform: showSettings ? "rotate(180deg)" : undefined }}
          aria-hidden="true"
        />
      </button>

      {showSettings && (
        <>
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

          {/* Update name and avatar */}
          <Card padding="md">
            <h4 className="text-base font-bold text-bark mb-4">Update Name and Avatar</h4>
            <form onSubmit={handleUpdateName} className="space-y-4">
              <Input
                label="First Name"
                id="new-name"
                type="text"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                required
              />
              <AvatarPicker selected={newAvatar} onSelect={setNewAvatar} />
              <Button type="submit" className="w-full">Update</Button>
              {nameMsg && (
                <div className="bg-forest/5 border border-forest/15 rounded-xl p-3">
                  <p className="text-sm text-forest font-medium">{nameMsg}</p>
                </div>
              )}
            </form>
          </Card>

          {/* Delete account */}
          <Card padding="md">
            <div className="flex items-center gap-2 mb-3">
              <Trash2 className="h-5 w-5 text-terracotta" aria-hidden="true" />
              <h4 className="text-base font-bold text-terracotta">Delete Account</h4>
            </div>
            <p className="text-sm text-bark-light mb-4">
              This action is permanent and cannot be undone. All of {child.first_name}&apos;s transactions, schedules, and account data will be permanently deleted.
            </p>
            <Input
              label={`Type "${child.first_name}" to confirm`}
              id="delete-confirm"
              type="text"
              value={deleteConfirmName}
              onChange={(e) => setDeleteConfirmName(e.target.value)}
            />
            <Button
              variant="secondary"
              onClick={handleDeleteChild}
              disabled={!nameMatches || deleting}
              className="w-full mt-4 !bg-terracotta/10 !text-terracotta !border-terracotta/20 hover:!bg-terracotta/20 disabled:opacity-50"
            >
              {deleting ? "Deleting..." : "Permanently Delete Account"}
            </Button>
          </Card>
        </>
      )}

      {error && (
        <div className="bg-terracotta/10 border border-terracotta/20 rounded-xl p-3">
          <p className="text-sm text-terracotta font-medium">{error}</p>
        </div>
      )}
    </div>
  );
}
