import { ReactNode } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { post } from "../api";
import { getRefreshToken, clearTokens } from "../auth";
import { AuthUser } from "../types";
import { Leaf, LayoutDashboard, Home, TrendingUp, LogOut, Settings } from "lucide-react";

interface LayoutProps {
  user: AuthUser;
  children: ReactNode;
  maxWidth?: "narrow" | "wide";
}

export default function Layout({ user, children, maxWidth = "narrow" }: LayoutProps) {
  const navigate = useNavigate();
  const location = useLocation();

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
    <div className="min-h-screen bg-cream flex flex-col lg:flex-row">
      {/* Desktop sidebar */}
      <nav className="hidden lg:flex flex-col w-56 h-screen sticky top-0 bg-white border-r border-sand" aria-label="Main navigation">
        {/* Branding */}
        <div className="flex items-center gap-2 px-5 py-5">
          <Leaf className="h-6 w-6 text-forest" aria-hidden="true" />
          <span className="text-lg font-bold text-forest">Bank of Dad</span>
        </div>

        {/* Navigation links */}
        <div className="flex-1 flex flex-col gap-1 px-3 py-2">
          {user.user_type === "parent" ? (
            <>
              <button
                onClick={() => navigate("/dashboard")}
                className={`
                  flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold
                  transition-colors text-left cursor-pointer
                  ${location.pathname === "/dashboard"
                    ? "bg-forest text-white"
                    : "text-bark-light hover:bg-cream-dark"
                  }
                `}
              >
                <LayoutDashboard className="h-4 w-4" aria-hidden="true" />
                Dashboard
              </button>
              <button
                onClick={() => navigate("/settings")}
                className={`
                  flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold
                  transition-colors text-left cursor-pointer
                  ${location.pathname === "/settings"
                    ? "bg-forest text-white"
                    : "text-bark-light hover:bg-cream-dark"
                  }
                `}
              >
                <Settings className="h-4 w-4" aria-hidden="true" />
                Settings
              </button>
            </>
          ) : (
            <>
              <button
                onClick={() => navigate("/child/dashboard")}
                className={`
                  flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold
                  transition-colors text-left cursor-pointer
                  ${location.pathname === "/child/dashboard"
                    ? "bg-forest text-white"
                    : "text-bark-light hover:bg-cream-dark"
                  }
                `}
              >
                <Home className="h-4 w-4" aria-hidden="true" />
                Home
              </button>
              <button
                onClick={() => navigate("/child/growth")}
                className={`
                  flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold
                  transition-colors text-left cursor-pointer
                  ${location.pathname === "/child/growth"
                    ? "bg-forest text-white"
                    : "text-bark-light hover:bg-cream-dark"
                  }
                `}
              >
                <TrendingUp className="h-4 w-4" aria-hidden="true" />
                Growth
              </button>
            </>
          )}
        </div>

        {/* User section */}
        <div className="border-t border-sand px-3 py-4 flex flex-col gap-2">
          <span className="px-4 text-sm font-medium text-bark-light truncate">{displayName}</span>
          <button
            onClick={handleLogout}
            className="flex items-center gap-3 px-4 py-2.5 rounded-xl text-sm font-semibold text-bark-light hover:text-terracotta hover:bg-cream-dark transition-colors cursor-pointer text-left"
          >
            <LogOut className="h-4 w-4" aria-hidden="true" />
            Log out
          </button>
        </div>
      </nav>

      {/* Main content */}
      <main className="flex-1 min-w-0">
        <div className={`w-full ${maxWidthClass} mx-auto px-4 py-6 pb-24 lg:pb-6`}>
          {children}
        </div>
      </main>

      {/* Mobile bottom tab bar */}
      <nav className="lg:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-sand px-4 py-2" aria-label="Mobile navigation">
        <div className="flex items-center justify-around">
          {user.user_type === "parent" ? (
            <>
              <button
                onClick={() => navigate("/dashboard")}
                className={`flex flex-col items-center gap-1 py-2 px-4 cursor-pointer ${location.pathname === "/dashboard" ? "text-forest" : "text-bark-light hover:text-forest transition-colors"}`}
              >
                <LayoutDashboard className="h-6 w-6" aria-hidden="true" />
                <span className="text-xs font-semibold">Dashboard</span>
              </button>
              <button
                onClick={() => navigate("/settings")}
                className={`flex flex-col items-center gap-1 py-2 px-4 cursor-pointer ${location.pathname === "/settings" ? "text-forest" : "text-bark-light hover:text-forest transition-colors"}`}
              >
                <Settings className="h-6 w-6" aria-hidden="true" />
                <span className="text-xs font-semibold">Settings</span>
              </button>
            </>
          ) : (
            <>
              <button
                onClick={() => navigate("/child/dashboard")}
                className={`flex flex-col items-center gap-1 py-2 px-4 cursor-pointer ${location.pathname === "/child/dashboard" ? "text-forest" : "text-bark-light hover:text-forest transition-colors"}`}
              >
                <Home className="h-6 w-6" aria-hidden="true" />
                <span className="text-xs font-semibold">Home</span>
              </button>
              <button
                onClick={() => navigate("/child/growth")}
                className={`flex flex-col items-center gap-1 py-2 px-4 cursor-pointer ${location.pathname === "/child/growth" ? "text-forest" : "text-bark-light hover:text-forest transition-colors"}`}
              >
                <TrendingUp className="h-6 w-6" aria-hidden="true" />
                <span className="text-xs font-semibold">Growth</span>
              </button>
            </>
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
