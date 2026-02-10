import { useState, useEffect } from "react";
import {
  setChildAllowance,
  deleteChildAllowance,
  pauseChildAllowance,
  resumeChildAllowance,
  ApiRequestError,
} from "../api";
import { AllowanceSchedule, Frequency } from "../types";

interface ChildAllowanceFormProps {
  childId: number;
  childName: string;
  allowance: AllowanceSchedule | null;
  onUpdated: (allowance: AllowanceSchedule | null) => void;
}

const DAYS_OF_WEEK = [
  { value: 0, label: "Sunday" },
  { value: 1, label: "Monday" },
  { value: 2, label: "Tuesday" },
  { value: 3, label: "Wednesday" },
  { value: 4, label: "Thursday" },
  { value: 5, label: "Friday" },
  { value: 6, label: "Saturday" },
];

export default function ChildAllowanceForm({
  childId,
  childName,
  allowance,
  onUpdated,
}: ChildAllowanceFormProps) {
  const [amount, setAmount] = useState("");
  const [frequency, setFrequency] = useState<Frequency>("weekly");
  const [dayOfWeek, setDayOfWeek] = useState(5);
  const [dayOfMonth, setDayOfMonth] = useState(1);
  const [note, setNote] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Pre-populate form from existing allowance
  useEffect(() => {
    if (allowance) {
      setAmount((allowance.amount_cents / 100).toFixed(2));
      setFrequency(allowance.frequency);
      if (allowance.day_of_week != null) setDayOfWeek(allowance.day_of_week);
      if (allowance.day_of_month != null) setDayOfMonth(allowance.day_of_month);
      setNote(allowance.note || "");
    }
  }, [allowance]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

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

    setSaving(true);
    try {
      const result = await setChildAllowance(childId, {
        amount_cents: amountCents,
        frequency,
        day_of_week: frequency !== "monthly" ? dayOfWeek : undefined,
        day_of_month: frequency === "monthly" ? dayOfMonth : undefined,
        note: note.trim() || undefined,
      });
      setSuccess(allowance ? "Allowance updated." : "Allowance created.");
      onUpdated(result);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to save allowance.");
      }
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    setError(null);
    setSuccess(null);
    setSaving(true);
    try {
      await deleteChildAllowance(childId);
      setSuccess("Allowance removed.");
      setAmount("");
      setNote("");
      onUpdated(null);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to remove allowance.");
      }
    } finally {
      setSaving(false);
    }
  };

  const handlePause = async () => {
    setError(null);
    setSuccess(null);
    setSaving(true);
    try {
      const result = await pauseChildAllowance(childId);
      setSuccess("Allowance paused.");
      onUpdated(result);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to pause allowance.");
      }
    } finally {
      setSaving(false);
    }
  };

  const handleResume = async () => {
    setError(null);
    setSuccess(null);
    setSaving(true);
    try {
      const result = await resumeChildAllowance(childId);
      setSuccess("Allowance resumed.");
      onUpdated(result);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to resume allowance.");
      }
    } finally {
      setSaving(false);
    }
  };

  const formatNextRun = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  return (
    <div className="child-allowance-form">
      <h4>Allowance for {childName}</h4>

      {allowance && (
        <div className="allowance-status">
          <p>
            Status: <strong>{allowance.status === "active" ? "Active" : "Paused"}</strong>
            {allowance.status === "active" && allowance.next_run_at && (
              <> &mdash; Next: {formatNextRun(allowance.next_run_at)}</>
            )}
          </p>
          <div className="allowance-actions">
            {allowance.status === "active" ? (
              <button type="button" onClick={handlePause} disabled={saving} className="btn-secondary">
                Pause
              </button>
            ) : (
              <button type="button" onClick={handleResume} disabled={saving} className="btn-primary">
                Resume
              </button>
            )}
            <button type="button" onClick={handleDelete} disabled={saving} className="btn-danger">
              Remove
            </button>
          </div>
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="allowance-amount">Amount ($)</label>
          <input
            type="number"
            id="allowance-amount"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            placeholder="0.00"
            step="0.01"
            min="0.01"
            max="999999.99"
            required
            disabled={saving}
          />
        </div>

        <div className="form-group">
          <label htmlFor="allowance-frequency">Frequency</label>
          <select
            id="allowance-frequency"
            value={frequency}
            onChange={(e) => setFrequency(e.target.value as Frequency)}
            disabled={saving}
          >
            <option value="weekly">Weekly</option>
            <option value="biweekly">Every 2 Weeks</option>
            <option value="monthly">Monthly</option>
          </select>
        </div>

        {frequency !== "monthly" && (
          <div className="form-group">
            <label htmlFor="allowance-day-of-week">Day of Week</label>
            <select
              id="allowance-day-of-week"
              value={dayOfWeek}
              onChange={(e) => setDayOfWeek(Number(e.target.value))}
              disabled={saving}
            >
              {DAYS_OF_WEEK.map((d) => (
                <option key={d.value} value={d.value}>
                  {d.label}
                </option>
              ))}
            </select>
          </div>
        )}

        {frequency === "monthly" && (
          <div className="form-group">
            <label htmlFor="allowance-day-of-month">Day of Month</label>
            <input
              type="number"
              id="allowance-day-of-month"
              value={dayOfMonth}
              onChange={(e) => setDayOfMonth(Number(e.target.value))}
              min={1}
              max={31}
              required
              disabled={saving}
            />
          </div>
        )}

        <div className="form-group">
          <label htmlFor="allowance-note">Note (optional)</label>
          <input
            type="text"
            id="allowance-note"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="e.g., Weekly allowance"
            maxLength={500}
            disabled={saving}
          />
        </div>

        <button type="submit" disabled={saving} className="btn-primary">
          {saving ? "Saving..." : allowance ? "Update Allowance" : "Set Up Allowance"}
        </button>
      </form>

      {success && <p className="success">{success}</p>}
      {error && <p className="error">{error}</p>}
    </div>
  );
}
