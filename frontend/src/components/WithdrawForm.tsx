import { useState } from "react";
import { withdraw } from "../api";
import { ApiRequestError } from "../api";
import { Child } from "../types";

interface WithdrawFormProps {
  child: Child;
  onSuccess: (newBalance: number) => void;
  onCancel: () => void;
}

export default function WithdrawForm({ child, onSuccess, onCancel }: WithdrawFormProps) {
  const [amount, setAmount] = useState("");
  const [note, setNote] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const currentBalanceDollars = child.balance_cents / 100;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const amountNum = parseFloat(amount);
    if (isNaN(amountNum) || amountNum <= 0) {
      setError("Please enter a valid amount greater than $0.00");
      return;
    }

    if (amountNum > 999999.99) {
      setError("Amount cannot exceed $999,999.99");
      return;
    }

    const amountCents = Math.round(amountNum * 100);

    if (amountCents > child.balance_cents) {
      setError(`Cannot withdraw more than the current balance of $${currentBalanceDollars.toFixed(2)}`);
      return;
    }

    setLoading(true);
    try {
      const response = await withdraw(child.id, {
        amount_cents: amountCents,
        note: note.trim() || undefined,
      });
      onSuccess(response.new_balance_cents);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to process withdrawal. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="withdraw-form">
      <h4>Withdraw from {child.first_name}'s Account</h4>
      <p className="current-balance">Current balance: ${currentBalanceDollars.toFixed(2)}</p>

      {error && <div className="error-message">{error}</div>}

      <div className="form-group">
        <label htmlFor="amount">Amount ($)</label>
        <input
          type="number"
          id="amount"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="0.00"
          step="0.01"
          min="0.01"
          max={currentBalanceDollars}
          required
          disabled={loading}
        />
      </div>

      <div className="form-group">
        <label htmlFor="note">Note (optional)</label>
        <input
          type="text"
          id="note"
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="e.g., Bought a book"
          maxLength={500}
          disabled={loading}
        />
      </div>

      <div className="form-actions">
        <button type="submit" disabled={loading || child.balance_cents === 0} className="btn-primary">
          {loading ? "Processing..." : "Withdraw"}
        </button>
        <button type="button" onClick={onCancel} disabled={loading} className="btn-secondary">
          Cancel
        </button>
      </div>
    </form>
  );
}
