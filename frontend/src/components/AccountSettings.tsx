import { useState } from "react";
import { deleteAccount, ApiRequestError } from "../api";
import { clearTokens } from "../auth";
import Card from "./ui/Card";
import Input from "./ui/Input";
import Button from "./ui/Button";
import { Trash2 } from "lucide-react";

export default function AccountSettings() {
  const [confirmText, setConfirmText] = useState("");
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
    </Card>
  );
}
