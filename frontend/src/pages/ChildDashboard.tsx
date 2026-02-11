import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { get, getBalance, getTransactions } from "../api";
import { ChildUser, Transaction } from "../types";
import Layout from "../components/Layout";
import Card from "../components/ui/Card";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import BalanceDisplay from "../components/BalanceDisplay";
import TransactionHistory from "../components/TransactionHistory";
import UpcomingPayments from "../components/UpcomingPayments";
import { TrendingUp } from "lucide-react";

export default function ChildDashboard() {
  const navigate = useNavigate();
  const [user, setUser] = useState<ChildUser | null>(null);
  const [loading, setLoading] = useState(true);
  const [balance, setBalance] = useState<number>(0);
  const [interestRateBps, setInterestRateBps] = useState<number>(0);
  const [interestRateDisplay, setInterestRateDisplay] = useState<string>("");
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loadingData, setLoadingData] = useState(false);

  useEffect(() => {
    get<ChildUser>("/auth/me")
      .then((data) => {
        if (data.user_type !== "child") {
          navigate("/");
          return;
        }
        setUser(data);
        setLoading(false);

        setLoadingData(true);
        Promise.all([
          getBalance(data.user_id),
          getTransactions(data.user_id)
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
      })
      .catch(() => {
        navigate("/");
      });
  }, [navigate]);

  if (loading || !user) {
    return (
      <div className="min-h-screen bg-cream flex items-center justify-center">
        <LoadingSpinner message="Loading..." />
      </div>
    );
  }

  return (
    <Layout user={user} maxWidth="narrow">
      <div className="space-y-6 animate-fade-in-up">
        {/* Welcome */}
        <div>
          <h2 className="text-2xl font-bold text-forest">
            Welcome, {user.first_name}!
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

        {/* Upcoming payments */}
        {!loadingData && (
          <UpcomingPayments
            childId={user.user_id}
            balanceCents={balance}
            interestRateBps={interestRateBps}
          />
        )}

        {/* Transaction history */}
        <Card padding="md">
          <h3 className="text-base font-bold text-bark mb-3">Recent Activity</h3>
          {loadingData ? (
            <LoadingSpinner variant="inline" />
          ) : (
            <TransactionHistory transactions={transactions} />
          )}
        </Card>
      </div>
    </Layout>
  );
}
