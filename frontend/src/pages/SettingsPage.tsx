import { useEffect, useState } from "react";
import { getSettings, updateTimezone, ApiRequestError } from "../api";
import { SettingsResponse } from "../types";
import Card from "../components/ui/Card";
import Button from "../components/ui/Button";
import LoadingSpinner from "../components/ui/LoadingSpinner";
import TimezoneSelect from "../components/TimezoneSelect";
import { Settings, Globe } from "lucide-react";

interface SettingsCategory {
  key: string;
  label: string;
  icon: typeof Settings;
}

const CATEGORIES: SettingsCategory[] = [
  { key: "general", label: "General", icon: Globe },
];

export default function SettingsPage() {
  const [loading, setLoading] = useState(true);
  const [activeCategory, setActiveCategory] = useState("general");

  // Settings state
  const [settings, setSettings] = useState<SettingsResponse | null>(null);
  const [selectedTimezone, setSelectedTimezone] = useState("");
  const [saving, setSaving] = useState(false);
  const [successMsg, setSuccessMsg] = useState("");
  const [errorMsg, setErrorMsg] = useState("");

  useEffect(() => {
    getSettings()
      .then((data) => {
        setSettings(data);
        setSelectedTimezone(data.timezone);
        setLoading(false);
      })
      .catch(() => {
        setLoading(false);
        setErrorMsg("Failed to load settings.");
      });
  }, []);

  const hasChanges = settings !== null && selectedTimezone !== settings.timezone;

  const handleSave = async () => {
    setSaving(true);
    setSuccessMsg("");
    setErrorMsg("");
    try {
      const result = await updateTimezone(selectedTimezone);
      setSettings({ timezone: result.timezone });
      setSuccessMsg("Timezone updated successfully.");
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setErrorMsg(err.body.message || err.body.error || "Failed to save.");
      } else {
        setErrorMsg("Failed to save timezone.");
      }
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="max-w-[960px] mx-auto flex items-center justify-center min-h-[300px]">
        <LoadingSpinner message="Loading settings..." />
      </div>
    );
  }

  return (
    <div className="max-w-[960px] mx-auto animate-fade-in-up">
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
          {activeCategory === "general" && (
            <Card>
              <h2 className="text-lg font-bold text-bark mb-4">General</h2>

              <div className="space-y-6">
                {/* Timezone setting */}
                <div>
                  <TimezoneSelect
                    value={selectedTimezone}
                    onChange={(tz) => {
                      setSelectedTimezone(tz);
                      setSuccessMsg("");
                      setErrorMsg("");
                    }}
                  />
                </div>

                {/* Save button + messages */}
                <div className="flex items-center gap-4">
                  <Button
                    onClick={handleSave}
                    disabled={!hasChanges || saving}
                    loading={saving}
                  >
                    Save changes
                  </Button>

                  {successMsg && (
                    <p className="text-sm font-medium text-forest">{successMsg}</p>
                  )}
                  {errorMsg && (
                    <p className="text-sm font-medium text-terracotta">{errorMsg}</p>
                  )}
                </div>
              </div>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
