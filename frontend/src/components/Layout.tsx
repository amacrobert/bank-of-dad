import { ReactNode } from "react";
import { useNavigate } from "react-router-dom";
import { post } from "../api";
import { AuthUser } from "../types";

interface LayoutProps {
  user: AuthUser;
  children: ReactNode;
}

export default function Layout({ user, children }: LayoutProps) {
  const navigate = useNavigate();

  const displayName =
    user.user_type === "parent" ? user.display_name : user.first_name;

  const handleLogout = async () => {
    try {
      await post("/auth/logout");
    } catch {
      // Even if logout fails server-side, redirect to home
    }
    navigate("/");
  };

  return (
    <div className="layout">
      <nav className="nav">
        <div className="nav-brand">Bank of Dad</div>
        <div className="nav-user">
          <span>{displayName}</span>
          <button onClick={handleLogout} className="btn-logout">
            Log out
          </button>
        </div>
      </nav>
      <main className="main">{children}</main>
    </div>
  );
}
