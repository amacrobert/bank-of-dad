import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { get, post } from "../api";
import { ChildUser } from "../types";

export default function ChildDashboard() {
  const navigate = useNavigate();
  const [user, setUser] = useState<ChildUser | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    get<ChildUser>("/auth/me")
      .then((data) => {
        if (data.user_type !== "child") {
          navigate("/");
          return;
        }
        setUser(data);
        setLoading(false);
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
        <p>Your account balance and transactions will appear here.</p>
      </main>
    </div>
  );
}
