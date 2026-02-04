import { Transaction } from "../types";

interface TransactionHistoryProps {
  transactions: Transaction[];
}

export default function TransactionHistory({ transactions }: TransactionHistoryProps) {
  if (transactions.length === 0) {
    return (
      <div className="transaction-history-empty">
        <p>No transactions yet.</p>
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
    return type === "deposit" ? `+$${dollars}` : `-$${dollars}`;
  };

  return (
    <div className="transaction-history">
      <ul className="transaction-list">
        {transactions.map((tx) => (
          <li key={tx.id} className={`transaction-item ${tx.type}`}>
            <div className="transaction-date">{formatDate(tx.created_at)}</div>
            <div className="transaction-details">
              <span className={`transaction-amount ${tx.type}`}>
                {formatAmount(tx.amount_cents, tx.type)}
              </span>
              {tx.note && <span className="transaction-note">{tx.note}</span>}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
