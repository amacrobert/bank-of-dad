import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { get, post, ApiRequestError } from "../api";
import { setTokens } from "../auth";
import { FamilyCheck } from "../types";
import { Users, AlertCircle } from "lucide-react";
import Card from "../components/ui/Card";
import Input from "../components/ui/Input";
import Button from "../components/ui/Button";
import LoadingSpinner from "../components/ui/LoadingSpinner";

export default function FamilyLogin() {
  const { familySlug } = useParams<{ familySlug: string }>();
  const navigate = useNavigate();
  const [familyExists, setFamilyExists] = useState<boolean | null>(null);
  const [firstName, setFirstName] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!familySlug) return;
    get<FamilyCheck>(`/families/${familySlug}`)
      .then((data) => {
        setFamilyExists(data.exists);
      })
      .catch(() => {
        setFamilyExists(false);
      });
  }, [familySlug]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);

    try {
      const resp = await post<{ access_token: string; refresh_token: string }>("/auth/child/login", {
        family_slug: familySlug,
        first_name: firstName,
        password,
      });
      setTokens(resp.access_token, resp.refresh_token);
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
        <Card padding="lg">
          {/* Avatar circle */}
          <div className="flex justify-center mb-6">
            <div className="w-16 h-16 bg-sage-light/40 rounded-full flex items-center justify-center">
              <Users className="h-8 w-8 text-forest" aria-hidden="true" />
            </div>
          </div>

          <h1 className="text-2xl font-bold text-forest text-center mb-1">
            Welcome!
          </h1>
          <p className="text-bark-light text-center mb-6">
            Log in to see your account.
          </p>

          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              label="Your First Name"
              id="child-first-name"
              type="text"
              value={firstName}
              onChange={(e) => setFirstName(e.target.value)}
              required
              disabled={submitting}
              autoComplete="given-name"
            />

            <Input
              label="Password"
              id="child-login-password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              disabled={submitting}
              autoComplete="current-password"
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
        </Card>
      </div>
    </div>
  );
}
