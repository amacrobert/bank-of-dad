import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { get } from "../api";
import { ParentUser, Child } from "../types";
import Layout from "../components/Layout";
import Card from "../components/ui/Card";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import AddChildForm from "../components/AddChildForm";
import ChildList from "../components/ChildList";
import ManageChild from "../components/ManageChild";
import { Link as LinkIcon, Copy, Check, ChevronDown, UserPlus } from "lucide-react";

export default function ParentDashboard() {
  const navigate = useNavigate();
  const [user, setUser] = useState<ParentUser | null>(null);
  const [loading, setLoading] = useState(true);
  const [childRefreshKey, setChildRefreshKey] = useState(0);
  const [selectedChild, setSelectedChild] = useState<Child | null>(null);
  const [copied, setCopied] = useState(false);
  const [showAddChild, setShowAddChild] = useState(false);
  const addChildInitialized = useRef(false);

  const handleCopyFamilyUrl = () => {
    const fullUrl = `${window.location.origin}/${user?.family_slug}`;
    navigator.clipboard.writeText(fullUrl).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

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
            <LinkIcon className="h-4 w-4 shrink-0" aria-hidden="true" />
            <span>Family URL:</span>
            <button
              onClick={handleCopyFamilyUrl}
              className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-cream-dark/50 hover:bg-cream-dark rounded-md transition-colors font-mono text-bark text-xs"
              title="Click to copy"
            >
              <span>{window.location.origin}/{user.family_slug}</span>
              {copied ? (
                <Check className="h-3.5 w-3.5 text-forest" aria-label="Copied" />
              ) : (
                <Copy className="h-3.5 w-3.5 text-bark-light" aria-label="Copy URL" />
              )}
            </button>
            {copied && <span className="text-xs text-forest font-medium">Copied!</span>}
          </div>
          <p className="text-sm text-bark-light mt-1">Share this link with your kids so they can log in.</p>
        </div>

        {/* Two-column layout on desktop */}
        <div className="md:grid md:grid-cols-[340px_1fr] md:gap-6">
          {/* Left column: child list + add child */}
          <div className="space-y-4 mb-6 md:mb-0">
            <Card padding="md">
              <ChildList
                refreshKey={childRefreshKey}
                onSelectChild={setSelectedChild}
                selectedChildId={selectedChild?.id}
                onLoaded={(count) => {
                  if (!addChildInitialized.current) {
                    addChildInitialized.current = true;
                    if (count === 0) setShowAddChild(true);
                  }
                }}
              />
            </Card>

            <button
              onClick={() => setShowAddChild((v) => !v)}
              className="w-full flex items-center justify-between p-3 rounded-xl bg-cream hover:bg-cream-dark transition-colors cursor-pointer"
            >
              <div className="flex items-center gap-2">
                <UserPlus className="h-5 w-5 text-forest" aria-hidden="true" />
                <span className="text-base font-bold text-bark">Add a Child</span>
              </div>
              <ChevronDown
                className="h-5 w-5 text-bark-light transition-transform"
                style={{ transform: showAddChild ? "rotate(180deg)" : undefined }}
                aria-hidden="true"
              />
            </button>

            {showAddChild && <AddChildForm onChildAdded={handleChildAdded} />}
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
