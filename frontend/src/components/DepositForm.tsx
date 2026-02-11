import { useState } from "react";
import { deposit, ApiRequestError } from "../api";
import { Child } from "../types";
import Card from "./ui/Card";
import Input from "./ui/Input";
import Button from "./ui/Button";

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
    <Card padding="md" className="animate-fade-in-up">
      <h4 className="text-base font-bold text-bark mb-4">
        Deposit to {child.first_name}'s Account
      </h4>

      {error && (
        <div className="mb-4 bg-terracotta/10 border border-terracotta/20 rounded-xl p-3">
          <p className="text-sm text-terracotta font-medium">{error}</p>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-1.5">
          <label htmlFor="deposit-amount" className="block text-sm font-semibold text-bark-light">
            Amount
          </label>
          <div className="flex items-center rounded-xl border border-sand bg-white overflow-hidden focus-within:ring-2 focus-within:ring-forest/30 focus-within:border-forest transition-all">
            <span className="px-3 py-3 bg-cream-dark text-bark-light text-base font-medium border-r border-sand">$</span>
            <input
              type="number"
              id="deposit-amount"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="0.00"
              step="0.01"
              min="0.01"
              max="999999.99"
              required
              disabled={loading}
              className="flex-1 min-h-[48px] px-3 py-3 bg-transparent text-bark text-base placeholder:text-bark-light/50 focus:outline-none disabled:cursor-not-allowed"
            />
          </div>
        </div>

        <Input
          label="Note (optional)"
          id="deposit-note"
          type="text"
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="e.g., Weekly allowance"
          maxLength={500}
          disabled={loading}
        />

        <div className="flex gap-3">
          <Button type="submit" loading={loading} className="flex-1">
            {loading ? "Processing..." : "Deposit"}
          </Button>
          <Button type="button" variant="secondary" onClick={onCancel} disabled={loading}>
            Cancel
          </Button>
        </div>
      </form>
    </Card>
  );
}
