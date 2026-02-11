import { useState, useEffect } from "react";
import { setInterestRate, ApiRequestError } from "../api";
import Card from "./ui/Card";
import Button from "./ui/Button";
import { TrendingUp } from "lucide-react";

interface InterestRateFormProps {
  childId: number;
  childName: string;
  currentRateBps: number;
  onSuccess: (newRateBps: number) => void;
}

export default function InterestRateForm({
  childId,
  childName,
  currentRateBps,
  onSuccess,
}: InterestRateFormProps) {
  const [ratePercent, setRatePercent] = useState(
    (currentRateBps / 100).toFixed(2)
  );

  useEffect(() => {
    setRatePercent((currentRateBps / 100).toFixed(2));
  }, [currentRateBps]);

  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    const parsed = parseFloat(ratePercent);
    if (isNaN(parsed) || parsed < 0 || parsed > 100) {
      setError("Rate must be between 0% and 100%.");
      return;
    }

    const bps = Math.round(parsed * 100);

    setSaving(true);
    try {
      const result = await setInterestRate(childId, {
        interest_rate_bps: bps,
      });
      setSuccess(
        bps === 0
          ? `Interest disabled for ${childName}.`
          : `Interest rate set to ${result.interest_rate_display} for ${childName}.`
      );
      onSuccess(result.interest_rate_bps);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to update interest rate.");
      }
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card padding="md">
      <div className="flex items-center gap-2 mb-4">
        <TrendingUp className="h-5 w-5 text-forest" aria-hidden="true" />
        <h4 className="text-base font-bold text-bark">Interest Rate</h4>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-1.5">
          <label htmlFor="interest-rate" className="block text-sm font-semibold text-bark-light">
            Annual Interest Rate
          </label>
          <div className="flex items-center rounded-xl border border-sand bg-white overflow-hidden focus-within:ring-2 focus-within:ring-forest/30 focus-within:border-forest transition-all">
            <input
              id="interest-rate"
              type="number"
              step="0.01"
              min="0"
              max="100"
              value={ratePercent}
              onChange={(e) => setRatePercent(e.target.value)}
              disabled={saving}
              className="flex-1 min-h-[48px] px-4 py-3 bg-transparent text-bark text-base placeholder:text-bark-light/50 focus:outline-none disabled:cursor-not-allowed"
            />
            <span className="px-3 py-3 bg-cream-dark text-bark-light text-base font-medium border-l border-sand">%</span>
          </div>
        </div>

        <Button type="submit" loading={saving} className="w-full">
          {saving ? "Saving..." : "Set Rate"}
        </Button>
      </form>

      {success && (
        <div className="mt-3 bg-forest/5 border border-forest/15 rounded-xl p-3">
          <p className="text-sm text-forest font-medium">{success}</p>
        </div>
      )}
      {error && (
        <div className="mt-3 bg-terracotta/10 border border-terracotta/20 rounded-xl p-3">
          <p className="text-sm text-terracotta font-medium">{error}</p>
        </div>
      )}
    </Card>
  );
}
