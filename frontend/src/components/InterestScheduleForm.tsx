import { useState, useEffect } from "react";
import {
  setInterestSchedule,
  deleteInterestSchedule,
  ApiRequestError,
} from "../api";
import { InterestSchedule, Frequency } from "../types";

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
    <div className="interest-schedule-form">
      <h4>Interest Schedule for {childName}</h4>

      {schedule && (
        <div className="schedule-status">
          <p>
            Status: <strong>{schedule.status === "active" ? "Active" : "Paused"}</strong>
            {schedule.status === "active" && schedule.next_run_at && (
              <> &mdash; Next accrual: {formatNextRun(schedule.next_run_at)}</>
            )}
          </p>
          <button type="button" onClick={handleDelete} disabled={saving} className="btn-danger">
            Remove Schedule
          </button>
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="interest-frequency">Frequency</label>
          <select
            id="interest-frequency"
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
            <label htmlFor="interest-day-of-week">Day of Week</label>
            <select
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
            </select>
          </div>
        )}

        {frequency === "monthly" && (
          <div className="form-group">
            <label htmlFor="interest-day-of-month">Day of Month</label>
            <input
              type="number"
              id="interest-day-of-month"
              value={dayOfMonth}
              onChange={(e) => setDayOfMonth(Number(e.target.value))}
              min={1}
              max={31}
              required
              disabled={saving}
            />
          </div>
        )}

        <button type="submit" disabled={saving} className="btn-primary">
          {saving ? "Saving..." : schedule ? "Update Schedule" : "Set Up Interest Schedule"}
        </button>
      </form>

      {success && <p className="success">{success}</p>}
      {error && <p className="error">{error}</p>}
    </div>
  );
}
