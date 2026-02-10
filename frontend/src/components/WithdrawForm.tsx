import { useState } from "react";
import { withdraw, ApiRequestError } from "../api";
import { Child } from "../types";
import Card from "./ui/Card";
import Input from "./ui/Input";
import Button from "./ui/Button";

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
    <Card padding="md" className="animate-fade-in-up">
      <h4 className="text-base font-bold text-bark mb-1">
        Withdraw from {child.first_name}'s Account
      </h4>
      <p className="text-sm text-bark-light mb-4">
        Current balance: <span className="font-semibold text-bark">${currentBalanceDollars.toFixed(2)}</span>
      </p>

      {error && (
        <div className="mb-4 bg-terracotta/10 border border-terracotta/20 rounded-xl p-3">
          <p className="text-sm text-terracotta font-medium">{error}</p>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-1.5">
          <label htmlFor="withdraw-amount" className="block text-sm font-semibold text-bark-light">
            Amount
          </label>
          <div className="flex items-center rounded-xl border border-sand bg-white overflow-hidden focus-within:ring-2 focus-within:ring-forest/30 focus-within:border-forest transition-all">
            <span className="px-3 py-3 bg-cream-dark text-bark-light text-base font-medium border-r border-sand">$</span>
            <input
              type="number"
              id="withdraw-amount"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="0.00"
              step="0.01"
              min="0.01"
              max={currentBalanceDollars}
              required
              disabled={loading}
              className="flex-1 min-h-[48px] px-3 py-3 bg-transparent text-bark text-base placeholder:text-bark-light/50 focus:outline-none disabled:cursor-not-allowed"
            />
          </div>
        </div>

        <Input
          label="Note (optional)"
          id="withdraw-note"
          type="text"
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="e.g., Bought a book"
          maxLength={500}
          disabled={loading}
        />

        <div className="flex gap-3">
          <Button
            type="submit"
            variant="danger"
            loading={loading}
            disabled={child.balance_cents === 0}
            className="flex-1"
          >
            {loading ? "Processing..." : "Withdraw"}
          </Button>
          <Button type="button" variant="secondary" onClick={onCancel} disabled={loading}>
            Cancel
          </Button>
        </div>
      </form>
    </Card>
  );
}
