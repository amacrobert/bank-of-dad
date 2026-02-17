import { useEffect, useState } from "react";
import { getUpcomingAllowances, getInterestSchedule } from "../api";
import type { Transaction, Frequency } from "../types";
import { ArrowDownCircle, ArrowUpCircle, Calendar, TrendingUp } from "lucide-react";
import Card from "./ui/Card";
import LoadingSpinner from "./ui/LoadingSpinner";

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

function formatRecentDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
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
            note: undefined,
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
      <div className="mb-5">
        <h4 className="text-sm font-semibold text-bark-light uppercase tracking-wide mb-3">
          Upcoming
        </h4>
        {loadingUpcoming ? (
          <LoadingSpinner variant="inline" message="Loading payments..." />
        ) : upcomingPayments.length === 0 ? (
          <p className="text-sm text-bark-light">No upcoming payments</p>
        ) : (
          <div className="space-y-3">
            {upcomingPayments.map((p, i) => (
              <div key={i} className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-bark">
                    {formatUpcomingAmount(p.amountCents, p.estimated)}
                  </span>
                  <span
                    className={`text-xs px-1.5 py-0.5 rounded-full font-medium ${
                      p.type === "interest"
                        ? "bg-sage-light/30 text-forest"
                        : "bg-sky-100 text-sky-700"
                    }`}
                  >
                    {p.type === "interest" ? (
                      <span className="inline-flex items-center gap-0.5">
                        <TrendingUp className="h-3 w-3" aria-hidden="true" />
                        Interest
                      </span>
                    ) : (
                      "Allowance"
                    )}
                  </span>
                  {p.note && (
                    <span className="text-xs text-bark-light">{p.note}</span>
                  )}
                </div>
                <span className="text-xs text-bark-light">
                  {new Date(p.date).toLocaleDateString(undefined, {
                    month: "short",
                    day: "numeric",
                  })}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Divider */}
      <div className="border-t border-sand/60 mb-5" />

      {/* Recent section */}
      <div>
        <h4 className="text-sm font-semibold text-bark-light uppercase tracking-wide mb-3">
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
                        {formatRecentDate(tx.created_at)}
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
