import { useState } from "react";
import { submitWithdrawalRequest, ApiRequestError } from "../api";
import Input from "./ui/Input";
import Button from "./ui/Button";

interface WithdrawalRequestFormProps {
  availableBalanceCents: number;
  onSuccess: () => void;
  onCancel: () => void;
}

export default function WithdrawalRequestForm({
  availableBalanceCents,
  onSuccess,
  onCancel,
}: WithdrawalRequestFormProps) {
  const [amount, setAmount] = useState("");
  const [reason, setReason] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const availableDollars = availableBalanceCents / 100;

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

    if (amountCents > availableBalanceCents) {
      setError(
        `Cannot request more than your available balance of $${availableDollars.toFixed(2)}`
      );
      return;
    }

    const trimmedReason = reason.trim();
    if (!trimmedReason) {
      setError("Please provide a reason for your request.");
      return;
    }

    if (trimmedReason.length > 500) {
      setError("Reason must be 500 characters or less.");
      return;
    }

    setLoading(true);
    try {
      await submitWithdrawalRequest({
        amount_cents: amountCents,
        reason: trimmedReason,
      });
      onSuccess();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to submit request. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h4 className="text-base font-bold text-bark mb-1">
        Request a Withdrawal
      </h4>
      <p className="text-sm text-bark-light mb-4">
        Available balance:{" "}
        <span className="font-semibold text-bark">
          ${availableDollars.toFixed(2)}
        </span>
      </p>

      {error && (
        <div className="mb-4 bg-terracotta/10 border border-terracotta/20 rounded-xl p-3">
          <p className="text-sm text-terracotta font-medium">{error}</p>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-1.5">
          <label
            htmlFor="wr-amount"
            className="block text-sm font-semibold text-bark-light"
          >
            Amount
          </label>
          <div className="flex items-center rounded-xl border border-sand bg-white overflow-hidden focus-within:ring-2 focus-within:ring-forest/30 focus-within:border-forest transition-all">
            <span className="px-3 py-3 bg-cream-dark text-bark-light text-base font-medium border-r border-sand">
              $
            </span>
            <input
              type="number"
              id="wr-amount"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="0.00"
              step="0.01"
              min="0.01"
              max={availableDollars}
              required
              disabled={loading}
              className="flex-1 min-h-[48px] px-3 py-3 bg-transparent text-bark text-base placeholder:text-bark-light/50 focus:outline-none disabled:cursor-not-allowed"
            />
          </div>
        </div>

        <Input
          label="What's it for?"
          id="wr-reason"
          type="text"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder="e.g., New video game"
          maxLength={500}
          disabled={loading}
        />

        <div className="flex gap-3">
          <Button
            type="submit"
            variant="primary"
            loading={loading}
            disabled={availableBalanceCents === 0}
            className="flex-1"
          >
            {loading ? "Submitting..." : "Submit Request"}
          </Button>
          <Button
            type="button"
            variant="secondary"
            onClick={onCancel}
            disabled={loading}
          >
            Cancel
          </Button>
        </div>
      </form>
    </div>
  );
}
