import { useState, useEffect } from "react";
import {
  setChildAllowance,
  deleteChildAllowance,
  pauseChildAllowance,
  resumeChildAllowance,
  ApiRequestError,
} from "../api";
import { AllowanceSchedule, Frequency } from "../types";
import Card from "./ui/Card";
import Input from "./ui/Input";
import Select from "./ui/Select";
import Button from "./ui/Button";
import { Calendar } from "lucide-react";
import { useTimezone } from "../context/TimezoneContext";

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
  const timezone = useTimezone();
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

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
      timeZone: timezone,
    });
  };

  return (
    <Card padding="md">
      <div className="flex items-center gap-2 mb-4">
        <Calendar className="h-5 w-5 text-forest" aria-hidden="true" />
        <h4 className="text-base font-bold text-bark">Allowance for {childName}</h4>
      </div>

      {allowance && (
        <div className="mb-4 p-3 bg-cream rounded-xl">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <span className="text-sm text-bark-light">Status:</span>
              <span className={`
                inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold
                ${allowance.status === "active"
                  ? "bg-sage-light/40 text-forest"
                  : "bg-sand text-bark-light"
                }
              `}>
                {allowance.status === "active" ? "Active" : "Paused"}
              </span>
            </div>
            {allowance.status === "active" && allowance.next_run_at && (
              <span className="text-xs text-bark-light">
                Next: {formatNextRun(allowance.next_run_at)}
              </span>
            )}
          </div>
          <div className="flex gap-2">
            {allowance.status === "active" ? (
              <Button variant="secondary" onClick={handlePause} disabled={saving} className="text-sm !min-h-[36px] !px-3 !py-1">
                Pause
              </Button>
            ) : (
              <Button variant="primary" onClick={handleResume} disabled={saving} className="text-sm !min-h-[36px] !px-3 !py-1">
                Resume
              </Button>
            )}
            <Button variant="danger" onClick={handleDelete} disabled={saving} className="text-sm !min-h-[36px] !px-3 !py-1">
              Remove
            </Button>
          </div>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-1.5">
          <label htmlFor="allowance-amount" className="block text-sm font-semibold text-bark-light">
            Amount
          </label>
          <div className="flex items-center rounded-xl border border-sand bg-white overflow-hidden focus-within:ring-2 focus-within:ring-forest/30 focus-within:border-forest transition-all">
            <span className="px-3 py-3 bg-cream-dark text-bark-light text-base font-medium border-r border-sand">$</span>
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
              className="flex-1 min-h-[48px] px-3 py-3 bg-transparent text-bark text-base placeholder:text-bark-light/50 focus:outline-none disabled:cursor-not-allowed"
            />
          </div>
        </div>

        <Select
          label="Frequency"
          id="allowance-frequency"
          value={frequency}
          onChange={(e) => setFrequency(e.target.value as Frequency)}
          disabled={saving}
        >
          <option value="weekly">Weekly</option>
          <option value="biweekly">Every 2 Weeks</option>
          <option value="monthly">Monthly</option>
        </Select>

        {frequency !== "monthly" && (
          <Select
            label="Day of Week"
            id="allowance-day-of-week"
            value={dayOfWeek}
            onChange={(e) => setDayOfWeek(Number(e.target.value))}
            disabled={saving}
          >
            {DAYS_OF_WEEK.map((d) => (
              <option key={d.value} value={d.value}>{d.label}</option>
            ))}
          </Select>
        )}

        {frequency === "monthly" && (
          <Input
            label="Day of Month"
            id="allowance-day-of-month"
            type="number"
            value={dayOfMonth}
            onChange={(e) => setDayOfMonth(Number(e.target.value))}
            min={1}
            max={31}
            required
            disabled={saving}
          />
        )}

        <Input
          label="Note (optional)"
          id="allowance-note"
          type="text"
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="e.g., Weekly allowance"
          maxLength={500}
          disabled={saving}
        />

        <Button type="submit" loading={saving} className="w-full">
          {saving ? "Saving..." : allowance ? "Update Allowance" : "Set Up Allowance"}
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
