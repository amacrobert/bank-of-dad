import { useState, useEffect } from "react";
import {
  setInterestSchedule,
  deleteInterestSchedule,
  ApiRequestError,
} from "../api";
import { InterestSchedule, Frequency } from "../types";
import Card from "./ui/Card";
import Input from "./ui/Input";
import Select from "./ui/Select";
import Button from "./ui/Button";
import { Clock } from "lucide-react";

interface InterestScheduleFormProps {
  childId: number;
  childName: string;
  schedule: InterestSchedule | null;
  onUpdated: (schedule: InterestSchedule | null) => void;
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

export default function InterestScheduleForm({
  childId,
  childName,
  schedule,
  onUpdated,
}: InterestScheduleFormProps) {
  const [frequency, setFrequency] = useState<Frequency>("monthly");
  const [dayOfWeek, setDayOfWeek] = useState(0);
  const [dayOfMonth, setDayOfMonth] = useState(1);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    if (schedule) {
      setFrequency(schedule.frequency);
      if (schedule.day_of_week != null) setDayOfWeek(schedule.day_of_week);
      if (schedule.day_of_month != null) setDayOfMonth(schedule.day_of_month);
    }
  }, [schedule]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    setSaving(true);

    try {
      const result = await setInterestSchedule(childId, {
        frequency,
        day_of_week: frequency !== "monthly" ? dayOfWeek : undefined,
        day_of_month: frequency === "monthly" ? dayOfMonth : undefined,
      });
      setSuccess(schedule ? "Interest schedule updated." : "Interest schedule created.");
      onUpdated(result);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to save interest schedule.");
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
      await deleteInterestSchedule(childId);
      setSuccess("Interest schedule removed.");
      onUpdated(null);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to remove interest schedule.");
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
    <Card padding="md">
      <div className="flex items-center gap-2 mb-4">
        <Clock className="h-5 w-5 text-forest" aria-hidden="true" />
        <h4 className="text-base font-bold text-bark">Interest Schedule for {childName}</h4>
      </div>

      {schedule && (
        <div className="mb-4 p-3 bg-cream rounded-xl">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <span className="text-sm text-bark-light">Status:</span>
              <span className={`
                inline-flex px-2.5 py-0.5 rounded-full text-xs font-semibold
                ${schedule.status === "active"
                  ? "bg-sage-light/40 text-forest"
                  : "bg-sand text-bark-light"
                }
              `}>
                {schedule.status === "active" ? "Active" : "Paused"}
              </span>
            </div>
            {schedule.status === "active" && schedule.next_run_at && (
              <span className="text-xs text-bark-light">
                Next: {formatNextRun(schedule.next_run_at)}
              </span>
            )}
          </div>
          <Button variant="danger" onClick={handleDelete} disabled={saving} className="text-sm !min-h-[36px] !px-3 !py-1">
            Remove Schedule
          </Button>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <Select
          label="Frequency"
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
              <option key={d.value} value={d.value}>{d.label}</option>
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

        <Button type="submit" loading={saving} className="w-full">
          {saving ? "Saving..." : schedule ? "Update Schedule" : "Set Up Interest Schedule"}
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
