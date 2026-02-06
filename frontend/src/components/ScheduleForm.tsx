import { useEffect, useState } from "react";
import { createSchedule, ApiRequestError } from "../api";
import { get } from "../api";
import { Child, ChildListResponse, Frequency } from "../types";

interface ScheduleFormProps {
  onCreated: () => void;
  onCancel: () => void;
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

export default function ScheduleForm({ onCreated, onCancel }: ScheduleFormProps) {
  const [children, setChildren] = useState<Child[]>([]);
  const [childId, setChildId] = useState<number | "">("");
  const [amount, setAmount] = useState("");
  const [frequency, setFrequency] = useState<Frequency>("weekly");
  const [dayOfWeek, setDayOfWeek] = useState(5); // Friday default
  const [dayOfMonth, setDayOfMonth] = useState(1);
  const [note, setNote] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    get<ChildListResponse>("/children").then((data) => {
      const list = data.children || [];
      setChildren(list);
      if (list.length === 1) {
        setChildId(list[0].id);
      }
    });
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (childId === "") {
      setError("Please select a child.");
      return;
    }

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

    setLoading(true);
    try {
      await createSchedule({
        child_id: childId as number,
        amount_cents: amountCents,
        frequency,
        day_of_week: frequency !== "monthly" ? dayOfWeek : undefined,
        day_of_month: frequency === "monthly" ? dayOfMonth : undefined,
        note: note.trim() || undefined,
      });
      onCreated();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to create schedule. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="schedule-form">
      <h4>Set Up Allowance</h4>

      {error && <div className="error-message">{error}</div>}

      <div className="form-group">
        <label htmlFor="schedule-child">Child</label>
        <select
          id="schedule-child"
          value={childId}
          onChange={(e) => setChildId(e.target.value ? Number(e.target.value) : "")}
          required
          disabled={loading}
        >
          <option value="">Select a child...</option>
          {children.map((c) => (
            <option key={c.id} value={c.id}>
              {c.first_name}
            </option>
          ))}
        </select>
      </div>

      <div className="form-group">
        <label htmlFor="schedule-amount">Amount ($)</label>
        <input
          type="number"
          id="schedule-amount"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="0.00"
          step="0.01"
          min="0.01"
          max="999999.99"
          required
          disabled={loading}
        />
      </div>

      <div className="form-group">
        <label htmlFor="schedule-frequency">Frequency</label>
        <select
          id="schedule-frequency"
          value={frequency}
          onChange={(e) => setFrequency(e.target.value as Frequency)}
          disabled={loading}
        >
          <option value="weekly">Weekly</option>
          <option value="biweekly">Every 2 Weeks</option>
          <option value="monthly">Monthly</option>
        </select>
      </div>

      {frequency !== "monthly" && (
        <div className="form-group">
          <label htmlFor="schedule-day-of-week">Day of Week</label>
          <select
            id="schedule-day-of-week"
            value={dayOfWeek}
            onChange={(e) => setDayOfWeek(Number(e.target.value))}
            disabled={loading}
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
          <label htmlFor="schedule-day-of-month">Day of Month</label>
          <input
            type="number"
            id="schedule-day-of-month"
            value={dayOfMonth}
            onChange={(e) => setDayOfMonth(Number(e.target.value))}
            min={1}
            max={31}
            required
            disabled={loading}
          />
        </div>
      )}

      <div className="form-group">
        <label htmlFor="schedule-note">Note (optional)</label>
        <input
          type="text"
          id="schedule-note"
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="e.g., Weekly allowance"
          maxLength={500}
          disabled={loading}
        />
      </div>

      <div className="form-actions">
        <button type="submit" disabled={loading} className="btn-primary">
          {loading ? "Creating..." : "Create Schedule"}
        </button>
        <button type="button" onClick={onCancel} disabled={loading} className="btn-secondary">
          Cancel
        </button>
      </div>
    </form>
  );
}
