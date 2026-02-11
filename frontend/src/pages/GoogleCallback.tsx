import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { get } from "../api";
import { AuthUser } from "../types";
import { Leaf, AlertCircle } from "lucide-react";
import Button from "../components/ui/Button";

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
      <div className="min-h-screen bg-cream flex flex-col items-center justify-center px-4">
        <div className="text-center animate-fade-in-up">
          <AlertCircle className="h-12 w-12 text-terracotta mx-auto mb-4" aria-hidden="true" />
          <p className="text-bark text-lg font-medium mb-6">{error}</p>
          <a href="/">
            <Button variant="secondary">Back to Home</Button>
          </a>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-cream flex flex-col items-center justify-center px-4">
      <div className="text-center animate-fade-in-up">
        <Leaf className="h-12 w-12 text-forest mx-auto mb-4 animate-pulse" aria-hidden="true" />
        <p className="text-bark-light text-lg font-medium">Signing you in...</p>
      </div>
    </div>
  );
}
