import { useEffect, useState } from "react";
import { getUpcomingAllowances } from "../api";
import { UpcomingAllowance } from "../types";
import { Calendar } from "lucide-react";
import Card from "./ui/Card";
import LoadingSpinner from "./ui/LoadingSpinner";

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
    return <LoadingSpinner variant="inline" message="Loading allowances..." />;
  }

  if (allowances.length === 0) {
    return null;
  }

  return (
    <Card padding="md">
      <div className="flex items-center gap-2 mb-4">
        <Calendar className="h-5 w-5 text-forest" aria-hidden="true" />
        <h3 className="text-base font-bold text-bark">Upcoming Allowances</h3>
      </div>
      <div className="space-y-3">
        {allowances.map((a, i) => (
          <div key={i} className="flex items-center justify-between">
            <div>
              <span className="text-sm font-medium text-bark">
                ${(a.amount_cents / 100).toFixed(2)}
              </span>
              {a.note && (
                <span className="text-xs text-bark-light ml-2">{a.note}</span>
              )}
            </div>
            <span className="text-xs text-bark-light">
              {new Date(a.next_date).toLocaleDateString(undefined, {
                month: "short",
                day: "numeric",
              })}
            </span>
          </div>
        ))}
      </div>
    </Card>
  );
}
