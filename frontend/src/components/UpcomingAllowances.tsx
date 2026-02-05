import { useEffect, useState } from "react";
import { getUpcomingAllowances } from "../api";
import { UpcomingAllowance } from "../types";
import BalanceDisplay from "./BalanceDisplay";

interface UpcomingAllowancesProps {
  childId: number;
}

export default function UpcomingAllowances({ childId }: UpcomingAllowancesProps) {
  const [allowances, setAllowances] = useState<UpcomingAllowance[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getUpcomingAllowances(childId)
      .then((data) => {
        setAllowances(data.allowances || []);
      })
      .catch(() => {
        // Silently fail
      })
      .finally(() => {
        setLoading(false);
      });
  }, [childId]);

  if (loading) {
    return <p>Loading...</p>;
  }

  if (allowances.length === 0) {
    return null;
  }

  return (
    <div className="upcoming-allowances">
      <h3>Upcoming Allowances</h3>
      <ul>
        {allowances.map((a, i) => (
          <li key={i} className="upcoming-item">
            <BalanceDisplay balanceCents={a.amount_cents} />
            <span className="upcoming-date">
              {new Date(a.next_date).toLocaleDateString()}
            </span>
            {a.note && <span className="upcoming-note">{a.note}</span>}
          </li>
        ))}
      </ul>
    </div>
  );
}
