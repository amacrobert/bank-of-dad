import { useEffect, useState } from "react";
import { getUpcomingAllowances, getInterestSchedule } from "../api";
import type { Transaction, Frequency } from "../types";
import { ArrowDownCircle, ArrowUpCircle, Calendar, TrendingUp } from "lucide-react";
import Card from "./ui/Card";
import LoadingSpinner from "./ui/LoadingSpinner";
import { useTimezone } from "../context/TimezoneContext";

interface TransactionsCardProps {
  childId: number;
  balanceCents: number;
  interestRateBps: number;
  transactions: Transaction[];
}

interface UpcomingPayment {
  type: "allowance" | "interest";
  amountCents: number;
  estimated: boolean;
  date: string;
  note?: string;
}

// --- Upcoming helpers ---

function periodsPerYear(frequency: Frequency): number {
  switch (frequency) {
    case "weekly":
      return 52;
    case "biweekly":
      return 26;
    case "monthly":
      return 12;
  }
}

function formatInterestNote(rateBps: number, frequency: Frequency): string {
  const rate = rateBps / 100;
  const rateStr = rate % 1 === 0 ? rate.toFixed(0) : rate.toString();
  return `${rateStr}% annual interest compounded ${frequency}`;
}

function estimateInterestCents(
  balanceCents: number,
  rateBps: number,
  frequency: Frequency
): number {
  return Math.round(
    (balanceCents * rateBps) / periodsPerYear(frequency) / 10000
  );
}

function formatUpcomingAmount(cents: number, estimated: boolean): string {
  const formatted = `$${(cents / 100).toFixed(2)}`;
  return estimated ? `~${formatted}` : formatted;
}

// --- Recent helpers ---

const typeConfig: Record<string, { icon: typeof ArrowDownCircle; color: string; amountColor: string }> = {
  deposit: { icon: ArrowDownCircle, color: "text-forest", amountColor: "text-forest" },
  withdrawal: { icon: ArrowUpCircle, color: "text-terracotta", amountColor: "text-terracotta" },
  allowance: { icon: Calendar, color: "text-forest", amountColor: "text-forest" },
  interest: { icon: TrendingUp, color: "text-amber", amountColor: "text-forest" },
};

function formatRecentDate(dateStr: string, timeZone: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    timeZone,
  });
}

function formatRecentAmount(cents: number, type: string): string {
  const dollars = (cents / 100).toFixed(2);
  return type === "withdrawal" ? `-$${dollars}` : `+$${dollars}`;
}

function getTypeLabel(type: string): string {
  switch (type) {
    case "deposit": return "Deposit";
    case "withdrawal": return "Withdrawal";
    case "allowance": return "Allowance";
    case "interest": return "Interest earned";
    default: return type;
  }
}

export default function TransactionsCard({
  childId,
  balanceCents,
  interestRateBps,
  transactions,
}: TransactionsCardProps) {
  const timezone = useTimezone();
  const [upcomingPayments, setUpcomingPayments] = useState<UpcomingPayment[]>([]);
  const [loadingUpcoming, setLoadingUpcoming] = useState(true);

  useEffect(() => {
    setLoadingUpcoming(true);
    Promise.all([
      getUpcomingAllowances(childId),
      interestRateBps > 0
        ? getInterestSchedule(childId)
        : Promise.resolve(null),
    ])
      .then(([allowanceRes, interestSched]) => {
        const items: UpcomingPayment[] = [];

        const allowances = allowanceRes.allowances || [];
        for (const a of allowances) {
          items.push({
            type: "allowance",
            amountCents: a.amount_cents,
            estimated: false,
            date: a.next_date,
            note: a.note,
          });
        }

        const activeInterest =
          interestSched &&
          interestSched.status === "active" &&
          interestSched.next_run_at;

        if (activeInterest && interestRateBps > 0) {
          items.push({
            type: "interest",
            amountCents: estimateInterestCents(
              balanceCents,
              interestRateBps,
              interestSched.frequency
            ),
            estimated: true,
            date: interestSched.next_run_at!,
            note: formatInterestNote(interestRateBps, interestSched.frequency),
          });
        }

        items.sort(
          (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()
        );

        setUpcomingPayments(items);
      })
      .catch(() => {
        // Silently fail — show empty upcoming section
      })
      .finally(() => {
        setLoadingUpcoming(false);
      });
  }, [childId, balanceCents, interestRateBps]);

  return (
    <Card padding="md">
      <h3 className="text-base font-bold text-bark mb-4">Transactions</h3>

      {/* Upcoming section */}
      {(loadingUpcoming || upcomingPayments.length > 0) && (
        <div className="mb-5">
          <h4 className="text-sm font-bold text-bark-light uppercase tracking-wide mb-3">
            Upcoming
          </h4>
          {loadingUpcoming ? (
            <LoadingSpinner variant="inline" message="Loading payments..." />
          ) : (
            <div className="space-y-0">
              {upcomingPayments.map((p, i) => {
                const config = typeConfig[p.type] || typeConfig.deposit;
                const Icon = config.icon;
                return (
                  <div
                    key={i}
                    className="flex items-center gap-3 py-3 border-b border-sand/60 last:border-b-0"
                  >
                    <div className={`flex-shrink-0 ${config.color}`}>
                      <Icon className="h-5 w-5" aria-hidden="true" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-baseline justify-between gap-2">
                        <span className="text-sm font-medium text-bark truncate">
                          {getTypeLabel(p.type)}
                        </span>
                        <span className="text-sm font-bold tabular-nums text-forest">
                          +{formatUpcomingAmount(p.amountCents, p.estimated)}
                        </span>
                      </div>
                      <div className="flex items-baseline justify-between gap-2 mt-0.5">
                        <span className="text-xs text-bark-light truncate">
                          {p.note || "\u00A0"}
                        </span>
                        <span className="text-xs text-bark-light/70 whitespace-nowrap">
                          {new Date(p.date).toLocaleDateString(undefined, {
                            month: "short",
                            day: "numeric",
                            year: "numeric",
                            timeZone: timezone,
                          })}
                        </span>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}

      {/* Recent section */}
      <div>
        <h4 className="text-sm font-bold text-bark-light uppercase tracking-wide mb-3">
          Recent
        </h4>
        {transactions.length === 0 ? (
          <div className="py-4 text-center">
            <p className="text-sm text-bark-light">No transactions yet.</p>
          </div>
        ) : (
          <div className="space-y-0">
            {transactions.map((tx) => {
              const config = typeConfig[tx.type] || typeConfig.deposit;
              const Icon = config.icon;
              return (
                <div
                  key={tx.id}
                  className="flex items-center gap-3 py-3 border-b border-sand/60 last:border-b-0"
                >
                  <div className={`flex-shrink-0 ${config.color}`}>
                    <Icon className="h-5 w-5" aria-hidden="true" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-baseline justify-between gap-2">
                      <span className="text-sm font-medium text-bark truncate">
                        {getTypeLabel(tx.type)}
                      </span>
                      <span className={`text-sm font-bold tabular-nums ${config.amountColor}`}>
                        {formatRecentAmount(tx.amount_cents, tx.type)}
                      </span>
                    </div>
                    <div className="flex items-baseline justify-between gap-2 mt-0.5">
                      <span className="text-xs text-bark-light truncate">
                        {tx.note || "\u00A0"}
                      </span>
                      <span className="text-xs text-bark-light/70 whitespace-nowrap">
                        {formatRecentDate(tx.created_at, timezone)}
                      </span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </Card>
  );
}
