import { ReactNode, useState, useRef, useEffect, ComponentType } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { post } from "../api";
import { getRefreshToken, clearTokens } from "../auth";
import { AuthUser } from "../types";
import { useTheme } from "../context/ThemeContext";
import { Leaf, LayoutDashboard, Home, Target, TrendingUp, LogOut, Settings, ClipboardList, MoreHorizontal } from "lucide-react";
import Footer from "./Footer";
import ContactFormModal from "./ContactFormModal";

interface NavItem {
  label: string;
  icon: ComponentType<{ className?: string }>;
  path: string;
  pathMatch: string;
  action?: () => void;
  variant?: "danger";
}

const MAX_VISIBLE_MOBILE_ITEMS = 4;

interface LayoutProps {
  user: AuthUser;
  children: ReactNode;
}

export default function Layout({ user, children }: LayoutProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { setTheme } = useTheme();
  const [showContact, setShowContact] = useState(false);

  const displayName =
    user.user_type === "parent" ? user.display_name : user.first_name;

  const handleLogout = async () => {
    try {
      await post("/auth/logout", { refresh_token: getRefreshToken() });
    } catch {
      // Even if logout fails server-side, clear tokens and redirect
    }
    clearTokens();
    setTheme("sapling");
    if (user.user_type === "child" && user.family_slug) {
      navigate(`/${user.family_slug}`);
    } else {
      navigate("/");
    }
  };

  const navItems: NavItem[] = user.user_type === "parent"
    ? [
        { label: "Dashboard", icon: LayoutDashboard, path: "/dashboard", pathMatch: "/dashboard" },
        { label: "Chores", icon: ClipboardList, path: "/chores", pathMatch: "/chores" },
        { label: "Growth", icon: TrendingUp, path: "/growth", pathMatch: "/growth" },
        { label: "Settings", icon: Settings, path: "/settings", pathMatch: "/settings" },
        { label: "Log out", icon: LogOut, path: "", pathMatch: "", action: handleLogout, variant: "danger" },
      ]
    : [
        { label: "Home", icon: Home, path: "/child/dashboard", pathMatch: "/child/dashboard" },
        { label: "Goals", icon: Target, path: "/child/goals", pathMatch: "/child/goals" },
        { label: "Chores", icon: ClipboardList, path: "/child/chores", pathMatch: "/child/chores" },
        { label: "Growth", icon: TrendingUp, path: "/child/growth", pathMatch: "/child/growth" },
        { label: "Settings", icon: Settings, path: "/child/settings", pathMatch: "/child/settings" },
        { label: "Log out", icon: LogOut, path: "", pathMatch: "", action: handleLogout, variant: "danger" },
      ];

  const visibleItems = navItems.slice(0, MAX_VISIBLE_MOBILE_ITEMS);
  const overflowItems = navItems.slice(MAX_VISIBLE_MOBILE_ITEMS);
  const hasOverflow = overflowItems.length > 0;
  const isMoreActive = overflowItems.some(
    item => item.pathMatch && location.pathname.startsWith(item.pathMatch)
  );

  const [moreMenuOpen, setMoreMenuOpen] = useState(false);
  const moreMenuRef = useRef<HTMLDivElement>(null);

  // Close More menu on outside click
  useEffect(() => {
    if (!moreMenuOpen) return;
    const handleClickOutside = (e: MouseEvent) => {
      if (moreMenuRef.current && !moreMenuRef.current.contains(e.target as Node)) {
        setMoreMenuOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [moreMenuOpen]);

  // Close More menu on navigation
  useEffect(() => {
    setMoreMenuOpen(false);
  }, [location.pathname]);

  // Close More menu on Escape
  useEffect(() => {
    if (!moreMenuOpen) return;
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape") setMoreMenuOpen(false);
    };
    document.addEventListener("keydown", handleEscape);
    return () => document.removeEventListener("keydown", handleEscape);
  }, [moreMenuOpen]);

  return (
    <div className={`min-h-screen flex flex-col lg:flex-row ${user.user_type === "parent" ? "bg-cream" : ""}`}>
      {/* Desktop sidebar */}
      <nav className="hidden lg:flex flex-col w-56 h-screen sticky top-0 bg-white border-r border-sand" aria-label="Main navigation">
        {/* Branding */}
        <div className="flex items-center gap-2 px-5 py-5">
          <Leaf className="h-6 w-6 text-forest" aria-hidden="true" />
          <span className="text-lg font-bold text-forest truncate">Bank of {user.bank_name || "Dad"}</span>
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
                  ${location.pathname.startsWith("/dashboard")
                    ? "bg-forest text-white"
                    : "text-bark-light hover:bg-cream-dark"
                  }
                `}
              >
                <LayoutDashboard className="h-4 w-4" aria-hidden="true" />
                Dashboard
              </button>
              <button
                onClick={() => navigate("/chores")}
                className={`
                  flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold
                  transition-colors text-left cursor-pointer
                  ${location.pathname.startsWith("/chores")
                    ? "bg-forest text-white"
                    : "text-bark-light hover:bg-cream-dark"
                  }
                `}
              >
                <ClipboardList className="h-4 w-4" aria-hidden="true" />
                Chores
              </button>
              <button
                onClick={() => navigate("/growth")}
                className={`
                  flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold
                  transition-colors text-left cursor-pointer
                  ${location.pathname.startsWith("/growth")
                    ? "bg-forest text-white"
                    : "text-bark-light hover:bg-cream-dark"
                  }
                `}
              >
                <TrendingUp className="h-4 w-4" aria-hidden="true" />
                Growth
              </button>
              <button
                onClick={() => navigate("/settings")}
                className={`
                  flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold
                  transition-colors text-left cursor-pointer
                  ${location.pathname.startsWith("/settings")
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
                  ${location.pathname.startsWith("/child/dashboard")
                    ? "bg-forest text-white"
                    : "text-bark-light hover:bg-cream-dark"
                  }
                `}
              >
                <Home className="h-4 w-4" aria-hidden="true" />
                Home
              </button>
              <button
                onClick={() => navigate("/child/goals")}
                className={`
                  flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold
                  transition-colors text-left cursor-pointer
                  ${location.pathname.startsWith("/child/goals")
                    ? "bg-forest text-white"
                    : "text-bark-light hover:bg-cream-dark"
                  }
                `}
              >
                <Target className="h-4 w-4" aria-hidden="true" />
                Goals
              </button>
              <button
                onClick={() => navigate("/child/chores")}
                className={`
                  flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold
                  transition-colors text-left cursor-pointer
                  ${location.pathname.startsWith("/child/chores")
                    ? "bg-forest text-white"
                    : "text-bark-light hover:bg-cream-dark"
                  }
                `}
              >
                <ClipboardList className="h-4 w-4" aria-hidden="true" />
                Chores
              </button>
              <button
                onClick={() => navigate("/child/growth")}
                className={`
                  flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold
                  transition-colors text-left cursor-pointer
                  ${location.pathname.startsWith("/child/growth")
                    ? "bg-forest text-white"
                    : "text-bark-light hover:bg-cream-dark"
                  }
                `}
              >
                <TrendingUp className="h-4 w-4" aria-hidden="true" />
                Growth
              </button>
              <button
                onClick={() => navigate("/child/settings")}
                className={`
                  flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold
                  transition-colors text-left cursor-pointer
                  ${location.pathname.startsWith("/child/settings")
                    ? "bg-forest text-white"
                    : "text-bark-light hover:bg-cream-dark"
                  }
                `}
              >
                <Settings className="h-4 w-4" aria-hidden="true" />
                Settings
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
      <main className="flex-1 min-w-0 flex flex-col">
        <div className="flex-1 px-4 py-6 pb-24 lg:pb-6">
          {children}
        </div>
        <div className="hidden lg:block">
          <Footer variant="subtle" onContactClick={user.user_type === "parent" ? () => setShowContact(true) : undefined} />
        </div>
      </main>

      {/* Mobile bottom tab bar */}
      <nav className="lg:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-sand px-2 py-2 z-40" aria-label="Mobile navigation">
        <div className="flex items-center justify-around">
          {visibleItems.map(item => (
            <button
              key={item.label}
              onClick={() => {
                if (item.action) item.action();
                else navigate(item.path);
              }}
              className={`flex flex-col items-center gap-1 py-2 px-2 cursor-pointer ${
                item.variant === "danger"
                  ? "text-bark-light hover:text-terracotta transition-colors"
                  : location.pathname.startsWith(item.pathMatch)
                    ? "text-forest"
                    : "text-bark-light hover:text-forest transition-colors"
              }`}
            >
              <item.icon className="h-6 w-6" aria-hidden="true" />
              <span className="text-xs font-semibold">{item.label}</span>
            </button>
          ))}
          {hasOverflow && (
            <div className="relative" ref={moreMenuRef}>
              <button
                onClick={() => setMoreMenuOpen(prev => !prev)}
                className={`flex flex-col items-center gap-1 py-2 px-2 cursor-pointer ${
                  isMoreActive || moreMenuOpen
                    ? "text-forest"
                    : "text-bark-light hover:text-forest transition-colors"
                }`}
              >
                <MoreHorizontal className="h-6 w-6" aria-hidden="true" />
                <span className="text-xs font-semibold">More</span>
              </button>
              {moreMenuOpen && (
                <div className="absolute bottom-full right-0 mb-2 bg-white border border-sand rounded-xl shadow-lg py-2 min-w-[160px] animate-[fade-in-up_0.15s_ease-out_both]">
                  {overflowItems.map(item => (
                    <button
                      key={item.label}
                      onClick={() => {
                        if (item.action) item.action();
                        else navigate(item.path);
                        setMoreMenuOpen(false);
                      }}
                      className={`flex items-center gap-3 w-full px-4 py-2.5 text-sm font-semibold text-left cursor-pointer transition-colors ${
                        item.variant === "danger"
                          ? "text-bark-light hover:text-terracotta hover:bg-cream-dark"
                          : item.pathMatch && location.pathname.startsWith(item.pathMatch)
                            ? "text-forest bg-cream-dark"
                            : "text-bark-light hover:bg-cream-dark"
                      }`}
                    >
                      <item.icon className="h-5 w-5" aria-hidden="true" />
                      {item.label}
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </nav>

      {user.user_type === "parent" && (
        <ContactFormModal open={showContact} onClose={() => setShowContact(false)} />
      )}
    </div>
  );
}
