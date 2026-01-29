import { useEffect, useState, ReactNode } from "react";
import { useNavigate } from "react-router-dom";
import { get } from "../api";
import { AuthUser } from "../types";

interface ProtectedRouteProps {
  children: ReactNode;
  requiredRole?: "parent" | "child";
}

export default function ProtectedRoute({
  children,
  requiredRole,
}: ProtectedRouteProps) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    get<AuthUser>("/auth/me")
      .then((data) => {
        if (requiredRole && data.user_type !== requiredRole) {
          navigate("/");
          return;
        }
        setUser(data);
        setLoading(false);
      })
      .catch(() => {
        navigate("/");
      });
  }, [navigate, requiredRole]);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (!user) {
    return null;
  }

  return <>{children}</>;
}
