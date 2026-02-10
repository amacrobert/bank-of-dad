import { useEffect, useState } from "react";
import { get } from "../api";
import { ChildListResponse, Child } from "../types";
import BalanceDisplay from "./BalanceDisplay";
import LoadingSpinner from "./ui/LoadingSpinner";
import { ChevronRight, Lock } from "lucide-react";

interface ChildListProps {
  refreshKey: number;
  onSelectChild?: (child: Child) => void;
  selectedChildId?: number | null;
}

export default function ChildList({ refreshKey, onSelectChild, selectedChildId }: ChildListProps) {
  const [children, setChildren] = useState<Child[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    get<ChildListResponse>("/children")
      .then((data) => {
        setChildren(data.children || []);
        setLoading(false);
      })
      .catch(() => {
        setLoading(false);
      });
  }, [refreshKey]);

  if (loading) {
    return <LoadingSpinner variant="inline" message="Loading children..." />;
  }

  if (children.length === 0) {
    return (
      <div className="py-8 text-center">
        <p className="text-bark-light mb-1">No children added yet.</p>
        <p className="text-sm text-bark-light/70">Add your first child above!</p>
      </div>
    );
  }

  return (
    <div>
      <h3 className="text-base font-bold text-bark mb-3">Children</h3>
      <div className="space-y-2">
        {children.map((child) => {
          const isSelected = selectedChildId === child.id;
          return (
            <button
              key={child.id}
              onClick={() => onSelectChild?.(child)}
              className={`
                w-full flex items-center gap-3 p-3 rounded-xl
                transition-all duration-200 text-left cursor-pointer
                ${isSelected
                  ? "bg-forest/5 ring-2 ring-forest"
                  : "bg-white border border-sand hover:bg-cream-dark"
                }
              `}
            >
              {/* Avatar circle */}
              <div className={`
                w-10 h-10 rounded-full flex items-center justify-center text-base font-bold flex-shrink-0
                ${isSelected ? "bg-forest text-white" : "bg-sage-light/40 text-forest"}
              `}>
                {child.first_name.charAt(0).toUpperCase()}
              </div>

              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-1.5">
                  <span className="font-semibold text-bark text-sm truncate">{child.first_name}</span>
                  {child.is_locked && (
                    <Lock className="h-3.5 w-3.5 text-terracotta flex-shrink-0" aria-label="Account locked" />
                  )}
                </div>
                <div className="mt-0.5">
                  <BalanceDisplay balanceCents={child.balance_cents} size="small" />
                </div>
              </div>

              <ChevronRight className="h-5 w-5 text-bark-light/50 flex-shrink-0" aria-hidden="true" />
            </button>
          );
        })}
      </div>
    </div>
  );
}
