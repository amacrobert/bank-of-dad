import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { get, post, ApiRequestError } from "../api";
import { setTokens } from "../auth";
import { setFamilySlug, clearFamilySlug } from "../utils/familyPreference";
import GoogleSignInButton from "../components/GoogleSignInButton";
import { FamilyCheck, FamilyChild, FamilyChildrenResponse } from "../types";
import { AlertCircle, ArrowLeft } from "lucide-react";
import Input from "../components/ui/Input";
import Button from "../components/ui/Button";
import LoadingSpinner from "../components/ui/LoadingSpinner";

export default function FamilyLogin() {
  const { familySlug } = useParams<{ familySlug: string }>();
  const navigate = useNavigate();
  const [familyExists, setFamilyExists] = useState<boolean | null>(null);
  const [children, setChildren] = useState<FamilyChild[]>([]);
  const [selectedChild, setSelectedChild] = useState<FamilyChild | null>(null);
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!familySlug) return;

    Promise.all([
      get<FamilyCheck>(`/families/${familySlug}`),
      get<FamilyChildrenResponse>(`/families/${familySlug}/children`),
    ])
      .then(([familyData, childrenData]) => {
        setFamilyExists(familyData.exists);
        setChildren(childrenData.children || []);
      })
      .catch(() => {
        setFamilyExists(false);
      });
  }, [familySlug]);

  const handleBack = () => {
    setSelectedChild(null);
    setPassword("");
    setError(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedChild) return;
    setSubmitting(true);
    setError(null);

    try {
      const resp = await post<{ access_token: string; refresh_token: string }>("/auth/child/login", {
        family_slug: familySlug,
        first_name: selectedChild.first_name,
        password,
      });
      setTokens(resp.access_token, resp.refresh_token);
      if (familySlug) setFamilySlug(familySlug);
      navigate("/child/dashboard", { replace: true });
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Something went wrong. Please try again.");
      }
      setSubmitting(false);
    }
  };

  if (familyExists === null) {
    return (
      <div className="min-h-screen bg-cream flex items-center justify-center">
        <LoadingSpinner message="Loading..." />
      </div>
    );
  }

  if (!familyExists) {
    return (
      <div className="min-h-screen bg-cream flex flex-col items-center justify-center px-4">
        <div className="text-center animate-fade-in-up">
          <AlertCircle className="h-12 w-12 text-terracotta mx-auto mb-4" aria-hidden="true" />
          <h1 className="text-2xl font-bold text-forest mb-2">Family not found</h1>
          <p className="text-bark-light mb-6">There is no family bank at this URL.</p>
          <a href="/">
            <Button variant="secondary">Go to Bank of Dad</Button>
          </a>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-cream flex items-center justify-center px-4 py-8">
      <div className="w-full max-w-sm animate-fade-in-up">
        {selectedChild ? (
          /* State 2: Password entry */
          <>
            <button
              type="button"
              onClick={handleBack}
              className="inline-flex items-center gap-1 text-sm text-bark-light hover:text-bark transition-colors mb-4 cursor-pointer"
            >
              <ArrowLeft className="h-4 w-4" />
              Back
            </button>

            <div className="flex flex-col items-center mb-6">
              <div className="w-16 h-16 bg-sage-light/40 rounded-full flex items-center justify-center mb-3">
                <span className="text-3xl">
                  {selectedChild.avatar || selectedChild.first_name.charAt(0).toUpperCase()}
                </span>
              </div>
              <h1 className="text-2xl font-bold text-forest">
                {selectedChild.first_name}
              </h1>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              <Input
                label="Password"
                id="child-login-password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                disabled={submitting}
                autoComplete="current-password"
                autoFocus
              />

              {error && (
                <div className="bg-terracotta/10 border border-terracotta/20 rounded-xl p-3">
                  <p className="text-sm text-terracotta font-medium">{error}</p>
                </div>
              )}

              <Button
                type="submit"
                loading={submitting}
                className="w-full"
              >
                {submitting ? "Logging in..." : "Log In"}
              </Button>
            </form>
          </>
        ) : (
          /* State 1: Child picker grid */
          <>
            <h1 className="text-2xl font-bold text-forest text-center mb-6">
              Who's logging in?
            </h1>

            {children.length > 0 ? (
              <div className="flex flex-wrap justify-center gap-3 mb-4">
                {children.map((child) => (
                  <button
                    key={child.first_name}
                    type="button"
                    onClick={() => setSelectedChild(child)}
                    className="w-[calc(50%-6px)] aspect-square flex flex-col items-center justify-center gap-2 p-4 rounded-xl border border-sand bg-white hover:border-forest hover:bg-sage-light/20 transition-all duration-200 cursor-pointer"
                  >
                    <div className="w-14 h-14 bg-sage-light/40 rounded-full flex items-center justify-center">
                      <span className="text-2xl">
                        {child.avatar || child.first_name.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <span className="text-sm font-semibold text-bark">
                      {child.first_name}
                    </span>
                  </button>
                ))}
              </div>
            ) : (
              <p className="text-bark-light text-center mb-4">
                No accounts have been set up yet.
              </p>
            )}

            {/* Parent login */}
            <div className="mt-6 pt-5 border-t border-sand space-y-3">
              <p className="text-sm text-bark-light text-center">Parent login</p>
              <div className="flex justify-center">
                <GoogleSignInButton size="default" />
              </div>
              <p className="text-center">
                <button
                  type="button"
                  onClick={() => { clearFamilySlug(); navigate("/"); }}
                  className="text-xs text-bark-light hover:underline cursor-pointer"
                >
                  Not your bank?
                </button>
              </p>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
