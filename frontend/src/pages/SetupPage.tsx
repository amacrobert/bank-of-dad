import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { post } from "../api";
import { ApiRequestError } from "../api";
import { setTokens, getRefreshToken } from "../auth";
import SlugPicker from "../components/SlugPicker";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import { Leaf, PartyPopper } from "lucide-react";

export default function SetupPage() {
  const navigate = useNavigate();
  const [selectedSlug, setSelectedSlug] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [created, setCreated] = useState(false);

  const handleSubmit = async () => {
    if (!selectedSlug) return;

    setSubmitting(true);
    setError(null);

    try {
      const resp = await post<{ id: number; slug: string; access_token: string }>("/families", { slug: selectedSlug });
      if (resp.access_token) {
        const refreshToken = getRefreshToken();
        if (refreshToken) {
          setTokens(resp.access_token, refreshToken);
        }
      }
      setCreated(true);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Something went wrong. Please try again.");
      }
      setSubmitting(false);
    }
  };

  const step = created ? 2 : 1;

  return (
    <div className="min-h-screen bg-cream flex items-center justify-center px-4 py-8">
      <div className="w-full max-w-md animate-fade-in-up">
        {/* Progress dots */}
        <div className="flex justify-center gap-2 mb-8">
          {[1, 2].map((s) => (
            <div
              key={s}
              className={`w-3 h-3 rounded-full transition-colors ${
                s <= step ? "bg-forest" : "bg-sand"
              }`}
            />
          ))}
        </div>

        {!created ? (
          <Card padding="lg">
            <div className="flex justify-center mb-4">
              <div className="w-14 h-14 bg-forest/10 rounded-2xl flex items-center justify-center">
                <Leaf className="h-7 w-7 text-forest" aria-hidden="true" />
              </div>
            </div>

            <h1 className="text-2xl font-bold text-forest text-center mb-2">
              Set up your family bank
            </h1>
            <p className="text-bark-light text-center mb-6">
              Choose a unique URL for your family. Your kids will use this to log in.
            </p>

            <SlugPicker onSelect={setSelectedSlug} disabled={submitting} />

            {error && (
              <div className="mt-4 bg-terracotta/10 border border-terracotta/20 rounded-xl p-3">
                <p className="text-sm text-terracotta font-medium">{error}</p>
              </div>
            )}

            {selectedSlug && (
              <div className="mt-6">
                <Button
                  onClick={handleSubmit}
                  loading={submitting}
                  className="w-full"
                >
                  {submitting ? "Creating..." : "Create Family Bank"}
                </Button>
              </div>
            )}
          </Card>
        ) : (
          <Card padding="lg">
            <div className="text-center">
              <div className="flex justify-center mb-4">
                <div className="w-16 h-16 bg-amber/15 rounded-full flex items-center justify-center">
                  <PartyPopper className="h-8 w-8 text-amber" aria-hidden="true" />
                </div>
              </div>
              <h2 className="text-2xl font-bold text-forest mb-2">You're all set!</h2>
              <p className="text-bark-light mb-6">
                Your family bank is ready. Start adding your kids!
              </p>
              <Button onClick={() => navigate("/dashboard", { replace: true })} className="w-full">
                Go to Dashboard
              </Button>
            </div>
          </Card>
        )}
      </div>
    </div>
  );
}
