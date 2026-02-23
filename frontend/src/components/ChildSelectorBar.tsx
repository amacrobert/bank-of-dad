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
                flex items-center gap-2 px-3 py-2 rounded-xl
                min-h-[44px] flex-shrink-0 max-w-[200px]
                transition-all duration-200 cursor-pointer
                ${isSelected
                  ? "bg-forest/5 ring-2 ring-forest"
                  : "bg-white border border-sand hover:bg-cream-dark"
                }
              `}
            >
              {/* Avatar */}
              <div className={`
                w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0
                ${child.avatar
                  ? "text-xl"
                  : `text-sm font-bold ${isSelected ? "bg-forest text-white" : "bg-sage-light/40 text-forest"}`
                }
                ${child.avatar && !isSelected ? "bg-cream" : ""}
                ${child.avatar && isSelected ? "bg-forest/10" : ""}
              `}>
                {child.avatar || child.first_name.charAt(0).toUpperCase()}
              </div>

              {/* Name + balance + lock */}
              <div className="min-w-0 text-left">
                <div className="flex items-center gap-1">
                  <span className="font-semibold text-bark text-sm truncate">
                    {child.first_name}
                  </span>
                  {child.is_locked && (
                    <Lock className="h-3 w-3 text-terracotta flex-shrink-0" aria-label="Account locked" />
                  )}
                </div>
                <div className="-mt-0.5">
                  <BalanceDisplay balanceCents={child.balance_cents} size="small" />
                </div>
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
