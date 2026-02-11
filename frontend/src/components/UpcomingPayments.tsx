import { useEffect, useState } from "react";
import { getUpcomingAllowances, getInterestSchedule } from "../api";
import type { Frequency } from "../types";
import { Calendar, TrendingUp } from "lucide-react";
import Card from "./ui/Card";
import LoadingSpinner from "./ui/LoadingSpinner";

interface UpcomingPaymentsProps {
  childId: number;
  balanceCents: number;
  interestRateBps: number;
}

interface UpcomingPayment {
  type: "allowance" | "interest";
  amountCents: number;
  estimated: boolean;
  date: string;
  note?: string;
}

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

function formatAmount(cents: number, estimated: boolean): string {
  const formatted = `$${(cents / 100).toFixed(2)}`;
  return estimated ? `~${formatted}` : formatted;
}

function cardTitle(hasAllowance: boolean, hasInterest: boolean): string {
  if (hasAllowance && hasInterest) return "Upcoming allowance and interest";
  if (hasInterest) return "Upcoming interest payment";
  return "Upcoming allowance";
}

export default function UpcomingPayments({
  childId,
  balanceCents,
  interestRateBps,
}: UpcomingPaymentsProps) {
  const [payments, setPayments] = useState<UpcomingPayment[]>([]);
  const [hasAllowance, setHasAllowance] = useState(false);
  const [hasInterest, setHasInterest] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
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

        // Sort by date
        items.sort(
          (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()
        );

        setPayments(items);
        setHasAllowance(allowances.length > 0);
        setHasInterest(!!activeInterest && interestRateBps > 0);
      })
      .catch(() => {
        // Silently fail
      })
      .finally(() => {
        setLoading(false);
      });
  }, [childId, balanceCents, interestRateBps]);

  if (loading) {
    return <LoadingSpinner variant="inline" message="Loading payments..." />;
  }

  if (payments.length === 0) {
    return null;
  }

  return (
    <Card padding="md">
      <div className="flex items-center gap-2 mb-4">
        <Calendar className="h-5 w-5 text-forest" aria-hidden="true" />
        <h3 className="text-base font-bold text-bark">
          {cardTitle(hasAllowance, hasInterest)}
        </h3>
      </div>
      <div className="space-y-3">
        {payments.map((p, i) => (
          <div key={i} className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-bark">
                {formatAmount(p.amountCents, p.estimated)}
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
    </Card>
  );
}
