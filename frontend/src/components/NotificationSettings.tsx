import { useEffect, useState } from "react";
import { getNotificationPrefs, updateNotificationPrefs, ApiRequestError } from "../api";
import { NotificationPreferences } from "../types";
import Card from "./ui/Card";
import LoadingSpinner from "./ui/LoadingSpinner";

interface ToggleRowProps {
  label: string;
  description: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled: boolean;
}

function ToggleRow({ label, description, checked, onChange, disabled }: ToggleRowProps) {
  return (
    <div className="flex items-center justify-between py-3">
      <div className="flex-1 min-w-0 pr-4">
        <p className="text-sm font-semibold text-bark">{label}</p>
        <p className="text-xs text-bark-light mt-0.5">{description}</p>
      </div>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        disabled={disabled}
        className={`
          relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full
          border-2 border-transparent transition-colors duration-200 ease-in-out
          focus:outline-none focus:ring-2 focus:ring-forest focus:ring-offset-2
          ${checked ? "bg-forest" : "bg-sand"}
          ${disabled ? "opacity-50 cursor-not-allowed" : ""}
        `}
      >
        <span
          className={`
            pointer-events-none inline-block h-5 w-5 transform rounded-full
            bg-white shadow ring-0 transition duration-200 ease-in-out
            ${checked ? "translate-x-5" : "translate-x-0"}
          `}
        />
      </button>
    </div>
  );
}

export default function NotificationSettings() {
  const [prefs, setPrefs] = useState<NotificationPreferences | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [successMsg, setSuccessMsg] = useState("");
  const [errorMsg, setErrorMsg] = useState("");

  useEffect(() => {
    getNotificationPrefs()
      .then((data) => {
        setPrefs(data);
        setLoading(false);
      })
      .catch(() => {
        setLoading(false);
        setErrorMsg("Failed to load notification preferences.");
      });
  }, []);

  const handleToggle = async (field: keyof NotificationPreferences, value: boolean) => {
    if (!prefs) return;

    // Optimistic update
    const prev = { ...prefs };
    setPrefs({ ...prefs, [field]: value });
    setSaving(true);
    setSuccessMsg("");
    setErrorMsg("");

    try {
      const updated = await updateNotificationPrefs({ [field]: value });
      setPrefs({
        notify_withdrawal_requests: updated.notify_withdrawal_requests,
        notify_chore_completions: updated.notify_chore_completions,
        notify_decisions: updated.notify_decisions,
      });
      setSuccessMsg("Preference updated.");
      setTimeout(() => setSuccessMsg(""), 2000);
    } catch (err) {
      // Revert on failure
      setPrefs(prev);
      if (err instanceof ApiRequestError) {
        setErrorMsg(err.body.message || err.body.error || "Failed to update.");
      } else {
        setErrorMsg("Failed to update preference.");
      }
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <Card>
        <div className="flex items-center justify-center min-h-[120px]">
          <LoadingSpinner message="Loading notifications..." />
        </div>
      </Card>
    );
  }

  if (!prefs) {
    return (
      <Card>
        <p className="text-sm text-terracotta">{errorMsg || "Failed to load preferences."}</p>
      </Card>
    );
  }

  return (
    <Card>
      <h2 className="text-lg font-bold text-bark mb-1">Notifications</h2>
      <p className="text-sm text-bark-light mb-4">
        Choose which email notifications you receive.
      </p>

      <div className="divide-y divide-sand">
        <ToggleRow
          label="Withdrawal requests"
          description="Get notified when a child requests a withdrawal"
          checked={prefs.notify_withdrawal_requests}
          onChange={(v) => handleToggle("notify_withdrawal_requests", v)}
          disabled={saving}
        />
        <ToggleRow
          label="Chore completions"
          description="Get notified when a child completes a chore"
          checked={prefs.notify_chore_completions}
          onChange={(v) => handleToggle("notify_chore_completions", v)}
          disabled={saving}
        />
        <ToggleRow
          label="Approval decisions"
          description="Get notified when another parent approves or denies a request"
          checked={prefs.notify_decisions}
          onChange={(v) => handleToggle("notify_decisions", v)}
          disabled={saving}
        />
      </div>

      {successMsg && (
        <p className="text-sm font-medium text-forest mt-4">{successMsg}</p>
      )}
      {errorMsg && (
        <p className="text-sm font-medium text-terracotta mt-4">{errorMsg}</p>
      )}
    </Card>
  );
}
