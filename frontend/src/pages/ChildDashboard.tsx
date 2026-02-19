import { useEffect, useState } from "react";
import { getBalance, getTransactions } from "../api";
import { Transaction } from "../types";
import { useChildUser } from "../hooks/useAuthOutletContext";
import Card from "../components/ui/Card";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import BalanceDisplay from "../components/BalanceDisplay";
import TransactionsCard from "../components/TransactionsCard";
import { TrendingUp } from "lucide-react";

export default function ChildDashboard() {
  const user = useChildUser();
  const [balance, setBalance] = useState<number>(0);
  const [interestRateBps, setInterestRateBps] = useState<number>(0);
  const [interestRateDisplay, setInterestRateDisplay] = useState<string>("");
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loadingData, setLoadingData] = useState(true);

  useEffect(() => {
    Promise.all([
      getBalance(user.user_id),
      getTransactions(user.user_id)
    ]).then(([balanceRes, txRes]) => {
      setBalance(balanceRes.balance_cents);
      setInterestRateBps(balanceRes.interest_rate_bps);
      setInterestRateDisplay(balanceRes.interest_rate_display);
      setTransactions(txRes.transactions || []);
    }).catch(() => {
      // Silently fail
    }).finally(() => {
      setLoadingData(false);
    });
  }, [user.user_id]);

  return (
    <div className="max-w-[480px] mx-auto space-y-6 animate-fade-in-up">
      {/* Welcome */}
      <div>
        <h2 className="text-2xl font-bold text-forest">
          Welcome, {user.first_name}!{user.avatar ? ` ${user.avatar}` : ''}
        </h2>
      </div>

      {/* Hero balance card */}
      <Card padding="lg" className="text-center">
        <p className="text-sm font-semibold text-bark-light uppercase tracking-wide mb-2">
          Your Balance
        </p>
        {loadingData ? (
          <LoadingSpinner variant="inline" />
        ) : (
          <BalanceDisplay balanceCents={balance} size="large" />
        )}
        {interestRateBps > 0 && !loadingData && (
          <div className="mt-4 flex justify-center items-center gap-1.5 bg-sage-light/30 text-forest text-sm font-medium px-3 py-1.5 rounded-full">
            <TrendingUp className="h-4 w-4" aria-hidden="true" />
            {interestRateDisplay} annual interest
          </div>
        )}
      </Card>

      {/* Transactions */}
      {!loadingData && (
        <TransactionsCard
          childId={user.user_id}
          balanceCents={balance}
          interestRateBps={interestRateBps}
          transactions={transactions}
        />
      )}
    </div>
  );
}
