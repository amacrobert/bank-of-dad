import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { get, post, ApiRequestError } from "../api";
import { FamilyCheck } from "../types";

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
      await post("/auth/child/login", {
        family_slug: familySlug,
        first_name: firstName,
        password,
      });
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
    return <div>Loading...</div>;
  }

  if (!familyExists) {
    return (
      <div className="family-login">
        <h1>Family not found</h1>
        <p>There is no family bank at this URL.</p>
        <a href="/">Go to Bank of Dad</a>
      </div>
    );
  }

  return (
    <div className="family-login">
      <h1>Welcome to your Family Bank!</h1>
      <p>Log in to see your account.</p>
      <form onSubmit={handleSubmit}>
        <div className="form-field">
          <label htmlFor="child-first-name">Your First Name</label>
          <input
            id="child-first-name"
            type="text"
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
            required
            disabled={submitting}
          />
        </div>
        <div className="form-field">
          <label htmlFor="child-login-password">Password</label>
          <input
            id="child-login-password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            disabled={submitting}
          />
        </div>
        {error && <p className="error">{error}</p>}
        <button type="submit" disabled={submitting}>
          {submitting ? "Logging in..." : "Log In"}
        </button>
      </form>
    </div>
  );
}
