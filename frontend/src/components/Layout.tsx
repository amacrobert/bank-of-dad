import { ReactNode } from "react";
import { useNavigate } from "react-router-dom";
import { post } from "../api";
import { getRefreshToken, clearTokens } from "../auth";
import { AuthUser } from "../types";
import { Leaf, LayoutDashboard, Home, LogOut } from "lucide-react";

interface LayoutProps {
  user: AuthUser;
  children: ReactNode;
  maxWidth?: "narrow" | "wide";
}

export default function Layout({ user, children, maxWidth = "narrow" }: LayoutProps) {
  const navigate = useNavigate();

  const displayName =
    user.user_type === "parent" ? user.display_name : user.first_name;

  const handleLogout = async () => {
    try {
      await post("/auth/logout", { refresh_token: getRefreshToken() });
    } catch {
      // Even if logout fails server-side, clear tokens and redirect
    }
    clearTokens();
    if (user.user_type === "child" && user.family_slug) {
      navigate(`/${user.family_slug}`);
    } else {
      navigate("/");
    }
  };

  const maxWidthClass = maxWidth === "wide" ? "max-w-[960px]" : "max-w-[480px]";

  return (
    <div className="min-h-screen bg-cream flex flex-col">
      {/* Desktop top nav */}
      <nav className="hidden md:flex items-center justify-between px-6 py-4 bg-white border-b border-sand" aria-label="Main navigation">
        <div className="flex items-center gap-2">
          <Leaf className="h-6 w-6 text-forest" aria-hidden="true" />
          <span className="text-lg font-bold text-forest">Bank of Dad</span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm font-medium text-bark-light">{displayName}</span>
          <button
            onClick={handleLogout}
            className="inline-flex items-center gap-1.5 text-sm font-medium text-bark-light hover:text-terracotta transition-colors cursor-pointer"
          >
            <LogOut className="h-4 w-4" aria-hidden="true" />
            Log out
          </button>
        </div>
      </nav>

      {/* Main content */}
      <main className={`flex-1 w-full ${maxWidthClass} mx-auto px-4 py-6 pb-24 md:pb-6`}>
        {children}
      </main>

      {/* Mobile bottom tab bar */}
      <nav className="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-sand px-4 py-2" aria-label="Mobile navigation">
        <div className="flex items-center justify-around">
          {user.user_type === "parent" ? (
            <button
              onClick={() => navigate("/dashboard")}
              className="flex flex-col items-center gap-1 py-2 px-4 text-forest cursor-pointer"
            >
              <LayoutDashboard className="h-6 w-6" aria-hidden="true" />
              <span className="text-xs font-semibold">Dashboard</span>
            </button>
          ) : (
            <button
              onClick={() => navigate("/child/dashboard")}
              className="flex flex-col items-center gap-1 py-2 px-4 text-forest cursor-pointer"
            >
              <Home className="h-6 w-6" aria-hidden="true" />
              <span className="text-xs font-semibold">Home</span>
            </button>
          )}
          <button
            onClick={handleLogout}
            className="flex flex-col items-center gap-1 py-2 px-4 text-bark-light hover:text-terracotta transition-colors cursor-pointer"
          >
            <LogOut className="h-6 w-6" aria-hidden="true" />
            <span className="text-xs font-semibold">Log out</span>
          </button>
        </div>
      </nav>
    </div>
  );
}
