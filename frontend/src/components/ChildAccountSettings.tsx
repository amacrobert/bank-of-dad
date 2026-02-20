import { useState } from "react";
import { put, deleteChild, ApiRequestError } from "../api";
import { Child } from "../types";
import Card from "./ui/Card";
import Input from "./ui/Input";
import Button from "./ui/Button";
import AvatarPicker from "./AvatarPicker";
import { Trash2 } from "lucide-react";

interface ChildAccountSettingsProps {
  child: Child;
  onUpdated: () => void;
  onDeleted: () => void;
}

export default function ChildAccountSettings({ child, onUpdated, onDeleted }: ChildAccountSettingsProps) {
  const [newPassword, setNewPassword] = useState("");
  const [newName, setNewName] = useState(child.first_name);
  const [newAvatar, setNewAvatar] = useState<string | null>(child.avatar ?? null);
  const [passwordMsg, setPasswordMsg] = useState<string | null>(null);
  const [nameMsg, setNameMsg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [deleteConfirmName, setDeleteConfirmName] = useState("");
  const [deleting, setDeleting] = useState(false);

  const nameMatches = deleteConfirmName.toLowerCase() === child.first_name.toLowerCase();

  const handleResetPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setPasswordMsg(null);

    try {
      const result = await put<{ message: string; account_unlocked: boolean }>(
        `/children/${child.id}/password`,
        { password: newPassword }
      );
      setPasswordMsg(
        result.account_unlocked
          ? "Password updated and account unlocked."
          : "Password updated."
      );
      setNewPassword("");
      onUpdated();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to reset password.");
      }
    }
  };

  const handleUpdateName = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setNameMsg(null);

    try {
      const result = await put<{ message: string; first_name: string; avatar: string | null }>(
        `/children/${child.id}/name`,
        { first_name: newName, avatar: newAvatar }
      );
      setNameMsg("Updated successfully.");
      setNewAvatar(result.avatar ?? null);
      onUpdated();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to update name.");
      }
    }
  };

  const handleDeleteChild = async () => {
    if (!nameMatches) return;
    setDeleting(true);
    setError(null);
    try {
      await deleteChild(child.id);
      onDeleted();
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
    <div className="space-y-4">
      <h3 className="text-lg font-bold text-bark">Account Settings for {child.first_name}</h3>

      {/* Reset password */}
      <Card padding="md">
        <h4 className="text-base font-bold text-bark mb-4">Reset Password</h4>
        <form onSubmit={handleResetPassword} className="space-y-4">
          <Input
            label="New Password (min 6 characters)"
            id="new-password"
            type="text"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            minLength={6}
            required
          />
          <Button type="submit" className="w-full">Reset Password</Button>
          {passwordMsg && (
            <div className="bg-forest/5 border border-forest/15 rounded-xl p-3">
              <p className="text-sm text-forest font-medium">{passwordMsg}</p>
            </div>
          )}
        </form>
      </Card>

      {/* Update name and avatar */}
      <Card padding="md">
        <h4 className="text-base font-bold text-bark mb-4">Update Name and Avatar</h4>
        <form onSubmit={handleUpdateName} className="space-y-4">
          <Input
            label="First Name"
            id="new-name"
            type="text"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            required
          />
          <AvatarPicker selected={newAvatar} onSelect={setNewAvatar} />
          <Button type="submit" className="w-full">Update</Button>
          {nameMsg && (
            <div className="bg-forest/5 border border-forest/15 rounded-xl p-3">
              <p className="text-sm text-forest font-medium">{nameMsg}</p>
            </div>
          )}
        </form>
      </Card>

      {/* Delete account */}
      <Card padding="md">
        <div className="flex items-center gap-2 mb-3">
          <Trash2 className="h-5 w-5 text-terracotta" aria-hidden="true" />
          <h4 className="text-base font-bold text-terracotta">Delete Account</h4>
        </div>
        <p className="text-sm text-bark-light mb-4">
          This action is permanent and cannot be undone. All of {child.first_name}&apos;s transactions, schedules, and account data will be permanently deleted.
        </p>
        <Input
          label={`Type "${child.first_name}" to confirm`}
          id="delete-confirm"
          type="text"
          value={deleteConfirmName}
          onChange={(e) => setDeleteConfirmName(e.target.value)}
        />
        <Button
          variant="secondary"
          onClick={handleDeleteChild}
          disabled={!nameMatches || deleting}
          className="w-full mt-4 !bg-terracotta/10 !text-terracotta !border-terracotta/20 hover:!bg-terracotta/20 disabled:opacity-50"
        >
          {deleting ? "Deleting..." : "Permanently Delete Account"}
        </Button>
      </Card>

      {error && (
        <div className="bg-terracotta/10 border border-terracotta/20 rounded-xl p-3">
          <p className="text-sm text-terracotta font-medium">{error}</p>
        </div>
      )}
    </div>
  );
}
