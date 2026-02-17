import { useState, useEffect } from "react";
import { setInterest, ApiRequestError } from "../api";
import { InterestSchedule, Frequency } from "../types";
import Card from "./ui/Card";
import Input from "./ui/Input";
import Select from "./ui/Select";
import Button from "./ui/Button";
import { TrendingUp } from "lucide-react";
import { useTimezone } from "../context/TimezoneContext";

interface InterestFormProps {
  childId: number;
  childName: string;
  currentRateBps: number;
  schedule: InterestSchedule | null;
  onUpdated: (rateBps: number, schedule: InterestSchedule | null) => void;
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

export default function InterestForm({
  childId,
  childName,
  currentRateBps,
  schedule,
  onUpdated,
}: InterestFormProps) {
  const [ratePercent, setRatePercent] = useState(
    (currentRateBps / 100).toFixed(2)
  );
  const [frequency, setFrequency] = useState<Frequency>(
    schedule?.frequency || "monthly"
  );
  const [dayOfWeek, setDayOfWeek] = useState(schedule?.day_of_week ?? 0);
  const [dayOfMonth, setDayOfMonth] = useState(schedule?.day_of_month ?? 1);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    setRatePercent((currentRateBps / 100).toFixed(2));
  }, [currentRateBps]);

  useEffect(() => {
    if (schedule) {
      setFrequency(schedule.frequency);
      if (schedule.day_of_week != null) setDayOfWeek(schedule.day_of_week);
      if (schedule.day_of_month != null) setDayOfMonth(schedule.day_of_month);
    }
  }, [schedule]);

  const timezone = useTimezone();
  const parsedRate = parseFloat(ratePercent);
  const rateIsPositive = !isNaN(parsedRate) && parsedRate > 0;

  const formatNextRun = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
      timeZone: timezone,
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    if (isNaN(parsedRate) || parsedRate < 0 || parsedRate > 100) {
      setError("Rate must be between 0% and 100%.");
      return;
    }

    const bps = Math.round(parsedRate * 100);

    setSaving(true);
    try {
      const result = await setInterest(childId, {
        interest_rate_bps: bps,
        frequency: bps > 0 ? frequency : undefined,
        day_of_week:
          bps > 0 && frequency !== "monthly" ? dayOfWeek : undefined,
        day_of_month:
          bps > 0 && frequency === "monthly" ? dayOfMonth : undefined,
      });

      if (bps === 0) {
        setSuccess(`Interest disabled for ${childName}.`);
      } else {
        setSuccess(
          `Interest rate set to ${result.interest_rate_display} for ${childName}.`
        );
      }
      onUpdated(result.interest_rate_bps, result.schedule);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to update interest settings.");
      }
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card padding="md">
      <div className="flex items-center gap-2 mb-4">
        <TrendingUp className="h-5 w-5 text-forest" aria-hidden="true" />
        <h4 className="text-base font-bold text-bark">Interest</h4>
      </div>

      {schedule && schedule.status === "active" && schedule.next_run_at && (
        <div className="mb-4 p-3 bg-cream rounded-xl">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="text-sm text-bark-light">Status:</span>
              <span className="inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold bg-sage-light/40 text-forest">
                Active
              </span>
            </div>
            <span className="text-xs text-bark-light">
              Next payout: {formatNextRun(schedule.next_run_at)}
            </span>
          </div>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-1.5">
          <label
            htmlFor="interest-rate"
            className="block text-sm font-semibold text-bark-light"
          >
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
            <span className="px-3 py-3 bg-cream-dark text-bark-light text-base font-medium border-l border-sand">
              %
            </span>
          </div>
          {!rateIsPositive && (
            <p className="text-xs text-bark-light">
              Set to 0% to disable interest.
            </p>
          )}
        </div>

        {rateIsPositive && (
          <>
            <Select
              label="Payout Frequency"
              id="interest-frequency"
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
                id="interest-day-of-week"
                value={dayOfWeek}
                onChange={(e) => setDayOfWeek(Number(e.target.value))}
                disabled={saving}
              >
                {DAYS_OF_WEEK.map((d) => (
                  <option key={d.value} value={d.value}>
                    {d.label}
                  </option>
                ))}
              </Select>
            )}

            {frequency === "monthly" && (
              <Input
                label="Day of Month"
                id="interest-day-of-month"
                type="number"
                value={dayOfMonth}
                onChange={(e) => setDayOfMonth(Number(e.target.value))}
                min={1}
                max={31}
                required
                disabled={saving}
              />
            )}
          </>
        )}

        <Button type="submit" loading={saving} className="w-full">
          {saving ? "Saving..." : "Save Interest Settings"}
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
