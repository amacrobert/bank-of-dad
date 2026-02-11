import { Transaction } from "../types";
import { ArrowDownCircle, ArrowUpCircle, Calendar, TrendingUp } from "lucide-react";

interface TransactionHistoryProps {
  transactions: Transaction[];
}

const typeConfig: Record<string, { icon: typeof ArrowDownCircle; color: string; amountColor: string }> = {
  deposit: { icon: ArrowDownCircle, color: "text-forest", amountColor: "text-forest" },
  withdrawal: { icon: ArrowUpCircle, color: "text-terracotta", amountColor: "text-terracotta" },
  allowance: { icon: Calendar, color: "text-forest", amountColor: "text-forest" },
  interest: { icon: TrendingUp, color: "text-amber", amountColor: "text-forest" },
};

export default function TransactionHistory({ transactions }: TransactionHistoryProps) {
  if (transactions.length === 0) {
    return (
      <div className="py-8 text-center">
        <p className="text-bark-light">No transactions yet.</p>
      </div>
    );
  }

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  const formatAmount = (cents: number, type: string) => {
    const dollars = (cents / 100).toFixed(2);
    return type === "withdrawal" ? `-$${dollars}` : `+$${dollars}`;
  };

  const getTypeLabel = (type: string) => {
    switch (type) {
      case "deposit": return "Deposit";
      case "withdrawal": return "Withdrawal";
      case "allowance": return "Allowance";
      case "interest": return "Interest earned";
      default: return type;
    }
  };

  return (
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
                  {formatAmount(tx.amount_cents, tx.type)}
                </span>
              </div>
              <div className="flex items-baseline justify-between gap-2 mt-0.5">
                <span className="text-xs text-bark-light truncate">
                  {tx.note || "\u00A0"}
                </span>
                <span className="text-xs text-bark-light/70 whitespace-nowrap">
                  {formatDate(tx.created_at)}
                </span>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
