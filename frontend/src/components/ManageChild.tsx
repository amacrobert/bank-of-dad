import { useState } from "react";
import { put, ApiRequestError } from "../api";
import { Child } from "../types";

interface ManageChildProps {
  child: Child;
  onUpdated: () => void;
  onClose: () => void;
}

export default function ManageChild({ child, onUpdated, onClose }: ManageChildProps) {
  const [newPassword, setNewPassword] = useState("");
  const [newName, setNewName] = useState(child.first_name);
  const [passwordMsg, setPasswordMsg] = useState<string | null>(null);
  const [nameMsg, setNameMsg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

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
      const result = await put<{ message: string; first_name: string }>(
        `/children/${child.id}/name`,
        { first_name: newName }
      );
      setNameMsg(`Name updated to ${result.first_name}.`);
      onUpdated();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.body.message || err.body.error);
      } else {
        setError("Failed to update name.");
      }
    }
  };

  return (
    <div className="manage-child">
      <h3>Manage {child.first_name}</h3>
      {child.is_locked && (
        <p className="warning">This account is locked. Reset the password to unlock it.</p>
      )}

      <form onSubmit={handleResetPassword}>
        <h4>Reset Password</h4>
        <div className="form-field">
          <label htmlFor="new-password">New Password (min 6 characters)</label>
          <input
            id="new-password"
            type="text"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            minLength={6}
            required
          />
        </div>
        <button type="submit">Reset Password</button>
        {passwordMsg && <p className="success">{passwordMsg}</p>}
      </form>

      <form onSubmit={handleUpdateName}>
        <h4>Update Name</h4>
        <div className="form-field">
          <label htmlFor="new-name">First Name</label>
          <input
            id="new-name"
            type="text"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            required
          />
        </div>
        <button type="submit">Update Name</button>
        {nameMsg && <p className="success">{nameMsg}</p>}
      </form>

      {error && <p className="error">{error}</p>}

      <button onClick={onClose} className="btn-secondary">
        Close
      </button>
    </div>
  );
}
