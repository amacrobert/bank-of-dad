import { useState, useRef, useEffect } from "react";
import { SavingsGoal } from "../types";
import Card from "./ui/Card";
import GoalProgressRing from "./GoalProgressRing";
import { Target, Sparkles, CheckCircle2, EllipsisVertical } from "lucide-react";

interface GoalCardProps {
  goal: SavingsGoal;
  childId: number;
  onAllocate?: (goalId: number, amountCents: number) => Promise<void>;
  onEdit?: (goal: SavingsGoal) => void;
  onDelete?: (goal: SavingsGoal) => void;
}

function formatCents(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}

export default function GoalCard({ goal, onAllocate, onEdit, onDelete }: GoalCardProps) {
  const [showAllocate, setShowAllocate] = useState(false);
  const [allocateMode, setAllocateMode] = useState<"add" | "remove">("add");
  const [amountStr, setAmountStr] = useState("");
  const [allocating, setAllocating] = useState(false);
  const [error, setError] = useState("");
  const [pulse, setPulse] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menuOpen) return;
    const handleClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [menuOpen]);

  const percent = goal.target_cents > 0
    ? Math.min(100, Math.round((goal.saved_cents / goal.target_cents) * 100))
    : 0;

  const isOverdue = goal.status === "active" && goal.target_date && new Date(goal.target_date) < new Date();

  const handleAllocateSubmit = async () => {
    setError("");
    const dollars = parseFloat(amountStr);
    if (isNaN(dollars) || dollars <= 0) {
      setError("Enter a valid amount");
      return;
    }
    const cents = Math.round(dollars * 100);
    const finalAmount = allocateMode === "remove" ? -cents : cents;

    if (!onAllocate) return;

    setAllocating(true);
    try {
      await onAllocate(goal.id, finalAmount);
      setShowAllocate(false);
      setAmountStr("");
      setAllocateMode("add");
      // Trigger pulse animation
      setPulse(true);
      setTimeout(() => setPulse(false), 600);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Failed to allocate funds.");
      }
    } finally {
      setAllocating(false);
    }
  };

  const openAllocate = (mode: "add" | "remove") => {
    setAllocateMode(mode);
    setAmountStr("");
    setError("");
    setShowAllocate(true);
  };

  return (
    <Card padding="sm" className="space-y-3">
      <div className="flex items-center gap-4">
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
          <div className="flex items-center gap-1.5">
            <p className="font-semibold text-bark truncate">{goal.name}</p>
            {percent >= 75 && goal.status === "active" && (
              <Sparkles className="h-4 w-4 text-amber flex-shrink-0" />
            )}
          </div>
          <p className="text-sm text-bark-light">
            {formatCents(goal.saved_cents)} / {formatCents(goal.target_cents)}
          </p>
          {isOverdue && (
            <span className="text-xs text-terracotta font-medium">Overdue</span>
          )}
          {goal.status === "completed" && goal.completed_at && (
            <div className="flex items-center gap-1 text-xs text-forest">
              <CheckCircle2 className="h-3 w-3" />
              <span>Achieved {new Date(goal.completed_at).toLocaleDateString()}</span>
            </div>
          )}
        </div>

        {/* Progress ring */}
        <div className={pulse ? "animate-pulse" : ""}>
          <GoalProgressRing percent={percent} size={48} strokeWidth={5} milestone={percent >= 75} />
        </div>

        {/* Actions for active goals */}
        {goal.status === "active" && (onAllocate || onEdit || onDelete) && (
          <div className="flex items-center gap-1">
            {onAllocate && (
              <button
                onClick={() => openAllocate("add")}
                className="text-xs text-forest font-medium hover:underline"
              >
                Add Funds
              </button>
            )}
            {(onAllocate && goal.saved_cents > 0) || onEdit || onDelete ? (
              <div className="relative" ref={menuRef}>
                <button
                  onClick={() => setMenuOpen(!menuOpen)}
                  className="p-1 text-bark-light hover:text-bark rounded-md hover:bg-sand/50"
                >
                  <EllipsisVertical className="h-4 w-4" />
                </button>
                {menuOpen && (
                  <div className="absolute right-0 top-full mt-1 bg-white border border-sand rounded-lg shadow-lg py-1 z-10 min-w-[120px]">
                    {onAllocate && goal.saved_cents > 0 && (
                      <button
                        onClick={() => { setMenuOpen(false); openAllocate("remove"); }}
                        className="w-full text-left px-3 py-1.5 text-xs text-bark hover:bg-sand/50"
                      >
                        Remove Funds
                      </button>
                    )}
                    {onEdit && (
                      <button
                        onClick={() => { setMenuOpen(false); onEdit(goal); }}
                        className="w-full text-left px-3 py-1.5 text-xs text-bark hover:bg-sand/50"
                      >
                        Edit
                      </button>
                    )}
                    {onDelete && (
                      <button
                        onClick={() => { setMenuOpen(false); onDelete(goal); }}
                        className="w-full text-left px-3 py-1.5 text-xs text-terracotta hover:bg-sand/50"
                      >
                        Delete
                      </button>
                    )}
                  </div>
                )}
              </div>
            ) : null}
          </div>
        )}

        {/* Delete action for completed goals */}
        {goal.status === "completed" && onDelete && (
          <div className="relative" ref={menuRef}>
            <button
              onClick={() => setMenuOpen(!menuOpen)}
              className="p-1 text-bark-light hover:text-bark rounded-md hover:bg-sand/50"
            >
              <EllipsisVertical className="h-4 w-4" />
            </button>
            {menuOpen && (
              <div className="absolute right-0 top-full mt-1 bg-white border border-sand rounded-lg shadow-lg py-1 z-10 min-w-[120px]">
                <button
                  onClick={() => { setMenuOpen(false); onDelete(goal); }}
                  className="w-full text-left px-3 py-1.5 text-xs text-terracotta hover:bg-sand/50"
                >
                  Delete
                </button>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Inline allocation form */}
      {showAllocate && (
        <div className="flex items-center gap-2 pt-1">
          <span className="text-sm text-bark-light font-medium">
            {allocateMode === "add" ? "Add:" : "Remove:"}
          </span>
          <div className="relative flex-1">
            <span className="absolute left-2 top-1/2 -translate-y-1/2 text-sm text-bark-light">$</span>
            <input
              type="number"
              step="0.01"
              min="0.01"
              value={amountStr}
              onChange={(e) => setAmountStr(e.target.value)}
              placeholder="0.00"
              className="w-full pl-6 pr-2 py-1.5 text-sm border border-sand rounded-lg focus:outline-none focus:ring-2 focus:ring-forest/30"
              autoFocus
              onKeyDown={(e) => {
                if (e.key === "Enter") handleAllocateSubmit();
                if (e.key === "Escape") setShowAllocate(false);
              }}
            />
          </div>
          <button
            onClick={handleAllocateSubmit}
            disabled={allocating}
            className="px-3 py-1.5 text-sm font-medium text-white bg-forest rounded-lg hover:bg-forest/90 disabled:opacity-50"
          >
            {allocating ? "..." : "Go"}
          </button>
          <button
            onClick={() => setShowAllocate(false)}
            className="px-2 py-1.5 text-sm text-bark-light hover:text-bark"
          >
            Cancel
          </button>
        </div>
      )}

      {/* Error */}
      {error && (
        <p className="text-xs text-terracotta">{error}</p>
      )}
    </Card>
  );
}
