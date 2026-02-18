import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { get, updateChildTheme, ApiRequestError } from "../api";
import { ChildUser } from "../types";
import { THEME_SLUGS, getTheme } from "../themes";
import { useTheme } from "../context/ThemeContext";
import Layout from "../components/Layout";
import Card from "../components/ui/Card";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import { Settings, Palette, Check } from "lucide-react";

interface SettingsCategory {
  key: string;
  label: string;
  icon: typeof Settings;
}

const CATEGORIES: SettingsCategory[] = [
  { key: "appearance", label: "Appearance", icon: Palette },
];

export default function ChildSettingsPage() {
  const navigate = useNavigate();
  const { theme: currentTheme, setTheme } = useTheme();
  const [user, setUser] = useState<ChildUser | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeCategory, setActiveCategory] = useState("appearance");
  const [saving, setSaving] = useState(false);
  const [successMsg, setSuccessMsg] = useState("");
  const [errorMsg, setErrorMsg] = useState("");

  useEffect(() => {
    get<ChildUser>("/auth/me")
      .then((data) => {
        if (data.user_type !== "child") {
          navigate("/");
          return;
        }
        setUser(data);
        setLoading(false);
      })
      .catch(() => {
        navigate("/");
      });
  }, [navigate]);

  const handleSelectTheme = async (slug: string) => {
    if (slug === currentTheme || saving) return;

    setSaving(true);
    setSuccessMsg("");
    setErrorMsg("");

    try {
      await updateChildTheme(slug);
      setTheme(slug);
      setSuccessMsg("Theme updated!");
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setErrorMsg(err.body.message || err.body.error || "Failed to save theme.");
      } else {
        setErrorMsg("Failed to save theme.");
      }
    } finally {
      setSaving(false);
    }
  };

  if (loading || !user) {
    return (
      <div className="min-h-screen bg-cream flex items-center justify-center">
        <LoadingSpinner message="Loading settings..." />
      </div>
    );
  }

  return (
    <Layout user={user} maxWidth="wide">
      <div className="animate-fade-in-up">
        <div className="flex items-center gap-3 mb-6">
          <Settings className="h-6 w-6 text-forest" aria-hidden="true" />
          <h1 className="text-2xl font-bold text-bark">Settings</h1>
        </div>

        <div className="flex flex-col md:flex-row gap-6">
          {/* Category navigation — sidebar on desktop, tabs on mobile */}
          <nav className="md:w-48 flex-shrink-0" aria-label="Settings categories">
            {/* Mobile: horizontal tabs */}
            <div className="flex md:hidden gap-2 overflow-x-auto pb-2">
              {CATEGORIES.map((cat) => {
                const Icon = cat.icon;
                const isActive = activeCategory === cat.key;
                return (
                  <button
                    key={cat.key}
                    onClick={() => setActiveCategory(cat.key)}
                    className={`
                      flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold whitespace-nowrap
                      transition-colors cursor-pointer
                      ${isActive
                        ? "bg-forest text-white"
                        : "bg-white text-bark-light border border-sand hover:bg-cream-dark"
                      }
                    `}
                  >
                    <Icon className="h-4 w-4" aria-hidden="true" />
                    {cat.label}
                  </button>
                );
              })}
            </div>

            {/* Desktop: vertical sidebar */}
            <div className="hidden md:flex flex-col gap-1">
              {CATEGORIES.map((cat) => {
                const Icon = cat.icon;
                const isActive = activeCategory === cat.key;
                return (
                  <button
                    key={cat.key}
                    onClick={() => setActiveCategory(cat.key)}
                    className={`
                      flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-semibold
                      transition-colors text-left cursor-pointer
                      ${isActive
                        ? "bg-forest text-white"
                        : "text-bark-light hover:bg-cream-dark"
                      }
                    `}
                  >
                    <Icon className="h-4 w-4" aria-hidden="true" />
                    {cat.label}
                  </button>
                );
              })}
            </div>
          </nav>

          {/* Settings content area */}
          <div className="flex-1 min-w-0">
            {activeCategory === "appearance" && (
              <Card>
                <h2 className="text-lg font-bold text-bark mb-2">Theme</h2>
                <p className="text-sm text-bark-light mb-5">Choose a visual theme for your experience.</p>

                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  {THEME_SLUGS.map((slug) => {
                    const themeDef = getTheme(slug);
                    const isActive = currentTheme === slug;
                    return (
                      <button
                        key={slug}
                        onClick={() => handleSelectTheme(slug)}
                        disabled={saving}
                        className={`
                          relative rounded-xl border-2 p-4 transition-all cursor-pointer
                          ${isActive
                            ? "border-current shadow-md"
                            : "border-sand hover:border-current hover:shadow-sm"
                          }
                          ${saving ? "opacity-60 cursor-wait" : ""}
                        `}
                        style={{
                          borderColor: isActive ? themeDef.colors.forest : undefined,
                          color: themeDef.colors.forest,
                        }}
                      >
                        {/* Active indicator */}
                        {isActive && (
                          <div
                            className="absolute top-2 right-2 w-5 h-5 rounded-full flex items-center justify-center"
                            style={{ backgroundColor: themeDef.colors.forest }}
                          >
                            <Check className="h-3 w-3 text-white" />
                          </div>
                        )}

                        {/* Theme preview */}
                        <div
                          className="w-full h-20 rounded-lg mb-3 border border-sand/50"
                          style={{
                            backgroundColor: themeDef.colors.cream,
                            backgroundImage: themeDef.backgroundSvg,
                            backgroundRepeat: "repeat",
                          }}
                        >
                          {/* Accent color bar */}
                          <div
                            className="h-2 rounded-t-lg"
                            style={{ backgroundColor: themeDef.colors.forest }}
                          />
                        </div>

                        {/* Theme name */}
                        <span className="text-sm font-semibold" style={{ color: themeDef.colors.forest }}>
                          {themeDef.label}
                        </span>
                      </button>
                    );
                  })}
                </div>

                {/* Status messages */}
                {successMsg && (
                  <p className="mt-4 text-sm font-medium text-forest">{successMsg}</p>
                )}
                {errorMsg && (
                  <p className="mt-4 text-sm font-medium text-terracotta">{errorMsg}</p>
                )}
              </Card>
            )}
          </div>
        </div>
      </div>
    </Layout>
  );
}
