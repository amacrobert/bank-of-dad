import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { post } from "../api";
import { ApiRequestError } from "../api";
import { Family } from "../types";
import SlugPicker from "../components/SlugPicker";

export default function SetupPage() {
  const navigate = useNavigate();
  const [selectedSlug, setSelectedSlug] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async () => {
    if (!selectedSlug) return;

    setSubmitting(true);
    setError(null);

    try {
      await post<Family>("/families", { slug: selectedSlug });
      navigate("/dashboard", { replace: true });
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Something went wrong. Please try again.");
      }
      setSubmitting(false);
    }
  };

  return (
    <div className="setup-page">
      <h1>Set up your family bank</h1>
      <p>Choose a unique URL for your family. Your kids will use this to log in.</p>

      <SlugPicker onSelect={setSelectedSlug} disabled={submitting} />

      {error && <p className="error">{error}</p>}

      {selectedSlug && (
        <button
          className="btn-primary"
          onClick={handleSubmit}
          disabled={submitting}
        >
          {submitting ? "Creating..." : "Create Family Bank"}
        </button>
      )}
    </div>
  );
}
