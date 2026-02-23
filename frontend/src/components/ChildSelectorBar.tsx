import { useRef, useState, useEffect, useCallback } from "react";
import { Child } from "../types";
import BalanceDisplay from "./BalanceDisplay";
import LoadingSpinner from "./ui/LoadingSpinner";
import { Lock } from "lucide-react";

interface ChildSelectorBarProps {
  children: Child[];
  selectedChildId: number | null;
  onSelectChild: (child: Child | null) => void;
  loading?: boolean;
}

export default function ChildSelectorBar({
  children,
  selectedChildId,
  onSelectChild,
  loading = false,
}: ChildSelectorBarProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [showLeftFade, setShowLeftFade] = useState(false);
  const [showRightFade, setShowRightFade] = useState(false);

  const updateFades = useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    const { scrollLeft, scrollWidth, clientWidth } = el;
    setShowLeftFade(scrollLeft > 4);
    setShowRightFade(scrollLeft + clientWidth < scrollWidth - 4);
  }, []);

  useEffect(() => {
    updateFades();
    const el = scrollRef.current;
    if (!el) return;
    el.addEventListener("scroll", updateFades, { passive: true });
    const ro = new ResizeObserver(updateFades);
    ro.observe(el);
    return () => {
      el.removeEventListener("scroll", updateFades);
      ro.disconnect();
    };
  }, [updateFades, children]);

  if (loading) {
    return (
      <div className="py-3">
        <LoadingSpinner variant="inline" message="Loading children..." />
      </div>
    );
  }

  if (children.length === 0) {
    return null;
  }

  const handleClick = (child: Child) => {
    if (selectedChildId === child.id) {
      onSelectChild(null);
    } else {
      onSelectChild(child);
    }
  };

  return (
    <div className="relative">
      {/* Left fade */}
      {showLeftFade && (
        <div className="absolute left-0 top-0 bottom-0 w-8 bg-gradient-to-r from-cream to-transparent z-10 pointer-events-none rounded-l-xl" />
      )}

      {/* Scrollable chip row */}
      <div
        ref={scrollRef}
        className="flex flex-nowrap gap-2 overflow-x-auto scrollbar-hide py-1 px-0.5"
        style={{ scrollbarWidth: "none", msOverflowStyle: "none" }}
      >
        {children.map((child) => {
          const isSelected = selectedChildId === child.id;
          return (
            <button
              key={child.id}
              onClick={() => handleClick(child)}
              aria-pressed={isSelected}
              className={`
                relative flex flex-col items-center justify-center
                w-[120px] aspect-square p-3 rounded-xl
                flex-shrink-0 transition-all duration-200 cursor-pointer
                ${isSelected
                  ? "bg-forest/5 ring-2 ring-forest"
                  : "bg-white border border-sand hover:border-forest hover:bg-sage-light/20"
                }
              `}
            >
              {/* Lock indicator */}
              {child.is_locked && (
                <div className="absolute top-1.5 right-1.5">
                  <Lock className="h-3 w-3 text-terracotta" aria-label="Account locked" />
                </div>
              )}

              {/* Avatar */}
              <div className={`
                w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0
                ${child.avatar
                  ? "text-2xl"
                  : `text-sm font-bold ${isSelected ? "bg-forest text-white" : "bg-sage-light/40 text-forest"}`
                }
                ${child.avatar && !isSelected ? "bg-cream" : ""}
                ${child.avatar && isSelected ? "bg-forest/10" : ""}
              `}>
                {child.avatar || child.first_name.charAt(0).toUpperCase()}
              </div>

              {/* Name */}
              <span className="font-semibold text-bark text-xs text-center truncate w-full mt-1">
                {child.first_name}
              </span>

              {/* Balance */}
              <div className="-mt-0.5">
                <BalanceDisplay balanceCents={child.balance_cents} size="small" />
              </div>
            </button>
          );
        })}
      </div>

      {/* Right fade */}
      {showRightFade && (
        <div className="absolute right-0 top-0 bottom-0 w-8 bg-gradient-to-l from-cream to-transparent z-10 pointer-events-none rounded-r-xl" />
      )}
    </div>
  );
}
