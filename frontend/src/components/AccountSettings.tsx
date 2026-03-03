import { useState, useEffect } from "react";
import { deleteAccount, getSubscription, ApiRequestError } from "../api";
import { clearTokens } from "../auth";
import Card from "./ui/Card";
import Input from "./ui/Input";
import Button from "./ui/Button";
import { Trash2, AlertTriangle } from "lucide-react";

export default function AccountSettings() {
  const [confirmText, setConfirmText] = useState("");
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasActiveSubscription, setHasActiveSubscription] = useState(false);

  useEffect(() => {
    getSubscription()
      .then((data) => {
        if (
          data.subscription_status === "active" &&
          !data.cancel_at_period_end
        ) {
          setHasActiveSubscription(true);
        }
      })
      .catch(() => {
        // Fail open — backend guard is the real enforcement
      });
  }, []);

  const confirmed = confirmText === "DELETE";

  const handleDelete = async () => {
    if (!confirmed) return;
    setDeleting(true);
    setError(null);
    try {
      await deleteAccount();
      clearTokens();
      window.location.href = "/";
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to delete account.");
      }
      setDeleting(false);
    }
  };

  return (
    <Card>
      <div className="flex items-center gap-2 mb-3">
        <Trash2 className="h-5 w-5 text-terracotta" aria-hidden="true" />
        <h2 className="text-lg font-bold text-terracotta">Delete Account</h2>
      </div>
      {hasActiveSubscription ? (
        <div className="bg-terracotta/10 border border-terracotta/20 rounded-xl p-4">
          <div className="flex items-start gap-3">
            <AlertTriangle className="h-5 w-5 text-terracotta shrink-0 mt-0.5" aria-hidden="true" />
            <div>
              <p className="text-sm font-medium text-terracotta mb-1">
                Active subscription
              </p>
              <p className="text-sm text-bark-light">
                You must cancel your subscription before deleting your account.
                Go to the <strong>Subscription</strong> tab to manage your plan.
              </p>
            </div>
          </div>
        </div>
      ) : (
        <>
          <p className="text-sm text-bark-light mb-4">
            This action is permanent and cannot be undone. Your family, all children, transactions, schedules, and account data will be permanently deleted.
          </p>
          <Input
            label='Type "DELETE" to confirm'
            id="delete-account-confirm"
            type="text"
            value={confirmText}
            onChange={(e) => setConfirmText(e.target.value)}
          />
          <Button
            variant="danger"
            onClick={handleDelete}
            disabled={!confirmed || deleting}
            loading={deleting}
            className="w-full mt-4"
          >
            {deleting ? "Deleting..." : "Permanently Delete Account"}
          </Button>
          {error && (
            <div className="bg-terracotta/10 border border-terracotta/20 rounded-xl p-3 mt-4">
              <p className="text-sm text-terracotta font-medium">{error}</p>
            </div>
          )}
        </>
      )}
    </Card>
  );
}
