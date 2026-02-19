import { useEffect, useState } from "react";
import { useNavigate, Outlet } from "react-router-dom";
import { get } from "../api";
import { AuthUser, ParentUser } from "../types";
import Layout from "./Layout";
import LoadingSpinner from "./ui/LoadingSpinner";

interface Props {
  userType: "parent" | "child";
}

export default function AuthenticatedLayout({ userType }: Props) {
  const navigate = useNavigate();
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    get<AuthUser>("/auth/me")
      .then((data) => {
        if (data.user_type !== userType) {
          navigate("/");
          return;
        }
        if (userType === "parent" && (data as ParentUser).family_id === 0) {
          navigate("/setup", { replace: true });
          return;
        }
        setUser(data);
        setLoading(false);
      })
      .catch(() => {
        navigate("/");
      });
  }, [navigate, userType]);

  if (loading || !user) {
    return (
      <div className="min-h-screen bg-cream flex items-center justify-center">
        <LoadingSpinner message="Loading..." />
      </div>
    );
  }

  return (
    <Layout user={user}>
      <Outlet context={{ user }} />
    </Layout>
  );
}
