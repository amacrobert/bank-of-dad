import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { get, post, updateBankName, ApiRequestError } from "../api";
import { setTokens, getRefreshToken } from "../auth";
import { ParentUser, Child, ChildListResponse } from "../types";
import SlugPicker from "../components/SlugPicker";
import BankNameInput from "../components/BankNameInput";
import AddChildForm from "../components/AddChildForm";
import ChildSelectorBar from "../components/ChildSelectorBar";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import Modal from "../components/ui/Modal";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import { Leaf, PartyPopper, Users, Sparkles } from "lucide-react";

export default function SetupPage() {
  const navigate = useNavigate();
  const [selectedSlug, setSelectedSlug] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  // Step tracking: 1 = slug picker, 2 = bank name, 3 = add children, 4 = confirmation
  const [step, setStep] = useState(1);
  const [bankName, setBankName] = useState("Dad");
  const [showAddChild, setShowAddChild] = useState(false);
  const [children, setChildren] = useState<Child[]>([]);
  const [childRefreshKey, setChildRefreshKey] = useState(0);

  const fetchChildren = useCallback(() => {
    get<ChildListResponse>("/children")
      .then((data) => setChildren(data.children || []))
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (step === 3) {
      fetchChildren();
    }
  }, [step, childRefreshKey, fetchChildren]);

  useEffect(() => {
    get<ParentUser>("/auth/me")
      .then((data) => {
        if (data.family_id > 0) {
          setStep(3);
        }
      })
      .catch(() => {
        // If the call fails, stay on step 1
      })
      .finally(() => setLoading(false));
  }, []);

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
      setStep(2);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Something went wrong. Please try again.");
      }
      setSubmitting(false);
    }
  };

  const handleBankNameSubmit = async () => {
    if (!bankName.trim()) return;
    setSubmitting(true);
    setError(null);
    try {
      await updateBankName(bankName.trim());
      setStep(3);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to save bank name. Please try again.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleChildAdded = () => {
    setChildRefreshKey((k) => k + 1);
    setShowAddChild(false);
  };

  const goToDashboard = () => {
    navigate("/dashboard", { replace: true });
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-cream flex items-center justify-center px-4 py-8">
        <LoadingSpinner />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-cream flex items-center justify-center px-4 py-8">
      <div className="w-full max-w-md animate-fade-in-up">
        {/* Progress dots */}
        <div className="flex justify-center gap-2 mb-8">
          {[1, 2, 3, 4].map((s) => (
            <div
              key={s}
              className={`w-3 h-3 rounded-full transition-colors ${
                s <= step ? "bg-forest" : "bg-sand"
              }`}
            />
          ))}
        </div>

        {/* Step 1: Choose family slug */}
        {step === 1 && (
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
        )}

        {/* Step 2: Bank name */}
        {step === 2 && (
          <Card padding="lg">
            <div className="flex justify-center mb-4">
              <div className="w-14 h-14 bg-forest/10 rounded-2xl flex items-center justify-center">
                <Sparkles className="h-7 w-7 text-forest" aria-hidden="true" />
              </div>
            </div>

            <h1 className="text-2xl font-bold text-forest text-center mb-2">
              Personalize your bank
            </h1>
            <p className="text-bark-light text-center mb-6">
              Who runs this bank? Pick a name or create your own.
            </p>

            <BankNameInput value={bankName} onChange={setBankName} />

            {error && (
              <div className="mt-4 bg-terracotta/10 border border-terracotta/20 rounded-xl p-3">
                <p className="text-sm text-terracotta font-medium">{error}</p>
              </div>
            )}

            <div className="mt-6">
              <Button
                onClick={handleBankNameSubmit}
                loading={submitting}
                disabled={!bankName.trim() || submitting}
                className="w-full"
              >
                {submitting ? "Saving..." : "Continue"}
              </Button>
            </div>
          </Card>
        )}

        {/* Step 3: Add children */}
        {step === 3 && (
          <div className="space-y-4">
            <Card padding="lg">
              <div className="flex justify-center mb-4">
                <div className="w-14 h-14 bg-forest/10 rounded-2xl flex items-center justify-center">
                  <Users className="h-7 w-7 text-forest" aria-hidden="true" />
                </div>
              </div>

              <h1 className="text-2xl font-bold text-forest text-center mb-2">
                Add your children
              </h1>
              <p className="text-bark-light text-center mb-2">
                Create accounts for your kids so they can log in and track their savings.
              </p>

              <ChildSelectorBar
                children={children}
                selectedChildId={null}
                onSelectChild={() => {}}
                onAddChild={() => setShowAddChild(true)}
                selectable={false}
              />
            </Card>

            <Modal open={showAddChild} onClose={() => setShowAddChild(false)}>
              <AddChildForm
                onChildAdded={handleChildAdded}
                onCancel={() => setShowAddChild(false)}
              />
            </Modal>

            <div className="space-y-3">
              <Button onClick={() => setStep(4)} className="w-full">
                Continue to Dashboard
              </Button>
              <button
                onClick={() => setStep(4)}
                className="w-full text-center text-sm text-bark-light hover:text-bark transition-colors cursor-pointer"
              >
                Skip for now
              </button>
            </div>
          </div>
        )}

        {/* Step 4: Confirmation */}
        {step === 4 && (
          <Card padding="lg">
            <div className="text-center">
              <div className="flex justify-center mb-4">
                <div className="w-16 h-16 bg-amber/15 rounded-full flex items-center justify-center">
                  <PartyPopper className="h-8 w-8 text-amber" aria-hidden="true" />
                </div>
              </div>
              <h2 className="text-2xl font-bold text-forest mb-2">You&apos;re all set!</h2>
              <p className="text-bark-light mb-6">
                Your family bank is ready. Head to the dashboard to get started!
              </p>
              <Button onClick={goToDashboard} className="w-full">
                Go to Dashboard
              </Button>
            </div>
          </Card>
        )}
      </div>
    </div>
  );
}
