import { useState, useEffect, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useParentUser } from "../hooks/useAuthOutletContext";
import { get } from "../api";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import ChildSelectorBar from "../components/ChildSelectorBar";
import ManageChild from "../components/ManageChild";
import { Child, ChildListResponse } from "../types";
import { Link as LinkIcon, Copy, Check, Users } from "lucide-react";

export default function ParentDashboard() {
  const user = useParentUser();
  const navigate = useNavigate();
  const { childName } = useParams<{ childName?: string }>();
  const [childRefreshKey, setChildRefreshKey] = useState(0);
  const [children, setChildren] = useState<Child[]>([]);
  const [copied, setCopied] = useState(false);
  const [loading, setLoading] = useState(true);

  // Derive selected child from URL param
  const selectedChild = useMemo(() => {
    if (!childName || children.length === 0) return null;
    return children.find(
      (c) => c.first_name.toLowerCase() === childName.toLowerCase()
    ) ?? null;
  }, [childName, children]);

  // Redirect if child name in URL is invalid
  useEffect(() => {
    if (childName && children.length > 0 && !selectedChild) {
      navigate("/dashboard", { replace: true });
    }
  }, [childName, children, selectedChild, navigate]);

  useEffect(() => {
    setLoading(true);
    get<ChildListResponse>("/children")
      .then((data) => {
        const list = data.children || [];
        setChildren(list);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [childRefreshKey]);

  const handleCopyFamilyUrl = () => {
    const fullUrl = `${window.location.origin}/${user.family_slug}`;
    navigator.clipboard.writeText(fullUrl).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  const handleChildUpdated = () => {
    setChildRefreshKey((k) => k + 1);
  };

  return (
    <div className="max-w-[960px] mx-auto animate-fade-in-up">
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

      {/* Empty state when no children */}
      {!loading && children.length === 0 && (
        <Card padding="lg">
          <div className="text-center py-8">
            <div className="flex justify-center mb-4">
              <div className="w-14 h-14 bg-forest/10 rounded-2xl flex items-center justify-center">
                <Users className="h-7 w-7 text-forest" aria-hidden="true" />
              </div>
            </div>
            <h3 className="text-lg font-bold text-bark mb-2">No children yet</h3>
            <p className="text-bark-light mb-4">
              Add your first child to start managing their finances.
            </p>
            <Button onClick={() => navigate("/settings/children")}>
              Go to Settings &rarr; Children
            </Button>
          </div>
        </Card>
      )}

      {/* Child selector + full-width content */}
      {(loading || children.length > 0) && (
        <div className="space-y-4">
          <ChildSelectorBar
            children={children}
            selectedChildId={selectedChild?.id ?? null}
            onSelectChild={(child) => {
              if (child) {
                navigate(`/dashboard/${child.first_name.toLowerCase()}`);
              } else {
                navigate("/dashboard");
              }
            }}
            loading={loading}
          />

          {selectedChild ? (
            <ManageChild
              key={selectedChild.id}
              child={selectedChild}
              onUpdated={handleChildUpdated}
            />
          ) : (
            !loading && (
              <Card padding="lg">
                <p className="text-bark-light text-center py-4">
                  Select a child to manage their finances.
                </p>
              </Card>
            )
          )}
        </div>
      )}
    </div>
  );
}
