import { useState } from "react";
import { deposit } from "../api";
import { ApiRequestError } from "../api";
import { Child } from "../types";

interface DepositFormProps {
  child: Child;
  onSuccess: (newBalance: number) => void;
  onCancel: () => void;
}

export default function DepositForm({ child, onSuccess, onCancel }: DepositFormProps) {
  const [amount, setAmount] = useState("");
  const [note, setNote] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

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

    setLoading(true);
    try {
      const response = await deposit(child.id, {
        amount_cents: amountCents,
        note: note.trim() || undefined,
      });
      onSuccess(response.new_balance_cents);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to process deposit. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="deposit-form">
      <h4>Deposit to {child.first_name}'s Account</h4>

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
          max="999999.99"
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
          placeholder="e.g., Weekly allowance"
          maxLength={500}
          disabled={loading}
        />
      </div>

      <div className="form-actions">
        <button type="submit" disabled={loading} className="btn-primary">
          {loading ? "Processing..." : "Deposit"}
        </button>
        <button type="button" onClick={onCancel} disabled={loading} className="btn-secondary">
          Cancel
        </button>
      </div>
    </form>
  );
}
