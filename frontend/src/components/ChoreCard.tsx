import { ChoreInstance } from "../types";
import Card from "./ui/Card";
import Button from "./ui/Button";
import { CheckCircle, Clock, AlertCircle } from "lucide-react";

interface ChoreCardProps {
  instance: ChoreInstance;
  onComplete?: (id: number) => void;
  loading?: boolean;
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function StatusBadge({ status }: { status: ChoreInstance["status"] }) {
  switch (status) {
    case "available":
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-forest/10 text-forest">
          <CheckCircle className="h-3 w-3" aria-hidden="true" />
          Available
        </span>
      );
    case "pending_approval":
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-honey text-bark">
          <Clock className="h-3 w-3" aria-hidden="true" />
          Pending
        </span>
      );
    case "approved":
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-sky-100 text-sky-700">
          <CheckCircle className="h-3 w-3" aria-hidden="true" />
          Done
        </span>
      );
    default:
      return null;
  }
}

export default function ChoreCard({ instance, onComplete, loading }: ChoreCardProps) {
  return (
    <Card padding="md">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <h3 className="font-bold text-bark">{instance.chore_name}</h3>
            <StatusBadge status={instance.status} />
          </div>

          {instance.chore_description && (
            <p className="text-sm text-bark-light mt-1">{instance.chore_description}</p>
          )}

          <p className="text-lg font-semibold text-forest mt-1">
            ${(instance.reward_cents / 100).toFixed(2)}
          </p>

          {instance.period_start && instance.period_end && (
            <p className="text-xs text-bark-light mt-1">
              {formatDate(instance.period_start)} &ndash; {formatDate(instance.period_end)}
            </p>
          )}

          {instance.rejection_reason && (
            <div className="flex items-center gap-1.5 mt-2 text-sm text-terracotta">
              <AlertCircle className="h-4 w-4 flex-shrink-0" aria-hidden="true" />
              <span>{instance.rejection_reason}</span>
            </div>
          )}
        </div>

        {instance.status === "available" && onComplete && (
          <Button
            onClick={() => onComplete(instance.id)}
            disabled={loading}
            loading={loading}
            className="flex-shrink-0"
          >
            Mark Done
          </Button>
        )}
      </div>
    </Card>
  );
}
