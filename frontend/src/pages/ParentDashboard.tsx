import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { get } from "../api";
import { ParentUser, Child } from "../types";
import Layout from "../components/Layout";
import Card from "../components/ui/Card";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import AddChildForm from "../components/AddChildForm";
import ChildList from "../components/ChildList";
import ManageChild from "../components/ManageChild";
import { Link as LinkIcon } from "lucide-react";

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

  const handleChildAdded = () => {
    setChildRefreshKey((k) => k + 1);
  };

  const handleChildUpdated = () => {
    setChildRefreshKey((k) => k + 1);
  };

  if (loading || !user) {
    return (
      <div className="min-h-screen bg-cream flex items-center justify-center">
        <LoadingSpinner message="Loading..." />
      </div>
    );
  }

  return (
    <Layout user={user} maxWidth="wide">
      <div className="animate-fade-in-up">
        {/* Family info header */}
        <div className="mb-6">
          <h2 className="text-2xl font-bold text-forest mb-2">Your Family Bank</h2>
          <div className="flex items-center gap-2 text-sm text-bark-light">
            <LinkIcon className="h-4 w-4" aria-hidden="true" />
            <span>Family URL: <strong className="text-bark">/{user.family_slug}</strong></span>
          </div>
          <p className="text-sm text-bark-light mt-1">Share this link with your kids so they can log in.</p>
        </div>

        {/* Two-column layout on desktop */}
        <div className="md:grid md:grid-cols-[340px_1fr] md:gap-6">
          {/* Left column: child list + add child */}
          <div className="space-y-4 mb-6 md:mb-0">
            <AddChildForm onChildAdded={handleChildAdded} />

            <Card padding="md">
              <ChildList
                refreshKey={childRefreshKey}
                onSelectChild={setSelectedChild}
                selectedChildId={selectedChild?.id}
              />
            </Card>
          </div>

          {/* Right column: manage child or empty state */}
          <div>
            {selectedChild ? (
              <ManageChild
                key={selectedChild.id}
                child={selectedChild}
                onUpdated={handleChildUpdated}
                onClose={() => setSelectedChild(null)}
              />
            ) : (
              <Card padding="lg" className="hidden md:flex items-center justify-center min-h-[300px]">
                <p className="text-bark-light text-center">
                  Select a child to manage their account.
                </p>
              </Card>
            )}
          </div>
        </div>
      </div>
    </Layout>
  );
}
