import { useState, useEffect } from "react";
import { setInterestRate, ApiRequestError } from "../api";

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
    <div className="interest-rate-form">
      <h4>Interest Rate</h4>
      <form onSubmit={handleSubmit}>
        <div className="form-field">
          <label htmlFor="interest-rate">Annual Interest Rate (%)</label>
          <input
            id="interest-rate"
            type="number"
            step="0.01"
            min="0"
            max="100"
            value={ratePercent}
            onChange={(e) => setRatePercent(e.target.value)}
          />
        </div>
        <button type="submit" disabled={saving}>
          {saving ? "Saving..." : "Set Rate"}
        </button>
      </form>
      {success && <p className="success">{success}</p>}
      {error && <p className="error">{error}</p>}
    </div>
  );
}
