import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { get, post } from "../api";
import { ParentUser, Child } from "../types";
import AddChildForm from "../components/AddChildForm";
import ChildList from "../components/ChildList";
import ManageChild from "../components/ManageChild";

export default function ParentDashboard() {
  const navigate = useNavigate();
  const [user, setUser] = useState<ParentUser | null>(null);
  const [loading, setLoading] = useState(true);
  const [childRefreshKey, setChildRefreshKey] = useState(0);
  const [selectedChild, setSelectedChild] = useState<Child | null>(null);

  useEffect(() => {
    get<ParentUser>("/auth/me")
      .then((data) => {
        if (data.user_type !== "parent") {
          navigate("/");
          return;
        }
        if (data.family_id === 0) {
          navigate("/setup", { replace: true });
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
    navigate("/");
  };

  const handleChildAdded = () => {
    setChildRefreshKey((k) => k + 1);
  };

  const handleChildUpdated = () => {
    setChildRefreshKey((k) => k + 1);
  };

  if (loading || !user) {
    return <div>Loading...</div>;
  }

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <h1>Bank of Dad</h1>
        <div className="user-info">
          <span>{user.display_name}</span>
          <button onClick={handleLogout}>Log out</button>
        </div>
      </header>

      <main>
        <section className="family-info">
          <h2>Your Family Bank</h2>
          <p>
            Family URL: <strong>/{user.family_slug}</strong>
          </p>
          <p>Share this link with your kids so they can log in.</p>
        </section>

        <section className="children-section">
          <AddChildForm onChildAdded={handleChildAdded} />
          <ChildList
            refreshKey={childRefreshKey}
            onSelectChild={setSelectedChild}
          />
          {selectedChild && (
            <ManageChild
              child={selectedChild}
              onUpdated={handleChildUpdated}
              onClose={() => setSelectedChild(null)}
            />
          )}
        </section>
      </main>
    </div>
  );
}
