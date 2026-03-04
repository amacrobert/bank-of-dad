import { useEffect, useState } from "react";
import { SavingsGoal } from "../types";
import Button from "./ui/Button";
import Input from "./ui/Input";

interface GoalFormProps {
  onSubmit: (data: { name: string; target_cents: number; emoji?: string; target_date?: string }) => Promise<void>;
  onCancel: () => void;
  initialGoal?: SavingsGoal;
}

export default function GoalForm({ onSubmit, onCancel, initialGoal }: GoalFormProps) {
  const [name, setName] = useState(initialGoal?.name ?? "");
  const [targetDollars, setTargetDollars] = useState(
    initialGoal ? (initialGoal.target_cents / 100).toFixed(2) : ""
  );
  const [emoji, setEmoji] = useState(initialGoal?.emoji ?? "");
  const [targetDate, setTargetDate] = useState(initialGoal?.target_date ?? "");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [nameError, setNameError] = useState<string | null>(null);
  const [targetError, setTargetError] = useState<string | null>(null);

  useEffect(() => {
    setName(initialGoal?.name ?? "");
    setTargetDollars(initialGoal ? (initialGoal.target_cents / 100).toFixed(2) : "");
    setEmoji(initialGoal?.emoji ?? "");
    setTargetDate(initialGoal?.target_date ?? "");
    setError(null);
    setNameError(null);
    setTargetError(null);
  }, [initialGoal]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setNameError(null);
    setTargetError(null);

    // Validate name
    const trimmedName = name.trim();
    if (!trimmedName) {
      setNameError("Name is required.");
      return;
    }
    if (trimmedName.length > 50) {
      setNameError("Name must be 50 characters or less.");
      return;
    }

    // Validate target amount
    const dollars = parseFloat(targetDollars);
    if (isNaN(dollars) || dollars < 0.01) {
      setTargetError("Amount must be at least $0.01.");
      return;
    }
    if (dollars > 999999.99) {
      setTargetError("Amount must be $999,999.99 or less.");
      return;
    }

    const targetCents = Math.round(dollars * 100);
    const trimmedEmoji = emoji.trim();
    const trimmedTargetDate = targetDate.trim();

    setLoading(true);
    try {
      await onSubmit({
        name: trimmedName,
        target_cents: targetCents,
        emoji: initialGoal ? trimmedEmoji : (trimmedEmoji || undefined),
        target_date: initialGoal ? trimmedTargetDate : (trimmedTargetDate || undefined),
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <Input
        id="goal-name"
        label="Goal Name"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="e.g., New Skateboard"
        maxLength={50}
        required
        error={nameError}
      />

      <div className="space-y-1.5">
        <label htmlFor="goal-target" className="block text-sm font-semibold text-bark-light">
          Target Amount
        </label>
        <div className="relative">
          <span className="absolute left-4 top-1/2 -translate-y-1/2 text-bark-light font-medium">$</span>
          <input
            id="goal-target"
            type="number"
            step="0.01"
            min="0.01"
            max="999999.99"
            value={targetDollars}
            onChange={(e) => setTargetDollars(e.target.value)}
            placeholder="0.00"
            required
            className={`
              w-full min-h-[48px] pl-8 pr-4 py-3
              rounded-xl border border-sand bg-white
              text-bark text-base placeholder:text-bark-light/50
              transition-all duration-200
              focus:outline-none focus:ring-2 focus:ring-forest/30 focus:border-forest
              ${targetError ? "border-terracotta ring-1 ring-terracotta/30" : ""}
            `}
          />
        </div>
        {targetError && (
          <p className="text-sm text-terracotta font-medium">{targetError}</p>
        )}
      </div>

      <Input
        id="goal-emoji"
        label="Emoji (optional)"
        value={emoji}
        onChange={(e) => setEmoji(e.target.value)}
        placeholder="🎯"
      />

      <Input
        id="goal-target-date"
        label="Target Date (optional)"
        type="date"
        value={targetDate}
        onChange={(e) => setTargetDate(e.target.value)}
      />

      {error && (
        <p className="text-sm text-terracotta font-medium">{error}</p>
      )}

      <div className="flex gap-3">
        <Button type="submit" loading={loading} className="flex-1">
          {initialGoal ? "Save Changes" : "Create Goal"}
        </Button>
        <Button type="button" variant="secondary" onClick={onCancel}>
          Cancel
        </Button>
      </div>
    </form>
  );
}
