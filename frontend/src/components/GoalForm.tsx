import { useState, useMemo } from "react";
import { SavingsGoal } from "../types";
import Button from "./ui/Button";
import Input from "./ui/Input";

const EMOJI_KEYWORDS: Record<string, string> = {
  bike: "🚲", bicycle: "🚲", skateboard: "🛹", scooter: "🛴",
  game: "🎮", gaming: "🎮", xbox: "🎮", playstation: "🎮", nintendo: "🎮", switch: "🎮",
  lego: "🧱", book: "📚", reading: "📚",
  phone: "📱", laptop: "💻", computer: "💻", tablet: "📱", ipad: "📱",
  toy: "🧸", toys: "🧸", doll: "🪆", car: "🚗",
  trip: "✈️", vacation: "✈️", travel: "✈️",
  pet: "🐕", dog: "🐕", cat: "🐱", puppy: "🐕", kitten: "🐱",
  shoes: "👟", sneakers: "👟", clothes: "👗", dress: "👗",
  candy: "🍬", sweets: "🍬", chocolate: "🍫",
  guitar: "🎸", music: "🎸", drum: "🥁", piano: "🎹",
  art: "🎨", paint: "🎨", draw: "🎨", camera: "📷",
  ball: "⚽", soccer: "⚽", football: "🏈", basketball: "🏀", sport: "⚽",
  swim: "🏊", tent: "⛺", camp: "⛺", camping: "⛺",
  robot: "🤖", science: "🔬", space: "🚀", rocket: "🚀",
  horse: "🐴", fish: "🐠", dinosaur: "🦕",
};

const DEFAULT_EMOJIS = ["🎯", "⭐", "💰", "🎁", "🏆", "💎"];

function suggestEmojis(goalName: string, selectedEmoji: string): string[] {
  const words = goalName.toLowerCase().split(/\s+/);
  const matched: string[] = [];
  const seen = new Set<string>();

  for (const word of words) {
    const emoji = EMOJI_KEYWORDS[word];
    if (emoji && !seen.has(emoji)) {
      seen.add(emoji);
      matched.push(emoji);
    }
    if (matched.length >= 6) break;
  }

  // Fill remaining slots with defaults
  for (const d of DEFAULT_EMOJIS) {
    if (matched.length >= 6) break;
    if (!seen.has(d)) {
      matched.push(d);
      seen.add(d);
    }
  }

  // If selected emoji isn't in the list, replace the last one
  if (selectedEmoji && !matched.includes(selectedEmoji)) {
    matched[matched.length - 1] = selectedEmoji;
  }

  return matched;
}

interface GoalFormProps {
  onSubmit: (data: { name: string; target_cents: number; emoji?: string }) => Promise<void>;
  onCancel: () => void;
  initialGoal?: SavingsGoal;
}

export default function GoalForm({ onSubmit, onCancel, initialGoal }: GoalFormProps) {
  const [name, setName] = useState(initialGoal?.name ?? "");
  const [targetDollars, setTargetDollars] = useState(
    initialGoal ? (initialGoal.target_cents / 100).toFixed(2) : ""
  );
  const [emoji, setEmoji] = useState(initialGoal?.emoji ?? "");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [nameError, setNameError] = useState<string | null>(null);
  const [targetError, setTargetError] = useState<string | null>(null);
  const suggestedEmojis = useMemo(() => suggestEmojis(name, emoji), [name, emoji]);

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

    setLoading(true);
    try {
      await onSubmit({
        name: trimmedName,
        target_cents: targetCents,
        emoji: emoji.trim() || undefined,
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

      <div className="space-y-1.5">
        <label className="block text-sm font-semibold text-bark-light">
          Emoji (optional)
        </label>
        <div className="flex gap-2">
          {suggestedEmojis.map((e) => (
            <button
              key={e}
              type="button"
              onClick={() => setEmoji(emoji === e ? "" : e)}
              className={`
                w-10 h-10 text-xl rounded-lg border-2 transition-all duration-150
                flex items-center justify-center cursor-pointer
                ${emoji === e
                  ? "border-forest bg-forest/10 scale-110 ring-2 ring-forest/30"
                  : "border-sand bg-white hover:border-forest/40 hover:bg-forest/5"
                }
              `}
            >
              {e}
            </button>
          ))}
        </div>
      </div>

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
