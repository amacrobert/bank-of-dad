import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { get } from "../api";
import { AuthUser } from "../types";

export default function GoogleCallback() {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    get<AuthUser>("/auth/me")
      .then((user) => {
        if (user.user_type === "parent" && user.family_id === 0) {
          navigate("/setup", { replace: true });
        } else {
          navigate("/dashboard", { replace: true });
        }
      })
      .catch(() => {
        setError("Authentication failed. Please try again.");
      });
  }, [navigate]);

  if (error) {
    return (
      <div className="callback">
        <p>{error}</p>
        <a href="/">Back to home</a>
      </div>
    );
  }

  return (
    <div className="callback">
      <p>Signing you in...</p>
    </div>
  );
}
