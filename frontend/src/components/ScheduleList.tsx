import { useEffect, useState } from "react";
import { listSchedules, pauseSchedule, resumeSchedule, deleteSchedule, ApiRequestError } from "../api";
import { ScheduleWithChild } from "../types";
import BalanceDisplay from "./BalanceDisplay";

interface ScheduleListProps {
  refreshKey: number;
}

const DAY_NAMES = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];

function formatFrequency(sched: ScheduleWithChild): string {
  if (sched.frequency === "weekly" && sched.day_of_week !== undefined) {
    return `Weekly on ${DAY_NAMES[sched.day_of_week]}`;
  }
  if (sched.frequency === "biweekly" && sched.day_of_week !== undefined) {
    return `Every 2 weeks on ${DAY_NAMES[sched.day_of_week]}`;
  }
  if (sched.frequency === "monthly" && sched.day_of_month !== undefined) {
    return `Monthly on the ${sched.day_of_month}${ordinalSuffix(sched.day_of_month)}`;
  }
  return sched.frequency;
}

function ordinalSuffix(n: number): string {
  if (n >= 11 && n <= 13) return "th";
  switch (n % 10) {
    case 1: return "st";
    case 2: return "nd";
    case 3: return "rd";
    default: return "th";
  }
}

export default function ScheduleList({ refreshKey }: ScheduleListProps) {
  const [schedules, setSchedules] = useState<ScheduleWithChild[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchSchedules = () => {
    setLoading(true);
    listSchedules()
      .then((data) => {
        setSchedules(data.schedules || []);
        setLoading(false);
      })
      .catch(() => {
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchSchedules();
  }, [refreshKey]);

  const handleToggleStatus = async (sched: ScheduleWithChild) => {
    setError(null);
    try {
      if (sched.status === "active") {
        await pauseSchedule(sched.id);
      } else {
        await resumeSchedule(sched.id);
      }
      fetchSchedules();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to update schedule.");
      }
    }
  };

  const handleDelete = async (sched: ScheduleWithChild) => {
    if (!confirm(`Delete the allowance schedule for ${sched.child_first_name}?`)) return;
    setError(null);
    try {
      await deleteSchedule(sched.id);
      fetchSchedules();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to delete schedule.");
      }
    }
  };

  if (loading) {
    return <p>Loading schedules...</p>;
  }

  if (schedules.length === 0) {
    return null;
  }

  return (
    <div className="schedule-list">
      <h3>Allowance Schedules</h3>
      {error && <p className="error">{error}</p>}
      <ul>
        {schedules.map((sched) => (
          <li key={sched.id} className={`schedule-item ${sched.status === "paused" ? "paused" : ""}`}>
            <div className="schedule-info">
              <span className="schedule-child">{sched.child_first_name}</span>
              <BalanceDisplay balanceCents={sched.amount_cents} />
              <span className="schedule-frequency">{formatFrequency(sched)}</span>
              {sched.note && <span className="schedule-note">{sched.note}</span>}
              {sched.status === "paused" && <span className="schedule-status">Paused</span>}
            </div>
            <div className="schedule-actions">
              <button
                onClick={() => handleToggleStatus(sched)}
                className="btn-secondary"
              >
                {sched.status === "active" ? "Pause" : "Resume"}
              </button>
              <button
                onClick={() => handleDelete(sched)}
                className="btn-danger"
              >
                Delete
              </button>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
