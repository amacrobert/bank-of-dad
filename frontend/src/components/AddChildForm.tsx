import { useState } from "react";
import { post, ApiRequestError } from "../api";
import { ChildCreateResponse } from "../types";

interface AddChildFormProps {
  onChildAdded: () => void;
}

export default function AddChildForm({ onChildAdded }: AddChildFormProps) {
  const [firstName, setFirstName] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [created, setCreated] = useState<ChildCreateResponse | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    setCreated(null);

    try {
      const result = await post<ChildCreateResponse>("/children", {
        first_name: firstName,
        password,
      });
      setCreated(result);
      setFirstName("");
      setPassword("");
      onChildAdded();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Something went wrong. Please try again.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="add-child-form">
      <h3>Add a Child</h3>
      <form onSubmit={handleSubmit}>
        <div className="form-field">
          <label htmlFor="child-name">First Name</label>
          <input
            id="child-name"
            type="text"
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
            required
            disabled={submitting}
          />
        </div>
        <div className="form-field">
          <label htmlFor="child-password">Password (min 6 characters)</label>
          <input
            id="child-password"
            type="text"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            minLength={6}
            required
            disabled={submitting}
          />
        </div>
        {error && <p className="error">{error}</p>}
        <button type="submit" disabled={submitting}>
          {submitting ? "Creating..." : "Add Child"}
        </button>
      </form>

      {created && (
        <div className="child-created-info">
          <h4>Account created for {created.first_name}!</h4>
          <p>Login URL: <strong>{created.login_url}</strong></p>
          <p>Name: <strong>{created.first_name}</strong></p>
          <p>Password: <strong>{password || "(the password you just set)"}</strong></p>
          <p className="note">Share these credentials with your child.</p>
        </div>
      )}
    </div>
  );
}
