import { useState, useEffect } from "react";
import { put, getBalance, ApiRequestError } from "../api";
import { Child } from "../types";
import BalanceDisplay from "./BalanceDisplay";
import DepositForm from "./DepositForm";
import WithdrawForm from "./WithdrawForm";
import InterestRateForm from "./InterestRateForm";

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

  useEffect(() => {
    getBalance(child.id).then((data) => {
      setInterestRateBps(data.interest_rate_bps);
    }).catch(() => {
      // Ignore - interest rate will show as 0
    });
  }, [child.id]);

  const handleDepositSuccess = (newBalance: number) => {
    setCurrentBalance(newBalance);
    setShowDeposit(false);
    onUpdated();
  };

  const handleWithdrawSuccess = (newBalance: number) => {
    setCurrentBalance(newBalance);
    setShowWithdraw(false);
    onUpdated();
  };

  // Update local child reference with current balance for withdrawal form
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
    <div className="manage-child">
      <h3>Manage {child.first_name}</h3>
      {child.is_locked && (
        <p className="warning">This account is locked. Reset the password to unlock it.</p>
      )}

      <div className="balance-section">
        <h4>Balance</h4>
        <BalanceDisplay balanceCents={currentBalance} size="large" />
        <div className="balance-actions">
          <button onClick={() => { setShowDeposit(true); setShowWithdraw(false); }} className="btn-primary">
            Deposit
          </button>
          <button onClick={() => { setShowWithdraw(true); setShowDeposit(false); }} className="btn-secondary" disabled={currentBalance === 0}>
            Withdraw
          </button>
        </div>
      </div>

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

      <InterestRateForm
        childId={child.id}
        childName={child.first_name}
        currentRateBps={interestRateBps}
        onSuccess={(newRate) => setInterestRateBps(newRate)}
      />

      <form onSubmit={handleResetPassword}>
        <h4>Reset Password</h4>
        <div className="form-field">
          <label htmlFor="new-password">New Password (min 6 characters)</label>
          <input
            id="new-password"
            type="text"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            minLength={6}
            required
          />
        </div>
        <button type="submit">Reset Password</button>
        {passwordMsg && <p className="success">{passwordMsg}</p>}
      </form>

      <form onSubmit={handleUpdateName}>
        <h4>Update Name</h4>
        <div className="form-field">
          <label htmlFor="new-name">First Name</label>
          <input
            id="new-name"
            type="text"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            required
          />
        </div>
        <button type="submit">Update Name</button>
        {nameMsg && <p className="success">{nameMsg}</p>}
      </form>

      {error && <p className="error">{error}</p>}

      <button onClick={onClose} className="btn-secondary">
        Close
      </button>
    </div>
  );
}
