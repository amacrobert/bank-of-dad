import { SavingsGoal } from "../types";
import Card from "./ui/Card";
import GoalProgressRing from "./GoalProgressRing";
import { Target } from "lucide-react";

interface GoalCardProps {
  goal: SavingsGoal;
  onAllocate?: (goal: SavingsGoal) => void;
  onEdit?: (goal: SavingsGoal) => void;
  onDelete?: (goal: SavingsGoal) => void;
}

function formatCents(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

export default function GoalCard({ goal, onAllocate, onEdit, onDelete }: GoalCardProps) {
  const percent = goal.target_cents > 0
    ? Math.min(100, Math.round((goal.saved_cents / goal.target_cents) * 100))
    : 0;

  return (
    <Card padding="sm" className="flex items-center gap-4">
      {/* Emoji or default icon */}
      <div className="flex-shrink-0 text-2xl">
        {goal.emoji ? (
          <span>{goal.emoji}</span>
        ) : (
          <Target className="h-7 w-7 text-forest" />
        )}
      </div>

      {/* Goal info */}
      <div className="flex-1 min-w-0">
        <p className="font-semibold text-bark truncate">{goal.name}</p>
        <p className="text-sm text-bark-light">
          {formatCents(goal.saved_cents)} / {formatCents(goal.target_cents)}
        </p>
      </div>

      {/* Progress ring */}
      <GoalProgressRing percent={percent} size={48} strokeWidth={5} milestone={percent >= 75} />

      {/* Actions (shown only for active goals) */}
      {goal.status === "active" && (onAllocate || onEdit || onDelete) && (
        <div className="flex flex-col gap-1">
          {onAllocate && (
            <button
              onClick={() => onAllocate(goal)}
              className="text-xs text-forest font-medium hover:underline"
            >
              Add Funds
            </button>
          )}
          {onEdit && (
            <button
              onClick={() => onEdit(goal)}
              className="text-xs text-bark-light hover:text-bark"
            >
              Edit
            </button>
          )}
          {onDelete && (
            <button
              onClick={() => onDelete(goal)}
              className="text-xs text-terracotta hover:text-terracotta/80"
            >
              Delete
            </button>
          )}
        </div>
      )}
    </Card>
  );
}
