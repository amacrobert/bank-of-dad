import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { get, post, getBalance, getTransactions } from "../api";
import { ChildUser, Transaction } from "../types";
import BalanceDisplay from "../components/BalanceDisplay";
import TransactionHistory from "../components/TransactionHistory";
import UpcomingAllowances from "../components/UpcomingAllowances";

export default function ChildDashboard() {
  const navigate = useNavigate();
  const [user, setUser] = useState<ChildUser | null>(null);
  const [loading, setLoading] = useState(true);
  const [balance, setBalance] = useState<number>(0);
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

        // Fetch balance and transactions
        setLoadingData(true);
        Promise.all([
          getBalance(data.user_id),
          getTransactions(data.user_id)
        ]).then(([balanceRes, txRes]) => {
          setBalance(balanceRes.balance_cents);
          setTransactions(txRes.transactions || []);
        }).catch(() => {
          // Silently fail for now - user can see their account
        }).finally(() => {
          setLoadingData(false);
        });
      })
      .catch(() => {
        navigate("/");
      });
  }, [navigate]);

  const handleLogout = async () => {
    try {
      await post("/auth/logout");
    } catch {
      // proceed regardless
    }
    if (user?.family_slug) {
      navigate(`/${user.family_slug}`);
    } else {
      navigate("/");
    }
  };

  if (loading || !user) {
    return <div>Loading...</div>;
  }

  return (
    <div className="child-dashboard">
      <header className="dashboard-header">
        <h1>Bank of Dad</h1>
        <div className="user-info">
          <span>{user.first_name}</span>
          <button onClick={handleLogout}>Log out</button>
        </div>
      </header>

      <main>
        <h2>Welcome, {user.first_name}!</h2>

        <section className="balance-section">
          <h3>Your Balance</h3>
          {loadingData ? (
            <p>Loading...</p>
          ) : (
            <BalanceDisplay balanceCents={balance} size="large" />
          )}
        </section>

        <section className="upcoming-section">
          <UpcomingAllowances childId={user.user_id} />
        </section>

        <section className="transactions-section">
          <h3>Transaction History</h3>
          {loadingData ? (
            <p>Loading...</p>
          ) : (
            <TransactionHistory transactions={transactions} />
          )}
        </section>
      </main>
    </div>
  );
}
